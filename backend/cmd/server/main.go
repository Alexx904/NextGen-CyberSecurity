package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"nextgen-cybersec/internal/config"
	"nextgen-cybersec/internal/delivery/http/handler"
	appmw "nextgen-cybersec/internal/delivery/http/middleware"
	"nextgen-cybersec/internal/repository/postgres"
	"nextgen-cybersec/internal/scanner"
	"nextgen-cybersec/internal/service"
	"nextgen-cybersec/pkg/auth"
)

func main() {
	// 1. Inizializzazione Structured Logger JSON
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// 2. Caricamento Configurazione
	cfg := config.Load()
	slog.Info("Inizializzazione servizio",
		"app", cfg.AppName,
		"version", cfg.AppVersion,
		"env", cfg.Environment,
		"port", cfg.Port,
	)

	// 3. Connessione al Database PostgreSQL (Connection Pool)
	initCtx, initCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer initCancel()

	db, err := postgres.New(initCtx, cfg)
	if err != nil {
		slog.Error("Inizializzazione connection pool PostgreSQL fallita", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// 4. Esecuzione Migrazioni DDL Pending (Zero-Dependency con embed.FS)
	if err := postgres.RunMigrations(cfg.DB.DSN()); err != nil {
		slog.Error("Applicazione migrazioni database fallita", "error", err)
		os.Exit(1)
	}

	// 5. Inizializzazione Layer di Sicurezza & Autenticazione
	hasher := auth.NewBcryptHasher(cfg.Auth.BcryptCost)
	tokenManager, err := auth.NewJWTManager(cfg.Auth.JWTSecret, cfg.Auth.AccessTokenExpiration, cfg.Auth.RefreshTokenExpiration)
	if err != nil {
		slog.Error("Inizializzazione JWTManager fallita", "error", err)
		os.Exit(1)
	}

	// 6. Seeding Idempotente Tenant e SuperAdmin di Default
	seedCfg := postgres.DefaultSeedConfig()
	if envEmail := os.Getenv("ADMIN_EMAIL"); envEmail != "" {
		seedCfg.AdminEmail = envEmail
	}
	if envPassword := os.Getenv("ADMIN_PASSWORD"); envPassword != "" {
		seedCfg.AdminPassword = envPassword
	}
	if err := postgres.SeedDefaultAdminAndTenant(initCtx, db.Pool, hasher, seedCfg); err != nil {
		slog.Error("Seeding iniziale tenant e superadmin fallito", "error", err)
		os.Exit(1)
	}

	// 7. Inizializzazione Repository e Service Layer
	userRepo := postgres.NewUserRepo(db.Pool)
	authService := service.NewAuthService(userRepo, hasher, tokenManager)

	targetRepo := postgres.NewTargetRepo(db.Pool)
	targetGroupRepo := postgres.NewTargetGroupRepo(db.Pool)
	targetService := service.NewTargetService(targetRepo, targetGroupRepo)

	scanRepo := postgres.NewScanRepo(db.Pool)
	scanProfileRepo := postgres.NewScanProfileRepo(db.Pool)
	findingRepo := postgres.NewFindingRepo(db.Pool)
	cveRepo := postgres.NewCVERepo(db.Pool)
	cveCorrelator := service.NewCVECorrelator(cveRepo)

	scannerRunner := scanner.NewNmapRunner("nmap", logger)
	scanOrchestrator := service.NewScanOrchestrator(
		service.DefaultOrchestratorConfig(),
		scanRepo,
		scanProfileRepo,
		targetRepo,
		scannerRunner,
		logger,
	)
	scanOrchestrator.SetFindingRepo(findingRepo)
	scanOrchestrator.SetCorrelator(cveCorrelator)
	scanOrchestrator.Start(context.Background())

	scheduleRepo := postgres.NewScheduleRepo(db.Pool)
	schedulerService := service.NewSchedulerService(scheduleRepo, scanRepo, scanOrchestrator, logger)
	if err := schedulerService.Start(context.Background()); err != nil {
		slog.Error("Avvio SchedulerService fallito", "error", err)
	}

	scanService := service.NewScanService(scanRepo, scanProfileRepo, targetRepo, findingRepo, scanOrchestrator)

	// 8. Configurazione Router Chi & Middleware Standard
	r := chi.NewRouter()
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)

	// 9. Registrazione Handler di Sistema, Auth, Targets & Scans
	healthHandler := handler.NewHealthHandler(cfg.AppName, cfg.AppVersion)
	authHandler := handler.NewAuthHandler(authService)
	targetHandler := handler.NewTargetHandler(targetService)
	scanHandler := handler.NewScanHandler(scanService)

	r.Get("/healthz", healthHandler.HealthCheck)
	r.Route("/api/v1", func(api chi.Router) {
		api.Get("/health", healthHandler.HealthCheck)

		api.Route("/auth", func(authRouter chi.Router) {
			authRouter.Post("/login", authHandler.Login)
			authRouter.Post("/refresh", authHandler.RefreshToken)

			authRouter.Group(func(protected chi.Router) {
				protected.Use(appmw.Authenticate(tokenManager))
				protected.Use(appmw.RequireTenant())
				protected.Get("/me", authHandler.Me)
			})
		})

		api.Route("/targets", func(targetRouter chi.Router) {
			targetRouter.Use(appmw.Authenticate(tokenManager))
			targetRouter.Use(appmw.RequireTenant())

			// Lettura target per tutti i ruoli autenticati del tenant
			targetRouter.Get("/", targetHandler.List)
			targetRouter.Get("/{id}", targetHandler.GetByID)

			// Creazione ed eliminazione riservate ad operatori SecOps o amministratori
			targetRouter.Group(func(secOpsRouter chi.Router) {
				secOpsRouter.Use(appmw.RequireSecOps())
				secOpsRouter.Post("/", targetHandler.Create)
				secOpsRouter.Delete("/{id}", targetHandler.Delete)
			})
		})

		api.Route("/scan-profiles", func(profileRouter chi.Router) {
			profileRouter.Use(appmw.Authenticate(tokenManager))
			profileRouter.Get("/", scanHandler.ListProfiles)
		})

		api.Route("/scans", func(scanRouter chi.Router) {
			scanRouter.Use(appmw.Authenticate(tokenManager))
			scanRouter.Use(appmw.RequireTenant())

			// Consultazione storico, dettagli scansioni e findings
			scanRouter.Get("/", scanHandler.List)
			scanRouter.Get("/{id}", scanHandler.GetByID)
			scanRouter.Get("/{id}/findings", scanHandler.GetFindings)

			// Avvio e cancellazione riservati ad operatori SecOps o amministratori
			scanRouter.Group(func(secOpsRouter chi.Router) {
				secOpsRouter.Use(appmw.RequireSecOps())
				secOpsRouter.Post("/", scanHandler.Start)
				secOpsRouter.Post("/{id}/cancel", scanHandler.Cancel)
			})
		})
	})

	// 7. Configurazione Server HTTP con Timeout di Sicurezza
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 8. Avvio Server in Goroutine
	serverErrors := make(chan error, 1)
	go func() {
		slog.Info("Server HTTP in ascolto", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
		}
	}()

	// 9. Gestione Segnali di Terminazione (Graceful Shutdown)
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		slog.Error("Errore critico durante l'avvio del server", "error", err)
		os.Exit(1)
	case sig := <-shutdown:
		slog.Info("Segnale di terminazione ricevuto. Avvio graceful shutdown...", "signal", sig.String())

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			slog.Error("Arresto forzato del server a causa di un errore", "error", err)
			_ = server.Close()
			os.Exit(1)
		}

		// Arresto coordinato del worker pool scansioni e scheduler
		schedulerService.Stop()
		scanOrchestrator.Stop()
		slog.Info("Server arrestato con successo in modo controllato")
	}
}

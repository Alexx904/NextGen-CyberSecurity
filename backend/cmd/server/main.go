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
	"nextgen-cybersec/internal/delivery/ws"
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
	tenantRepo := postgres.NewTenantRepo(db.Pool)
	tenantService := service.NewTenantService(tenantRepo)

	userRepo := postgres.NewUserRepo(db.Pool)
	authService := service.NewAuthService(userRepo, hasher, tokenManager)

	targetRepo := postgres.NewTargetRepo(db.Pool)
	targetGroupRepo := postgres.NewTargetGroupRepo(db.Pool)
	targetService := service.NewTargetService(targetRepo, targetGroupRepo)
	targetGroupService := service.NewTargetGroupService(targetGroupRepo)

	scanRepo := postgres.NewScanRepo(db.Pool)
	scanProfileRepo := postgres.NewScanProfileRepo(db.Pool)
	findingRepo := postgres.NewFindingRepo(db.Pool)
	cveRepo := postgres.NewCVERepo(db.Pool)
	cveSyncService := service.NewCVESyncService(cveRepo, 500)
	cveCount, err := cveRepo.Count(initCtx)
	if err == nil && cveCount == 0 {
		slog.Info("Catalogo CVE locale vuoto: importazione feed predefinito incorporato...")
		if stats, syncErr := cveSyncService.SyncDefaultFeed(initCtx); syncErr != nil {
			slog.Warn("Importazione automatica feed CVE fallita", "error", syncErr)
		} else {
			slog.Info("Feed CVE predefinito importato con successo", "upserted", stats.TotalUpserted, "duration", stats.Duration)
		}
	}
	cveCorrelator := service.NewCVECorrelator(cveRepo)

	scannerRunner := scanner.NewNmapRunner("nmap", logger)
	eventBroadcaster := service.NewEventBroadcaster(64, logger)
	defer eventBroadcaster.Close()

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
	scanOrchestrator.SetEventBroadcaster(eventBroadcaster)
	scanOrchestrator.Start(context.Background())

	scheduleRepo := postgres.NewScheduleRepo(db.Pool)
	schedulerService := service.NewSchedulerService(scheduleRepo, scanRepo, scanOrchestrator, logger)
	if err := schedulerService.Start(context.Background()); err != nil {
		slog.Error("Avvio SchedulerService fallito", "error", err)
	}

	scanService := service.NewScanService(scanRepo, scanProfileRepo, targetRepo, findingRepo, targetGroupRepo, scanOrchestrator)
	reportRepo := postgres.NewReportRepo(db.Pool)
	reportService := service.NewReportService(reportRepo, scanRepo, scanProfileRepo, targetRepo, findingRepo, tenantRepo, nil)

	ouiDB := scanner.NewOUIDatabase()
	discoveryService := service.NewDiscoveryService(scannerRunner, ouiDB, logger, eventBroadcaster)

	sshServerRepo := postgres.NewSSHServerRepo(db.Pool)
	sshService := service.NewSSHService(sshServerRepo, cfg.SSHMasterKey)
	terminalWSHandler := ws.NewSSHTerminalHandler(sshService)
	firewallService := service.NewFirewallService(sshService, sshServerRepo)
	attackPathService := service.NewAttackPathService(targetRepo, findingRepo, sshServerRepo)

	// 8. Configurazione Router Chi & Middleware Standard
	r := chi.NewRouter()
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)

	// 9. Registrazione Handler di Sistema, Auth, Targets & Scans
	healthHandler := handler.NewHealthHandler(cfg.AppName, cfg.AppVersion)
	authHandler := handler.NewAuthHandler(authService)
	tenantHandler := handler.NewTenantHandler(tenantService)
	targetHandler := handler.NewTargetHandler(targetService)
	targetGroupHandler := handler.NewTargetGroupHandler(targetGroupService, scanService)
	discoveryHandler := handler.NewDiscoveryHandler(discoveryService, targetService, targetGroupService, eventBroadcaster)
	sshFirewallHandler := handler.NewSSHFirewallHandler(sshService, firewallService)
	scanHandler := handler.NewScanHandler(scanService, eventBroadcaster)
	reportHandler := handler.NewReportHandler(reportService)
	scheduleHandler := handler.NewScheduleHandler(schedulerService, targetRepo, scanProfileRepo)
	cveHandler := handler.NewCVEHandler(cveRepo, cveSyncService)
	attackPathHandler := handler.NewAttackPathHandler(attackPathService)

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

		api.Route("/tenants", func(tenantRouter chi.Router) {
			tenantRouter.Use(appmw.Authenticate(tokenManager))
			tenantRouter.Use(appmw.RequireTenant())

			// Lettura impostazioni tenant consentita a tutti i ruoli autenticati
			tenantRouter.Get("/current", tenantHandler.GetCurrent)

			// Configurazione branding e ragione sociale riservata ad amministratori
			tenantRouter.Group(func(adminRouter chi.Router) {
				adminRouter.Use(appmw.RequireAdmin())
				adminRouter.Put("/current/branding", tenantHandler.UpdateBranding)
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

		api.Route("/target-groups", func(groupRouter chi.Router) {
			groupRouter.Use(appmw.Authenticate(tokenManager))
			groupRouter.Use(appmw.RequireTenant())

			// Lettura albero e dettagli per tutti i ruoli autenticati del tenant
			groupRouter.Get("/", targetGroupHandler.GetTree)
			groupRouter.Get("/{id}", targetGroupHandler.GetByID)

			// Creazione, modifica, eliminazione e avvio scansione cartella per SecOps e admin
			groupRouter.Group(func(secOpsRouter chi.Router) {
				secOpsRouter.Use(appmw.RequireSecOps())
				secOpsRouter.Post("/", targetGroupHandler.Create)
				secOpsRouter.Put("/{id}", targetGroupHandler.Update)
				secOpsRouter.Delete("/{id}", targetGroupHandler.Delete)
				secOpsRouter.Post("/{id}/scan", targetGroupHandler.StartScan)
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
			scanRouter.Get("/{id}/stream", scanHandler.Stream)
			scanRouter.Get("/{id}/findings", scanHandler.GetFindings)

			// Esportazione Report PDF White-Label e CSV Tabellare (RFC 4180)
			scanRouter.Get("/{id}/reports/pdf", reportHandler.DownloadPDF)
			scanRouter.Get("/{id}/reports/csv", reportHandler.DownloadCSV)

			// Avvio e cancellazione riservati ad operatori SecOps o amministratori
			scanRouter.Group(func(secOpsRouter chi.Router) {
				secOpsRouter.Use(appmw.RequireSecOps())
				secOpsRouter.Post("/", scanHandler.Start)
				secOpsRouter.Post("/{id}/cancel", scanHandler.Cancel)
			})
		})

		api.Route("/network-discovery", func(discRouter chi.Router) {
			discRouter.Use(appmw.Authenticate(tokenManager))
			discRouter.Use(appmw.RequireTenant())

			// Lettura ultimi risultati e live stream SSE consentiti a tutti i ruoli autenticati del tenant
			discRouter.Get("/results", discoveryHandler.GetResults)
			discRouter.Get("/stream", discoveryHandler.Stream)

			// Scansione rapida subnet e importazione massiva target riservate ad operatori SecOps o admin
			discRouter.Group(func(secOpsRouter chi.Router) {
				secOpsRouter.Use(appmw.RequireSecOps())
				secOpsRouter.Post("/scan", discoveryHandler.Scan)
				secOpsRouter.Post("/import-targets", discoveryHandler.ImportTargets)
			})
		})

		api.Route("/reports", func(reportRouter chi.Router) {
			reportRouter.Use(appmw.Authenticate(tokenManager))
			reportRouter.Use(appmw.RequireTenant())

			// Consultazione storico globale report con filtri e paginazione
			reportRouter.Get("/", reportHandler.List)

			// Download report unificato con selettore formato (type=pdf|csv) e template (executive|technical)
			reportRouter.Get("/{id}/download", reportHandler.DownloadReport)
		})

		api.Route("/schedules", func(schedRouter chi.Router) {
			schedRouter.Use(appmw.Authenticate(tokenManager))
			schedRouter.Use(appmw.RequireTenant())

			schedRouter.Get("/", scheduleHandler.List)
			schedRouter.Get("/{id}", scheduleHandler.GetByID)

			schedRouter.Group(func(secOpsRouter chi.Router) {
				secOpsRouter.Use(appmw.RequireSecOps())
				secOpsRouter.Post("/", scheduleHandler.Create)
				secOpsRouter.Put("/{id}/toggle", scheduleHandler.Toggle)
				secOpsRouter.Delete("/{id}", scheduleHandler.Delete)
			})
		})

		api.Route("/cve", func(cveRouter chi.Router) {
			cveRouter.Use(appmw.Authenticate(tokenManager))
			cveRouter.Get("/stats", cveHandler.GetStats)

			cveRouter.Group(func(adminRouter chi.Router) {
				adminRouter.Use(appmw.RequireAdmin())
				adminRouter.Post("/sync", cveHandler.Sync)
			})
		})

		api.Route("/ssh/servers", func(sshRouter chi.Router) {
			sshRouter.Use(appmw.Authenticate(tokenManager))
			sshRouter.Use(appmw.RequireTenant())

			// Lettura inventario server consentita a tutti i ruoli autenticati del tenant
			sshRouter.Get("/", sshFirewallHandler.ListServers)
			sshRouter.Get("/{id}", sshFirewallHandler.GetServer)

			// Gestione credenziali, test connessione e firewall riservate ad operatori SecOps o admin
			sshRouter.Group(func(secOpsRouter chi.Router) {
				secOpsRouter.Use(appmw.RequireSecOps())
				secOpsRouter.Post("/", sshFirewallHandler.CreateServer)
				secOpsRouter.Put("/{id}", sshFirewallHandler.UpdateServer)
				secOpsRouter.Delete("/{id}", sshFirewallHandler.DeleteServer)
				secOpsRouter.Post("/{id}/test-connection", sshFirewallHandler.TestConnection)

				// Gestione Firewall iptables con Anti-Lockout Safety Timer
				secOpsRouter.Get("/{id}/firewall", sshFirewallHandler.GetFirewallStatus)
				secOpsRouter.Post("/{id}/firewall/rules", sshFirewallHandler.AddFirewallRule)
				secOpsRouter.Delete("/{id}/firewall/rules/{line}", sshFirewallHandler.DeleteFirewallRule)
				secOpsRouter.Post("/{id}/firewall/apply", sshFirewallHandler.ApplyWithSafetyTimer)
				secOpsRouter.Post("/{id}/firewall/confirm", sshFirewallHandler.ConfirmFirewall)
				secOpsRouter.Post("/{id}/firewall/rollback", sshFirewallHandler.RollbackFirewall)
			})
		})

		api.Route("/attack-paths", func(apRouter chi.Router) {
			apRouter.Use(appmw.Authenticate(tokenManager))
			apRouter.Use(appmw.RequireTenant())

			apRouter.Get("/graph", attackPathHandler.GetGraph)
			apRouter.Get("/chokepoints", attackPathHandler.GetChokepoints)
		})
	})

	// 10. Rotte WebSocket /ws/v1
	r.Route("/ws/v1", func(wsRouter chi.Router) {
		wsRouter.Route("/ssh/servers/{id}/terminal", func(termRouter chi.Router) {
			termRouter.Use(appmw.Authenticate(tokenManager))
			termRouter.Use(appmw.RequireTenant())
			termRouter.Use(appmw.RequireSecOps())
			termRouter.Get("/", terminalWSHandler.HandleTerminal)
		})
	})

	// 11. Configurazione Server HTTP con Timeout di Sicurezza
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 300 * time.Second, // 5 minuti di tolleranza per scansioni e PDF
		IdleTimeout:  120 * time.Second,
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

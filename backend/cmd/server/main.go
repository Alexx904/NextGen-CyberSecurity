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
	"github.com/go-chi/chi/v5/middleware"

	"nextgen-cybersec/internal/config"
	"nextgen-cybersec/internal/delivery/http/handler"
	"nextgen-cybersec/internal/repository/postgres"
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

	// 5. Configurazione Router Chi & Middleware Standard
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// 6. Registrazione Handler di Sistema
	healthHandler := handler.NewHealthHandler(cfg.AppName, cfg.AppVersion)
	r.Get("/healthz", healthHandler.HealthCheck)
	r.Route("/api/v1", func(api chi.Router) {
		api.Get("/health", healthHandler.HealthCheck)
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
		slog.Info("Server arrestato con successo in modo controllato")
	}
}

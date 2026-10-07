package postgres

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"nextgen-cybersec/internal/config"
)

// DB incapsula il connection pool PostgreSQL
type DB struct {
	Pool *pgxpool.Pool
}

// New inizializza un connection pool verso PostgreSQL con retry loop esponenziale/costante
func New(ctx context.Context, cfg *config.Config) (*DB, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.DB.DSN())
	if err != nil {
		return nil, fmt.Errorf("parsing configurazione pool DSN fallito: %w", err)
	}

	poolConfig.MaxConns = cfg.DB.MaxConns
	poolConfig.MinConns = cfg.DB.MinConns
	poolConfig.MaxConnLifetime = cfg.DB.MaxConnLifetime
	poolConfig.MaxConnIdleTime = cfg.DB.MaxConnIdleTime

	const maxRetries = 5
	const retryInterval = 2 * time.Second

	var pool *pgxpool.Pool
	for attempt := 1; attempt <= maxRetries; attempt++ {
		slog.Info("Connessione a PostgreSQL in corso...",
			"attempt", attempt,
			"maxRetries", maxRetries,
			"host", cfg.DB.Host,
			"port", cfg.DB.Port,
			"database", cfg.DB.Name,
		)

		pool, err = pgxpool.NewWithConfig(ctx, poolConfig)
		if err == nil {
			pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
			pingErr := pool.Ping(pingCtx)
			cancel()

			if pingErr == nil {
				slog.Info("Connessione al database PostgreSQL stabilita con successo",
					"host", cfg.DB.Host,
					"database", cfg.DB.Name,
					"max_conns", cfg.DB.MaxConns,
					"min_conns", cfg.DB.MinConns,
				)
				return &DB{Pool: pool}, nil
			}

			// Se il ping fallisce chiudiamo il pool prima di riprovare
			pool.Close()
			err = pingErr
		}

		slog.Warn("Database non ancora pronto, nuovo tentativo in corso...",
			"attempt", attempt,
			"error", err.Error(),
		)

		if attempt == maxRetries {
			break
		}

		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("interruzione connessione database dal context: %w", ctx.Err())
		case <-time.After(retryInterval):
		}
	}

	return nil, fmt.Errorf("impossibile connettersi al database dopo %d tentativi: %w", maxRetries, err)
}

// Ping verifica la raggiungibilità del database
func (db *DB) Ping(ctx context.Context) error {
	return db.Pool.Ping(ctx)
}

// Close rilascia tutte le connessioni del pool
func (db *DB) Close() {
	if db != nil && db.Pool != nil {
		slog.Info("Chiusura del connection pool PostgreSQL...")
		db.Pool.Close()
	}
}

// Stats restituisce le metriche correnti del pool
func (db *DB) Stats() *pgxpool.Stat {
	if db != nil && db.Pool != nil {
		return db.Pool.Stat()
	}
	return nil
}


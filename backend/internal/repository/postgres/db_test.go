//go:build integration

package postgres_test

import (
	"context"
	"testing"
	"time"

	"nextgen-cybersec/internal/config"
	"nextgen-cybersec/internal/repository/postgres"
)

func TestPostgresConnectionAndPing(t *testing.T) {
	cfg := getTestConfig(t)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := postgres.New(ctx, cfg)
	if err != nil {
		t.Fatal("Integration test database is unavailable")
		return
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Ping fallito dopo connessione riuscita: %v", err)
	}

	stats := db.Stats()
	if stats == nil {
		t.Fatal("Stats() non dovrebbe restituire nil su pool attivo")
	}
	if stats.MaxConns() != cfg.DB.MaxConns {
		t.Errorf("Atteso MaxConns %d, ottenuto %d", cfg.DB.MaxConns, stats.MaxConns())
	}
}

func TestRunMigrationsIdempotence(t *testing.T) {
	cfg := getTestConfig(t)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Verifica se DB è attivo
	db, err := postgres.New(ctx, cfg)
	if err != nil {
		t.Fatal("Integration test database is unavailable")
		return
	}
	db.Close()

	// Prima esecuzione
	err = postgres.RunMigrations(cfg.DB.DSN())
	if err != nil {
		t.Fatalf("Prima esecuzione RunMigrations fallita: %v", err)
	}

	// Seconda esecuzione (idempotente: deve rilevare ErrNoChange e restituire nil)
	err = postgres.RunMigrations(cfg.DB.DSN())
	if err != nil {
		t.Fatalf("Seconda esecuzione RunMigrations fallita (idempotenza): %v", err)
	}
}

func TestInvalidConnectionRetryFailure(t *testing.T) {
	cfg := &config.Config{DB: config.DBConfig{User: "test", Name: "nextgen_test", SSLMode: "disable", MaxConns: 1}}
	cfg.DB.Host = "192.0.2.1" // Indirizzo TEST-NET non instradabile (RFC 5737)
	cfg.DB.Port = "5432"

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := postgres.New(ctx, cfg)
	if err == nil {
		t.Fatal("Atteso errore di connessione su host non raggiungibile, ma New ha restituito nil")
	}
}

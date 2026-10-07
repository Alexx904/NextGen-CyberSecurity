package config_test

import (
	"os"
	"testing"
	"time"

	"nextgen-cybersec/internal/config"
)

func TestLoadDefaults(t *testing.T) {
	// Pulisci variabili temporanee
	os.Unsetenv("PORT")
	os.Unsetenv("APP_ENV")
	os.Unsetenv("DB_HOST")
	os.Unsetenv("DB_PORT")
	os.Unsetenv("DB_USER")
	os.Unsetenv("DB_PASSWORD")
	os.Unsetenv("DB_NAME")
	os.Unsetenv("DB_SSLMODE")
	os.Unsetenv("DB_MAX_CONNS")
	os.Unsetenv("DB_MIN_CONNS")

	cfg := config.Load()

	if cfg.Port != "8080" {
		t.Errorf("Atteso PORT 8080, ottenuto %s", cfg.Port)
	}
	if cfg.Environment != "development" {
		t.Errorf("Atteso APP_ENV development, ottenuto %s", cfg.Environment)
	}
	if cfg.DB.Host != "localhost" {
		t.Errorf("Atteso DB_HOST localhost, ottenuto %s", cfg.DB.Host)
	}
	if cfg.DB.Port != "5432" {
		t.Errorf("Atteso DB_PORT 5432, ottenuto %s", cfg.DB.Port)
	}
	if cfg.DB.User != "nextgen" {
		t.Errorf("Atteso DB_USER nextgen, ottenuto %s", cfg.DB.User)
	}
	if cfg.DB.Name != "nextgen_cybersec" {
		t.Errorf("Atteso DB_NAME nextgen_cybersec, ottenuto %s", cfg.DB.Name)
	}
	if cfg.DB.MaxConns != 25 {
		t.Errorf("Atteso DB_MAX_CONNS 25, ottenuto %d", cfg.DB.MaxConns)
	}
	if cfg.DB.MinConns != 5 {
		t.Errorf("Atteso DB_MIN_CONNS 5, ottenuto %d", cfg.DB.MinConns)
	}
	if cfg.DB.MaxConnLifetime != 1*time.Hour {
		t.Errorf("Atteso MaxConnLifetime 1h, ottenuto %v", cfg.DB.MaxConnLifetime)
	}
}

func TestLoadCustomEnvAndDSN(t *testing.T) {
	os.Setenv("PORT", "9090")
	os.Setenv("APP_ENV", "production")
	os.Setenv("DB_HOST", "db.example.com")
	os.Setenv("DB_PORT", "5433")
	os.Setenv("DB_USER", "prod_user")
	os.Setenv("DB_PASSWORD", "super_secret")
	os.Setenv("DB_NAME", "prod_cybersec")
	os.Setenv("DB_SSLMODE", "require")
	os.Setenv("DB_MAX_CONNS", "50")
	os.Setenv("DB_MIN_CONNS", "10")

	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("APP_ENV")
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_PORT")
		os.Unsetenv("DB_USER")
		os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("DB_NAME")
		os.Unsetenv("DB_SSLMODE")
		os.Unsetenv("DB_MAX_CONNS")
		os.Unsetenv("DB_MIN_CONNS")
	}()

	cfg := config.Load()

	if cfg.Port != "9090" {
		t.Errorf("Atteso PORT 9090, ottenuto %s", cfg.Port)
	}
	if cfg.Environment != "production" {
		t.Errorf("Atteso APP_ENV production, ottenuto %s", cfg.Environment)
	}
	if cfg.DB.Host != "db.example.com" {
		t.Errorf("Atteso DB_HOST db.example.com, ottenuto %s", cfg.DB.Host)
	}
	if cfg.DB.MaxConns != 50 {
		t.Errorf("Atteso DB_MAX_CONNS 50, ottenuto %d", cfg.DB.MaxConns)
	}
	if cfg.DB.MinConns != 10 {
		t.Errorf("Atteso DB_MIN_CONNS 10, ottenuto %d", cfg.DB.MinConns)
	}

	expectedDSN := "postgres://prod_user:super_secret@db.example.com:5433/prod_cybersec?sslmode=require"
	if cfg.DB.DSN() != expectedDSN {
		t.Errorf("Atteso DSN %s, ottenuto %s", expectedDSN, cfg.DB.DSN())
	}
}


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
	if cfg.Auth.JWTSecret != "nextgen_jwt_default_secret_key_change_me_in_prod" {
		t.Errorf("Atteso default JWTSecret, ottenuto %s", cfg.Auth.JWTSecret)
	}
	if cfg.Auth.AccessTokenExpiration != 15*time.Minute {
		t.Errorf("Atteso AccessTokenExpiration 15m, ottenuto %v", cfg.Auth.AccessTokenExpiration)
	}
	if cfg.Auth.RefreshTokenExpiration != 168*time.Hour {
		t.Errorf("Atteso RefreshTokenExpiration 168h, ottenuto %v", cfg.Auth.RefreshTokenExpiration)
	}
	if cfg.Auth.BcryptCost != 12 {
		t.Errorf("Atteso BcryptCost 12, ottenuto %d", cfg.Auth.BcryptCost)
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
	os.Setenv("JWT_SECRET", "super_custom_jwt_secret_value")
	os.Setenv("JWT_ACCESS_EXPIRATION_MINUTES", "30")
	os.Setenv("JWT_REFRESH_EXPIRATION_HOURS", "72")
	os.Setenv("BCRYPT_COST", "14")

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
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("JWT_ACCESS_EXPIRATION_MINUTES")
		os.Unsetenv("JWT_REFRESH_EXPIRATION_HOURS")
		os.Unsetenv("BCRYPT_COST")
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
	if cfg.Auth.JWTSecret != "super_custom_jwt_secret_value" {
		t.Errorf("Atteso JWTSecret personalizzato, ottenuto %s", cfg.Auth.JWTSecret)
	}
	if cfg.Auth.AccessTokenExpiration != 30*time.Minute {
		t.Errorf("Atteso AccessTokenExpiration 30m, ottenuto %v", cfg.Auth.AccessTokenExpiration)
	}
	if cfg.Auth.RefreshTokenExpiration != 72*time.Hour {
		t.Errorf("Atteso RefreshTokenExpiration 72h, ottenuto %v", cfg.Auth.RefreshTokenExpiration)
	}
	if cfg.Auth.BcryptCost != 14 {
		t.Errorf("Atteso BcryptCost 14, ottenuto %d", cfg.Auth.BcryptCost)
	}

	expectedDSN := "postgres://prod_user:super_secret@db.example.com:5433/prod_cybersec?sslmode=require"
	if cfg.DB.DSN() != expectedDSN {
		t.Errorf("Atteso DSN %s, ottenuto %s", expectedDSN, cfg.DB.DSN())
	}
}


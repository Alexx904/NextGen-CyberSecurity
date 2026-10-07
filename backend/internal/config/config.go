package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// DBConfig racchiude i parametri di configurazione del connection pool PostgreSQL
type DBConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	Name            string
	SSLMode         string
	MaxConns        int32
	MinConns        int32
	MaxConnLifetime time.Duration
	MaxConnIdleTime time.Duration
}

// DSN restituisce la stringa di connessione URI PostgreSQL formattata
func (d DBConfig) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		d.User, d.Password, d.Host, d.Port, d.Name, d.SSLMode)
}

// Config contiene le variabili di configurazione principali del server
type Config struct {
	Port        string
	Environment string
	AppName     string
	AppVersion  string
	DB          DBConfig
}

// Load carica la configurazione dalle variabili di ambiente con fallback predefiniti
func Load() *Config {
	port := getEnvOrDefault("PORT", "8080")
	env := getEnvOrDefault("APP_ENV", "development")

	dbHost := getEnvOrDefault("DB_HOST", "localhost")
	dbPort := getEnvOrDefault("DB_PORT", "5432")
	dbUser := getEnvOrDefault("DB_USER", "nextgen")
	dbPassword := getEnvOrDefault("DB_PASSWORD", "nextgen_secret_pwd")
	dbName := getEnvOrDefault("DB_NAME", "nextgen_cybersec")
	dbSSLMode := getEnvOrDefault("DB_SSLMODE", "disable")

	maxConns := int32(getEnvAsIntOrDefault("DB_MAX_CONNS", 25))
	minConns := int32(getEnvAsIntOrDefault("DB_MIN_CONNS", 5))

	return &Config{
		Port:        port,
		Environment: env,
		AppName:     "NextGen-CyberSecurity",
		AppVersion:  "1.0.0-mvp",
		DB: DBConfig{
			Host:            dbHost,
			Port:            dbPort,
			User:            dbUser,
			Password:        dbPassword,
			Name:            dbName,
			SSLMode:         dbSSLMode,
			MaxConns:        maxConns,
			MinConns:        minConns,
			MaxConnLifetime: 1 * time.Hour,
			MaxConnIdleTime: 30 * time.Minute,
		},
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}

func getEnvAsIntOrDefault(key string, defaultValue int) int {
	if val := os.Getenv(key); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			return parsed
		}
	}
	return defaultValue
}

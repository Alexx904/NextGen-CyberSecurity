//go:build integration

package postgres_test

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"testing"
	"time"

	"nextgen-cybersec/internal/config"
)

// No application configuration or .env fallback is allowed in database tests.
func loadTestDBConfig(lookup func(string) string) (*config.Config, error) {
	for _, key := range []string{"TEST_DB_HOST", "TEST_DB_PORT", "TEST_DB_NAME", "TEST_DB_USER", "TEST_DB_PASSWORD"} {
		if lookup(key) == "" {
			return nil, fmt.Errorf("%s is required for integration tests", key)
		}
	}
	if host := lookup("TEST_DB_HOST"); host != "localhost" && host != "127.0.0.1" {
		return nil, fmt.Errorf("TEST_DB_HOST must identify a local test service")
	}
	if !regexp.MustCompile(`^[a-z][a-z0-9_]*_test(?:_[a-z0-9_]+)?$`).MatchString(lookup("TEST_DB_NAME")) {
		return nil, fmt.Errorf("TEST_DB_NAME must end in _test or contain _test_ followed by an identifier")
	}
	port, err := strconv.Atoi(lookup("TEST_DB_PORT"))
	if err != nil || port < 1 || port > 65535 {
		return nil, fmt.Errorf("TEST_DB_PORT must be a valid TCP port")
	}
	return &config.Config{DB: config.DBConfig{
		Host: lookup("TEST_DB_HOST"), Port: lookup("TEST_DB_PORT"),
		Name: lookup("TEST_DB_NAME"), User: lookup("TEST_DB_USER"), Password: lookup("TEST_DB_PASSWORD"),
		SSLMode: "disable", MaxConns: 5, MinConns: 0,
		MaxConnLifetime: time.Hour, MaxConnIdleTime: time.Minute,
	}}, nil
}

func getTestConfig(t *testing.T) *config.Config {
	t.Helper()
	cfg, err := loadTestDBConfig(os.Getenv)
	if err != nil {
		t.Fatal(err)
	}
	return cfg
}

func TestIntegrationConfigSafety(t *testing.T) {
	valid := map[string]string{
		"TEST_DB_HOST": "127.0.0.1", "TEST_DB_PORT": "5432", "TEST_DB_NAME": "nextgen_test",
		"TEST_DB_USER": "test", "TEST_DB_PASSWORD": "test-only",
	}
	for _, tc := range []struct {
		name, key, value string
		wantError        bool
	}{
		{"dedicated database", "TEST_DB_NAME", "nextgen_test", false},
		{"named test run", "TEST_DB_NAME", "nextgen_company_ux02_test_20261017", false},
		{"application database", "TEST_DB_NAME", "nextgen_cybersec", true},
		{"empty database", "TEST_DB_NAME", "", true},
		{"remote host", "TEST_DB_HOST", "production.example.com", true},
		{"missing host", "TEST_DB_HOST", "", true},
		{"missing user", "TEST_DB_USER", "", true},
		{"missing port", "TEST_DB_PORT", "", true},
		{"missing password", "TEST_DB_PASSWORD", "", true},
		{"invalid port", "TEST_DB_PORT", "0", true},
		{"port outside range", "TEST_DB_PORT", "65536", true},
		{"nonnumeric port", "TEST_DB_PORT", "invalid", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			lookup := func(key string) string {
				if key == tc.key {
					return tc.value
				}
				return valid[key]
			}
			_, err := loadTestDBConfig(lookup)
			if (err != nil) != tc.wantError {
				t.Fatalf("configuration error = %v; wantError = %v", err, tc.wantError)
			}
		})
	}
}

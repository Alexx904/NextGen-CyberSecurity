package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHealthCheckHandler(t *testing.T) {
	appName := "NextGen-CyberSec-Test"
	appVersion := "1.0.0-test"
	h := NewHealthHandler(appName, appVersion)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()

	h.HealthCheck(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status code %d, got %d", http.StatusOK, rr.Code)
	}

	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", contentType)
	}

	var resp HealthResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response JSON: %v", err)
	}

	if resp.Status != "healthy" {
		t.Errorf("expected status 'healthy', got '%s'", resp.Status)
	}

	if resp.App != appName {
		t.Errorf("expected app '%s', got '%s'", appName, resp.App)
	}

	if resp.Version != appVersion {
		t.Errorf("expected version '%s', got '%s'", appVersion, resp.Version)
	}

	if time.Since(resp.Timestamp) > 10*time.Second {
		t.Errorf("timestamp is too old: %v", resp.Timestamp)
	}
}


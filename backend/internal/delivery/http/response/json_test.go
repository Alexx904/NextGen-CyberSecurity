package response_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"nextgen-cybersec/internal/delivery/http/response"
)

func TestJSONResponse(t *testing.T) {
	rec := httptest.NewRecorder()
	payload := map[string]string{"foo": "bar"}

	response.JSON(rec, http.StatusOK, payload)

	if rec.Code != http.StatusOK {
		t.Errorf("Atteso status 200, ottenuto %d", rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Atteso Content-Type application/json, ottenuto %s", contentType)
	}

	var env response.Envelope
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("Unmarshal fallito: %v", err)
	}

	if !env.Success {
		t.Errorf("Atteso success = true, ottenuto %v", env.Success)
	}
	if env.Error != nil {
		t.Errorf("Atteso error = nil, ottenuto %+v", env.Error)
	}
}

func TestErrorResponse(t *testing.T) {
	rec := httptest.NewRecorder()

	response.Error(rec, http.StatusUnauthorized, "UNAUTHORIZED", "Credenziali non valide")

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Atteso status 401, ottenuto %d", rec.Code)
	}

	var env response.Envelope
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("Unmarshal fallito: %v", err)
	}

	if env.Success {
		t.Errorf("Atteso success = false, ottenuto %v", env.Success)
	}
	if env.Data != nil {
		t.Errorf("Atteso data = nil, ottenuto %+v", env.Data)
	}
	if env.Error == nil || env.Error.Code != "UNAUTHORIZED" || env.Error.Message != "Credenziali non valide" {
		t.Errorf("Errore non strutturato come atteso: %+v", env.Error)
	}
}


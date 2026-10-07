package handler

import (
	"encoding/json"
	"net/http"
	"time"
)

// HealthResponse modella la risposta JSON per gli endpoint di health check
type HealthResponse struct {
	Status    string    `json:"status"`
	App       string    `json:"app"`
	Version   string    `json:"version"`
	Timestamp time.Time `json:"timestamp"`
}

// HealthHandler gestisce le richieste di integrità del servizio
type HealthHandler struct {
	appName    string
	appVersion string
}

// NewHealthHandler crea una nuova istanza di HealthHandler
func NewHealthHandler(appName, appVersion string) *HealthHandler {
	return &HealthHandler{
		appName:    appName,
		appVersion: appVersion,
	}
}

// HealthCheck gestisce la richiesta HTTP restituendo lo stato del servizio in JSON
func (h *HealthHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	resp := HealthResponse{
		Status:    "healthy",
		App:       h.appName,
		Version:   h.appVersion,
		Timestamp: time.Now().UTC(),
	}

	_ = json.NewEncoder(w).Encode(resp)
}


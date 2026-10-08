package handler

import (
	"net/http"

	"nextgen-cybersec/internal/delivery/http/response"
	"nextgen-cybersec/internal/domain"
	"nextgen-cybersec/internal/service"
)

// CVEHandler gestisce gli endpoint REST per l'ispezione delle statistiche e la sincronizzazione del catalogo CVE
type CVEHandler struct {
	cveRepo     domain.CVERepository
	syncService *service.CVESyncService
}

// NewCVEHandler crea una nuova istanza di CVEHandler
func NewCVEHandler(cveRepo domain.CVERepository, syncService *service.CVESyncService) *CVEHandler {
	return &CVEHandler{
		cveRepo:     cveRepo,
		syncService: syncService,
	}
}

// GetStats gestisce GET /api/v1/cve/stats restituendo conteggi e metadati del feed locale
func (h *CVEHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.cveRepo.GetStats(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Errore durante il recupero delle statistiche del catalogo CVE")
		return
	}

	response.JSON(w, http.StatusOK, stats)
}

// Sync gestisce POST /api/v1/cve/sync per avviare la sincronizzazione del catalogo CVE locale
func (h *CVEHandler) Sync(w http.ResponseWriter, r *http.Request) {
	if h.syncService == nil {
		response.Error(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Servizio di sincronizzazione CVE non configurato")
		return
	}

	stats, err := h.syncService.SyncDefaultFeed(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "SYNC_FAILED", "Sincronizzazione del catalogo CVE fallita")
		return
	}

	response.JSON(w, http.StatusOK, stats)
}


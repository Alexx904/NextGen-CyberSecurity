package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"nextgen-cybersec/internal/delivery/http/middleware"
	"nextgen-cybersec/internal/delivery/http/response"
	"nextgen-cybersec/internal/domain"
	"nextgen-cybersec/internal/service"
)

// StartScanRequest definisce il payload per l'avvio on-demand di una scansione
type StartScanRequest struct {
	TargetID  uuid.UUID `json:"target_id"`
	ProfileID uuid.UUID `json:"profile_id"`
}

// ScanResponse rappresenta il modello pubblico di una scansione
type ScanResponse struct {
	ID           string                  `json:"id"`
	TenantID     string                  `json:"tenant_id"`
	TargetID     string                  `json:"target_id"`
	ProfileID    string                  `json:"profile_id"`
	TriggerType  string                  `json:"trigger_type"`
	Status       string                  `json:"status"`
	StartedAt    *time.Time              `json:"started_at,omitempty"`
	CompletedAt  *time.Time              `json:"completed_at,omitempty"`
	SummaryStats domain.ScanSummaryStats `json:"summary_stats"`
	ErrorMessage *string                 `json:"error_message,omitempty"`
	CreatedAt    time.Time               `json:"created_at"`
	UpdatedAt    time.Time               `json:"updated_at"`
}

func toScanResponse(s *domain.Scan) ScanResponse {
	return ScanResponse{
		ID:           s.ID.String(),
		TenantID:     s.TenantID.String(),
		TargetID:     s.TargetID.String(),
		ProfileID:    s.ProfileID.String(),
		TriggerType:  string(s.TriggerType),
		Status:       string(s.Status),
		StartedAt:    s.StartedAt,
		CompletedAt:  s.CompletedAt,
		SummaryStats: s.SummaryStats,
		ErrorMessage: s.ErrorMessage,
		CreatedAt:    s.CreatedAt,
		UpdatedAt:    s.UpdatedAt,
	}
}

// ScanHandler gestisce le richieste HTTP per il ciclo di vita delle scansioni
type ScanHandler struct {
	scanService service.ScanService
}

// NewScanHandler crea una nuova istanza di ScanHandler
func NewScanHandler(scanService service.ScanService) *ScanHandler {
	return &ScanHandler{
		scanService: scanService,
	}
}

// Cancel gestisce POST /api/v1/scans/{id}/cancel per richiedere l'interruzione immediata
func (h *ScanHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Identificativo tenant mancante o non valido")
		return
	}

	idParam := chi.URLParam(r, "id")
	scanID, err := uuid.Parse(idParam)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "ID scansione non valido")
		return
	}

	if err := h.scanService.CancelScan(r.Context(), tenantID, scanID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "Scansione non trovata")
			return
		}
		if errors.Is(err, domain.ErrScanCannotBeCancelled) {
			response.Error(w, http.StatusConflict, "CONFLICT", "La scansione non si trova in uno stato annullabile")
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Impossibile annullare la scansione")
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"message": "Richiesta di annullamento scansione inoltrata con successo",
		"scan_id": scanID.String(),
	})
}

// Start gestisce POST /api/v1/scans per l'avvio di una nuova scansione
func (h *ScanHandler) Start(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Identificativo tenant mancante o non valido")
		return
	}

	var req StartScanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "Payload JSON non valido")
		return
	}

	if req.TargetID == uuid.Nil || req.ProfileID == uuid.Nil {
		response.Error(w, http.StatusBadRequest, "VALIDATION_FAILED", "target_id e profile_id sono obbligatori")
		return
	}

	scan, err := h.scanService.StartScan(r.Context(), tenantID, req.TargetID, req.ProfileID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "Target o profilo di scansione non trovato")
			return
		}
		if errors.Is(err, domain.ErrTargetBlacklisted) {
			response.Error(w, http.StatusForbidden, "FORBIDDEN", domain.ErrTargetBlacklisted.Error())
			return
		}
		if errors.Is(err, domain.ErrScanAlreadyRunning) {
			response.Error(w, http.StatusConflict, "CONFLICT", domain.ErrScanAlreadyRunning.Error())
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	response.JSON(w, http.StatusAccepted, toScanResponse(scan))
}

// GetByID gestisce GET /api/v1/scans/{id} per recuperare i dettagli di una scansione
func (h *ScanHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Identificativo tenant mancante o non valido")
		return
	}

	idParam := chi.URLParam(r, "id")
	scanID, err := uuid.Parse(idParam)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "ID scansione non valido")
		return
	}

	scan, err := h.scanService.GetScan(r.Context(), tenantID, scanID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "Scansione non trovata")
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Errore durante il recupero della scansione")
		return
	}

	response.JSON(w, http.StatusOK, toScanResponse(scan))
}

// List gestisce GET /api/v1/scans per recuperare la cronologia delle scansioni
func (h *ScanHandler) List(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Identificativo tenant mancante o non valido")
		return
	}

	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 && val <= 100 {
			limit = val
		}
	}

	offset := 0
	if o := r.URL.Query().Get("offset"); o != "" {
		if val, err := strconv.Atoi(o); err == nil && val >= 0 {
			offset = val
		}
	}

	scans, total, err := h.scanService.ListScans(r.Context(), tenantID, limit, offset)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Errore durante il recupero dell'elenco scansioni")
		return
	}

	items := make([]ScanResponse, 0, len(scans))
	for _, s := range scans {
		items = append(items, toScanResponse(s))
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"items":  items,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// FindingResponse modella la rappresentazione pubblica di una vulnerabilità o servizio
type FindingResponse struct {
	ID             string    `json:"id"`
	ScanID         string    `json:"scan_id"`
	TenantID       string    `json:"tenant_id"`
	HostIP         string    `json:"host_ip"`
	Hostname       *string   `json:"hostname,omitempty"`
	Port           int       `json:"port"`
	Protocol       string    `json:"protocol"`
	ServiceName    *string   `json:"service_name,omitempty"`
	ServiceVersion *string   `json:"service_version,omitempty"`
	CVEID          *string   `json:"cve_id,omitempty"`
	CVSSScore      *float64  `json:"cvss_score,omitempty"`
	Severity       string    `json:"severity"`
	RawBanner      *string   `json:"raw_banner,omitempty"`
	Remediation    *string   `json:"remediation,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

func toFindingResponse(f *domain.Finding) FindingResponse {
	return FindingResponse{
		ID:             f.ID.String(),
		ScanID:         f.ScanID.String(),
		TenantID:       f.TenantID.String(),
		HostIP:         f.HostIP,
		Hostname:       f.Hostname,
		Port:           f.Port,
		Protocol:       f.Protocol,
		ServiceName:    f.ServiceName,
		ServiceVersion: f.ServiceVersion,
		CVEID:          f.CVEID,
		CVSSScore:      f.CVSSScore,
		Severity:       string(f.Severity),
		RawBanner:      f.RawBanner,
		Remediation:    f.Remediation,
		CreatedAt:      f.CreatedAt,
	}
}

// ProfileResponse modella la rappresentazione pubblica di un profilo di scansione
type ProfileResponse struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	NmapArgs       string  `json:"nmap_args"`
	Description    *string `json:"description,omitempty"`
	TimeoutSeconds int     `json:"timeout_seconds"`
}

// GetFindings gestisce GET /api/v1/scans/{id}/findings con filtri e paginazione
func (h *ScanHandler) GetFindings(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Identificativo tenant mancante o non valido")
		return
	}

	idParam := chi.URLParam(r, "id")
	scanID, err := uuid.Parse(idParam)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "ID scansione non valido")
		return
	}

	filter := domain.FindingFilter{
		HostIP: r.URL.Query().Get("host_ip"),
		CVEID:  r.URL.Query().Get("cve_id"),
		Limit:  50,
		Offset: 0,
	}

	if sev := r.URL.Query().Get("severity"); sev != "" {
		s := domain.SeverityLevel(sev)
		filter.Severity = &s
	}

	if portStr := r.URL.Query().Get("port"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil && p > 0 {
			filter.Port = &p
		}
	}

	if l := r.URL.Query().Get("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 && val <= 100 {
			filter.Limit = val
		}
	}

	if o := r.URL.Query().Get("offset"); o != "" {
		if val, err := strconv.Atoi(o); err == nil && val >= 0 {
			filter.Offset = val
		}
	}

	findings, total, err := h.scanService.GetScanFindings(r.Context(), tenantID, scanID, filter)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "Scansione non trovata")
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Errore durante il recupero dei finding")
		return
	}

	items := make([]FindingResponse, 0, len(findings))
	for _, f := range findings {
		items = append(items, toFindingResponse(f))
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"items":  items,
		"total":  total,
		"limit":  filter.Limit,
		"offset": filter.Offset,
	})
}

// ListProfiles gestisce GET /api/v1/scan-profiles per la lista dei profili disponibili
func (h *ScanHandler) ListProfiles(w http.ResponseWriter, r *http.Request) {
	profiles, err := h.scanService.ListProfiles(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Errore durante il recupero dei profili di scansione")
		return
	}

	items := make([]ProfileResponse, 0, len(profiles))
	for _, p := range profiles {
		items = append(items, ProfileResponse{
			ID:             p.ID.String(),
			Name:           p.Name,
			NmapArgs:       p.NmapArgs,
			Description:    p.Description,
			TimeoutSeconds: p.TimeoutSeconds,
		})
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"items": items,
		"total": len(items),
	})
}

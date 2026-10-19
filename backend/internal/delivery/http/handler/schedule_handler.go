package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"nextgen-cybersec/internal/delivery/http/middleware"
	"nextgen-cybersec/internal/delivery/http/response"
	"nextgen-cybersec/internal/domain"
	"nextgen-cybersec/internal/service"
)

// CreateScheduleRequest modella il payload per la registrazione di una pianificazione
type CreateScheduleRequest struct {
	Authorization  *domain.ScanAuthorizationRequest `json:"authorization"`
	TargetID       uuid.UUID                        `json:"target_id"`
	ProfileID      uuid.UUID                        `json:"profile_id"`
	CronExpression string                           `json:"cron_expression"`
	IsActive       *bool                            `json:"is_active,omitempty"`
}

// ScheduleResponse modella la rappresentazione JSON pubblica di una pianificazione
type ScheduleResponse struct {
	AuthorizationRequired bool       `json:"authorization_required"`
	ID                    string     `json:"id"`
	TenantID              string     `json:"tenant_id"`
	TargetID              string     `json:"target_id"`
	ProfileID             string     `json:"profile_id"`
	CronExpression        string     `json:"cron_expression"`
	IsActive              bool       `json:"is_active"`
	NextRunAt             *time.Time `json:"next_run_at,omitempty"`
	LastRunAt             *time.Time `json:"last_run_at,omitempty"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}

func toScheduleResponse(s *domain.Schedule) ScheduleResponse {
	return ScheduleResponse{
		ID:             s.ID.String(),
		TenantID:       s.TenantID.String(),
		TargetID:       s.TargetID.String(),
		ProfileID:      s.ProfileID.String(),
		CronExpression: s.CronExpression,
		IsActive:       s.IsActive,
		NextRunAt:      s.NextRunAt,
		LastRunAt:      s.LastRunAt,
		CreatedAt:      s.CreatedAt,
		UpdatedAt:      s.UpdatedAt,
	}
}

func (h *ScheduleHandler) scheduleResponse(ctx context.Context, s *domain.Schedule) ScheduleResponse {
	result := toScheduleResponse(s)
	result.AuthorizationRequired = h.schedulerService.AuthorizationRequired(ctx, s)
	return result
}

// ScheduleHandler gestisce le richieste HTTP relative alle pianificazioni ricorrenti
type ScheduleHandler struct {
	schedulerService *service.SchedulerService
	targetRepo       domain.TargetRepository
	profileRepo      domain.ScanProfileRepository
}

// NewScheduleHandler crea una nuova istanza di ScheduleHandler
func NewScheduleHandler(
	schedulerService *service.SchedulerService,
	targetRepo domain.TargetRepository,
	profileRepo domain.ScanProfileRepository,
) *ScheduleHandler {
	return &ScheduleHandler{
		schedulerService: schedulerService,
		targetRepo:       targetRepo,
		profileRepo:      profileRepo,
	}
}

// List gestisce GET /api/v1/schedules per consultare le pianificazioni del tenant
func (h *ScheduleHandler) List(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Identificativo tenant mancante nel contesto")
		return
	}

	list, err := h.schedulerService.ListByTenant(r.Context(), tenantID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Impossibile recuperare le pianificazioni")
		return
	}

	items := make([]ScheduleResponse, 0, len(list))
	for _, s := range list {
		items = append(items, h.scheduleResponse(r.Context(), s))
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"items": items,
		"total": len(items),
	})
}

// GetByID gestisce GET /api/v1/schedules/{id} per recuperare il dettaglio di una pianificazione
func (h *ScheduleHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Identificativo tenant mancante nel contesto")
		return
	}

	idParam := chi.URLParam(r, "id")
	scheduleID, err := uuid.Parse(idParam)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "ID pianificazione non valido")
		return
	}

	schedule, err := h.schedulerService.GetByID(r.Context(), tenantID, scheduleID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "Pianificazione non trovata")
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Errore nel recupero della pianificazione")
		return
	}

	response.JSON(w, http.StatusOK, h.scheduleResponse(r.Context(), schedule))
}

// Create gestisce POST /api/v1/schedules per creare una nuova pianificazione periodica
func (h *ScheduleHandler) Create(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Identificativo tenant mancante nel contesto")
		return
	}

	var req CreateScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "Payload JSON non valido")
		return
	}

	if req.TargetID == uuid.Nil || req.ProfileID == uuid.Nil || req.CronExpression == "" {
		response.Error(w, http.StatusBadRequest, "VALIDATION_FAILED", "target_id, profile_id e cron_expression sono obbligatori")
		return
	}

	// 1. Verifica esistenza e idoneità del target
	target, err := h.targetRepo.GetByID(r.Context(), tenantID, req.TargetID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "Target bersaglio non trovato per l'organizzazione")
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Verifica target fallita")
		return
	}

	if target.IsBlacklisted {
		response.Error(w, http.StatusForbidden, "FORBIDDEN", domain.ErrTargetBlacklisted.Error())
		return
	}

	// 2. Verifica profilo di scansione
	_, err = h.profileRepo.GetByID(r.Context(), req.ProfileID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "Profilo di scansione specificato non trovato")
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Verifica profilo fallita")
		return
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	schedule := &domain.Schedule{
		ID:             uuid.New(),
		TenantID:       tenantID,
		TargetID:       req.TargetID,
		ProfileID:      req.ProfileID,
		CronExpression: req.CronExpression,
		IsActive:       isActive,
	}

	userID, _ := middleware.GetUserID(r.Context())
	ctx := domain.WithScanAuthorization(r.Context(), tenantID, userID, req.Authorization)
	if err := h.schedulerService.AddSchedule(ctx, schedule); err != nil {
		if errors.Is(err, domain.ErrScanAuthorizationRequired) {
			response.Error(w, http.StatusConflict, "SCAN_AUTHORIZATION_REQUIRED", err.Error())
			return
		}
		response.Error(w, http.StatusBadRequest, "INVALID_CRON", err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, h.scheduleResponse(r.Context(), schedule))
}

// Toggle gestisce PUT /api/v1/schedules/{id}/toggle per commutare lo stato attivo/disattivo
func (h *ScheduleHandler) Toggle(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Identificativo tenant mancante nel contesto")
		return
	}

	idParam := chi.URLParam(r, "id")
	scheduleID, err := uuid.Parse(idParam)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "ID pianificazione non valido")
		return
	}

	updated, err := h.schedulerService.ToggleSchedule(r.Context(), tenantID, scheduleID)
	if err != nil {
		if errors.Is(err, domain.ErrScanAuthorizationRequired) {
			response.Error(w, http.StatusConflict, "SCAN_AUTHORIZATION_REQUIRED", err.Error())
			return
		}
		if errors.Is(err, domain.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "Pianificazione non trovata")
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Impossibile commutare lo stato della pianificazione")
		return
	}

	response.JSON(w, http.StatusOK, h.scheduleResponse(r.Context(), updated))
}

// Delete gestisce DELETE /api/v1/schedules/{id} per eliminare la pianificazione
func (h *ScheduleHandler) Delete(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Identificativo tenant mancante nel contesto")
		return
	}

	idParam := chi.URLParam(r, "id")
	scheduleID, err := uuid.Parse(idParam)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "ID pianificazione non valido")
		return
	}

	if err := h.schedulerService.DeleteSchedule(r.Context(), tenantID, scheduleID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "Pianificazione non trovata")
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Impossibile eliminare la pianificazione")
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{
		"message": "Pianificazione eliminata con successo",
	})
}

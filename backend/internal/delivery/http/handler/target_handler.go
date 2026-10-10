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

// CreateTargetRequest modella il payload JSON per la creazione di un target
type CreateTargetRequest struct {
	Value         string     `json:"value"`
	GroupID       *uuid.UUID `json:"group_id,omitempty"`
	Description   *string    `json:"description,omitempty"`
	IsBlacklisted bool       `json:"is_blacklisted"`
}

// TargetResponse modella la rappresentazione JSON pubblica di un target
type TargetResponse struct {
	ID            string    `json:"id"`
	TenantID      string    `json:"tenant_id"`
	GroupID       *string   `json:"group_id,omitempty"`
	Type          string    `json:"type"`
	Value         string    `json:"value"`
	Description   *string   `json:"description,omitempty"`
	IsBlacklisted bool      `json:"is_blacklisted"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// TargetListResponse modella l'elenco paginato dei target
type TargetListResponse struct {
	Items  []TargetResponse `json:"items"`
	Total  int              `json:"total"`
	Limit  int              `json:"limit"`
	Offset int              `json:"offset"`
}

// TargetHandler gestisce gli endpoint REST per i bersagli
type TargetHandler struct {
	targetService service.TargetService
}

// NewTargetHandler crea una nuova istanza di TargetHandler
func NewTargetHandler(targetService service.TargetService) *TargetHandler {
	return &TargetHandler{
		targetService: targetService,
	}
}

// List gestisce GET /api/v1/targets con paginazione e filtri
func (h *TargetHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, ok := middleware.GetTenantID(ctx)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Tenant non identificato nel contesto")
		return
	}

	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))

	filter := domain.TargetFilter{
		Search: q.Get("search"),
		Limit:  limit,
		Offset: offset,
	}

	if typeStr := q.Get("type"); typeStr != "" {
		t := domain.TargetType(typeStr)
		if t.IsValid() {
			filter.Type = &t
		}
	}

	if blStr := q.Get("is_blacklisted"); blStr != "" {
		if blVal, err := strconv.ParseBool(blStr); err == nil {
			filter.IsBlacklisted = &blVal
		}
	}

	if groupStr := q.Get("group_id"); groupStr != "" {
		if gID, err := uuid.Parse(groupStr); err == nil {
			filter.GroupID = &gID
		}
	}

	targets, total, err := h.targetService.ListTargets(ctx, tenantID, filter)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Impossibile recuperare i target")
		return
	}

	items := make([]TargetResponse, 0, len(targets))
	for _, t := range targets {
		items = append(items, toTargetResponse(t))
	}

	response.JSON(w, http.StatusOK, TargetListResponse{
		Items:  items,
		Total:  total,
		Limit:  filter.Limit,
		Offset: filter.Offset,
	})
}

// GetByID gestisce GET /api/v1/targets/{id}
func (h *TargetHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, ok := middleware.GetTenantID(ctx)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Tenant non identificato nel contesto")
		return
	}

	idStr := chi.URLParam(r, "id")
	targetID, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_TARGET_ID", "ID target non valido")
		return
	}

	target, err := h.targetService.GetTarget(ctx, tenantID, targetID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "TARGET_NOT_FOUND", "Target non trovato")
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Errore durante il recupero del target")
		return
	}

	response.JSON(w, http.StatusOK, toTargetResponse(target))
}

// Create gestisce POST /api/v1/targets
func (h *TargetHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, ok := middleware.GetTenantID(ctx)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Tenant non identificato nel contesto")
		return
	}

	var req CreateTargetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_PAYLOAD", "Payload JSON della richiesta non valido")
		return
	}

	target, err := h.targetService.CreateTarget(ctx, service.CreateTargetInput{
		TenantID:      tenantID,
		GroupID:       req.GroupID,
		Value:         req.Value,
		Description:   req.Description,
		IsBlacklisted: req.IsBlacklisted,
	})
	if err != nil {
		if errors.Is(err, domain.ErrTargetInternalRestricted) {
			response.Error(w, http.StatusBadRequest, "TARGET_RESTRICTED", err.Error())
			return
		}
		if errors.Is(err, domain.ErrTargetScopeTooBroad) {
			response.Error(w, http.StatusBadRequest, "SCOPE_TOO_BROAD", err.Error())
			return
		}
		if errors.Is(err, domain.ErrTargetInvalid) {
			response.Error(w, http.StatusBadRequest, "INVALID_TARGET_FORMAT", "Formato bersaglio non valido (IP, CIDR, Range o FQDN atteso, metacaratteri vietati)")
			return
		}
		if errors.Is(err, domain.ErrTargetBlacklisted) {
			response.Error(w, http.StatusConflict, "TARGET_BLACKLISTED", "Il target collide con la lista nera di esclusione configurata per l'organizzazione")
			return
		}
		if errors.Is(err, domain.ErrNotFound) {
			response.Error(w, http.StatusBadRequest, "GROUP_NOT_FOUND", "Il gruppo specificato non appartiene al tenant o non esiste")
			return
		}
		if errors.Is(err, domain.ErrConflict) {
			response.Error(w, http.StatusConflict, "TARGET_CONFLICT", "Target già registrato per l'organizzazione")
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Errore durante la registrazione del target")
		return
	}

	response.JSON(w, http.StatusCreated, toTargetResponse(target))
}

// Delete gestisce DELETE /api/v1/targets/{id}
func (h *TargetHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, ok := middleware.GetTenantID(ctx)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Tenant non identificato nel contesto")
		return
	}

	idStr := chi.URLParam(r, "id")
	targetID, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_TARGET_ID", "ID target non valido")
		return
	}

	if err := h.targetService.DeleteTarget(ctx, tenantID, targetID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "TARGET_NOT_FOUND", "Target non trovato")
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Errore durante l'eliminazione del target")
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "Target eliminato con successo"})
}

func toTargetResponse(t *domain.Target) TargetResponse {
	var groupIDStr *string
	if t.GroupID != nil {
		s := t.GroupID.String()
		groupIDStr = &s
	}

	return TargetResponse{
		ID:            t.ID.String(),
		TenantID:      t.TenantID.String(),
		GroupID:       groupIDStr,
		Type:          string(t.Type),
		Value:         t.Value,
		Description:   t.Description,
		IsBlacklisted: t.IsBlacklisted,
		CreatedAt:     t.CreatedAt,
		UpdatedAt:     t.UpdatedAt,
	}
}


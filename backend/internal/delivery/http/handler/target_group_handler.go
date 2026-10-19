package handler

import (
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

// CreateTargetGroupRequest modella il payload JSON per la creazione di un gruppo/cartella
type CreateTargetGroupRequest struct {
	Name        string     `json:"name"`
	Description *string    `json:"description,omitempty"`
	ParentID    *uuid.UUID `json:"parent_id,omitempty"`
	ColorTag    *string    `json:"color_tag,omitempty"`
}

// UpdateTargetGroupRequest modella il payload JSON per la modifica di un gruppo/cartella
type UpdateTargetGroupRequest struct {
	Name        string     `json:"name"`
	Description *string    `json:"description,omitempty"`
	ParentID    *uuid.UUID `json:"parent_id,omitempty"`
	ColorTag    *string    `json:"color_tag,omitempty"`
}

// StartGroupScanRequest modella la richiesta di avvio scansione ricorsiva su cartella
type StartGroupScanRequest struct {
	Authorization *domain.ScanAuthorizationRequest `json:"authorization"`
	ProfileID     uuid.UUID                        `json:"profile_id"`
}

// TargetGroupResponse rappresenta la struttura JSON pubblica di un gruppo
type TargetGroupResponse struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	ParentID    *string   `json:"parent_id,omitempty"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	ColorTag    *string   `json:"color_tag,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func toTargetGroupResponse(g *domain.TargetGroup) TargetGroupResponse {
	var parentIDStr *string
	if g.ParentID != nil {
		s := g.ParentID.String()
		parentIDStr = &s
	}

	return TargetGroupResponse{
		ID:          g.ID.String(),
		TenantID:    g.TenantID.String(),
		ParentID:    parentIDStr,
		Name:        g.Name,
		Description: g.Description,
		ColorTag:    g.ColorTag,
		CreatedAt:   g.CreatedAt,
		UpdatedAt:   g.UpdatedAt,
	}
}

// TargetGroupHandler gestisce gli endpoint REST per l'albero dei target groups e le scansioni su cartella
type TargetGroupHandler struct {
	groupService service.TargetGroupService
	scanService  service.ScanService
}

// NewTargetGroupHandler crea una nuova istanza di TargetGroupHandler
func NewTargetGroupHandler(groupService service.TargetGroupService, scanService service.ScanService) *TargetGroupHandler {
	return &TargetGroupHandler{
		groupService: groupService,
		scanService:  scanService,
	}
}

// GetTree gestisce GET /api/v1/target-groups restituendo l'albero nidificato con conteggi e rischio aggregato
func (h *TargetGroupHandler) GetTree(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Tenant non identificato nel contesto")
		return
	}

	tree, err := h.groupService.GetTree(r.Context(), tenantID)
	if err != nil {
		if errors.Is(err, domain.ErrValidation) {
			response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Parametri di richiesta non validi")
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Errore durante il recupero dell'albero cartelle")
		return
	}

	response.JSON(w, http.StatusOK, tree)
}

// GetByID gestisce GET /api/v1/target-groups/{id} per recuperare i dettagli di una cartella
func (h *TargetGroupHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Tenant non identificato nel contesto")
		return
	}

	idStr := chi.URLParam(r, "id")
	groupID, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "Identificativo cartella non valido")
		return
	}

	group, err := h.groupService.GetGroup(r.Context(), tenantID, groupID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "Cartella non trovata")
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Errore durante il recupero della cartella")
		return
	}

	response.JSON(w, http.StatusOK, toTargetGroupResponse(group))
}

// Create gestisce POST /api/v1/target-groups per creare una nuova cartella o sottocartella
func (h *TargetGroupHandler) Create(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Tenant non identificato nel contesto")
		return
	}

	var req CreateTargetGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "Payload JSON non valido")
		return
	}

	group, err := h.groupService.CreateGroup(r.Context(), service.CreateGroupInput{
		TenantID:    tenantID,
		ParentID:    req.ParentID,
		Name:        req.Name,
		Description: req.Description,
		ColorTag:    req.ColorTag,
	})
	if err != nil {
		if errors.Is(err, domain.ErrValidation) {
			response.Error(w, http.StatusBadRequest, "VALIDATION_FAILED", "Nome cartella non valido o parametri incompleti")
			return
		}
		if errors.Is(err, domain.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "PARENT_NOT_FOUND", "Cartella genitore specificata non trovata")
			return
		}
		if errors.Is(err, domain.ErrConflict) {
			response.Error(w, http.StatusConflict, "CONFLICT", "Esiste già una cartella con questo nome")
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Errore durante la creazione della cartella")
		return
	}

	response.JSON(w, http.StatusCreated, toTargetGroupResponse(group))
}

// Update gestisce PUT /api/v1/target-groups/{id} per rinominare, descrivere o spostare una cartella
func (h *TargetGroupHandler) Update(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Tenant non identificato nel contesto")
		return
	}

	idStr := chi.URLParam(r, "id")
	groupID, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "Identificativo cartella non valido")
		return
	}

	var req UpdateTargetGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "Payload JSON non valido")
		return
	}

	group, err := h.groupService.UpdateGroup(r.Context(), service.UpdateGroupInput{
		TenantID:    tenantID,
		ID:          groupID,
		ParentID:    req.ParentID,
		Name:        req.Name,
		Description: req.Description,
		ColorTag:    req.ColorTag,
	})
	if err != nil {
		if errors.Is(err, domain.ErrValidation) {
			response.Error(w, http.StatusBadRequest, "VALIDATION_FAILED", "Aggiornamento non valido: nome vuoto o ciclo rilevato nella gerarchia")
			return
		}
		if errors.Is(err, domain.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "Cartella o genitore non trovato")
			return
		}
		if errors.Is(err, domain.ErrConflict) {
			response.Error(w, http.StatusConflict, "CONFLICT", "Nome cartella già utilizzato per questo livello")
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Errore durante l'aggiornamento della cartella")
		return
	}

	response.JSON(w, http.StatusOK, toTargetGroupResponse(group))
}

// Delete gestisce DELETE /api/v1/target-groups/{id} per eliminare una cartella
func (h *TargetGroupHandler) Delete(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Tenant non identificato nel contesto")
		return
	}

	idStr := chi.URLParam(r, "id")
	groupID, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "Identificativo cartella non valido")
		return
	}

	if err := h.groupService.DeleteGroup(r.Context(), tenantID, groupID); err != nil {
		if errors.Is(err, domain.ErrConflict) {
			response.Error(w, http.StatusConflict, "GROUP_IN_USE", "Rimuovere o riassegnare i canali di notifica collegati prima di eliminare l’azienda")
			return
		}
		if errors.Is(err, domain.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "Cartella non trovata")
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Errore durante l'eliminazione della cartella")
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "Cartella eliminata con successo"})
}

// StartScan gestisce POST /api/v1/target-groups/{id}/scan per avviare una scansione ricorsiva sul sotto-albero
func (h *TargetGroupHandler) StartScan(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Tenant non identificato nel contesto")
		return
	}

	idStr := chi.URLParam(r, "id")
	groupID, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "Identificativo cartella non valido")
		return
	}

	var req StartGroupScanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "Payload JSON non valido")
		return
	}

	if req.ProfileID == uuid.Nil {
		response.Error(w, http.StatusBadRequest, "MISSING_PROFILE", "Identificativo profilo di scansione obbligatorio")
		return
	}

	userID, _ := middleware.GetUserID(r.Context())
	ctx := domain.WithScanAuthorization(r.Context(), tenantID, userID, req.Authorization)
	scans, err := h.scanService.StartGroupScan(ctx, tenantID, groupID, req.ProfileID)
	if err != nil {
		if errors.Is(err, domain.ErrScanAuthorizationRequired) {
			response.Error(w, http.StatusConflict, "SCAN_AUTHORIZATION_REQUIRED", err.Error())
			return
		}
		if errors.Is(err, domain.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "Cartella o profilo di scansione non trovato")
			return
		}
		if errors.Is(err, domain.ErrValidation) {
			response.Error(w, http.StatusBadRequest, "VALIDATION_FAILED", "Parametri di avvio scansione non validi")
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Errore durante l'avvio della scansione cartella")
		return
	}

	type GroupScanResponse struct {
		GroupID      string         `json:"group_id"`
		ProfileID    string         `json:"profile_id"`
		ScansCount   int            `json:"scans_count"`
		CreatedScans []ScanResponse `json:"created_scans"`
	}

	var scanResponses []ScanResponse
	for _, s := range scans {
		scanResponses = append(scanResponses, toScanResponse(s))
	}

	response.JSON(w, http.StatusAccepted, GroupScanResponse{
		GroupID:      groupID.String(),
		ProfileID:    req.ProfileID.String(),
		ScansCount:   len(scans),
		CreatedScans: scanResponses,
	})
}

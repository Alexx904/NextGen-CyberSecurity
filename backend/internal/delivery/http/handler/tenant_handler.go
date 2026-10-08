package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"nextgen-cybersec/internal/delivery/http/middleware"
	"nextgen-cybersec/internal/delivery/http/response"
	"nextgen-cybersec/internal/domain"
	"nextgen-cybersec/internal/service"
)

// UpdateBrandingRequest definisce il payload per la personalizzazione grafica e anagrafica del tenant
type UpdateBrandingRequest struct {
	Name        string              `json:"name"`
	LogoURL     *string             `json:"logo_url"`
	BrandColors *domain.BrandColors `json:"brand_colors"`
}

// TenantResponse modella la rappresentazione pubblica del tenant con il branding configurato
type TenantResponse struct {
	ID          string             `json:"id"`
	Name        string             `json:"name"`
	Slug        string             `json:"slug"`
	LogoURL     *string            `json:"logo_url,omitempty"`
	BrandColors domain.BrandColors `json:"brand_colors"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
}

// TenantHandler espone gli endpoint HTTP per la consultazione e gestione del tenant corrente
type TenantHandler struct {
	tenantService service.TenantService
}

// NewTenantHandler crea una nuova istanza di TenantHandler
func NewTenantHandler(tenantService service.TenantService) *TenantHandler {
	return &TenantHandler{
		tenantService: tenantService,
	}
}

// GetCurrent gestisce GET /api/v1/tenants/current
func (h *TenantHandler) GetCurrent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, ok := middleware.GetTenantID(ctx)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Tenant non identificato nel contesto")
		return
	}

	tenant, err := h.tenantService.GetTenant(ctx, tenantID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "TENANT_NOT_FOUND", "Organizzazione tenant non trovata")
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Errore durante il recupero delle informazioni del tenant")
		return
	}

	response.JSON(w, http.StatusOK, toTenantResponse(tenant))
}

// UpdateBranding gestisce PUT /api/v1/tenants/current/branding
func (h *TenantHandler) UpdateBranding(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, ok := middleware.GetTenantID(ctx)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Tenant non identificato nel contesto")
		return
	}

	var req UpdateBrandingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_PAYLOAD", "Payload JSON della richiesta non valido")
		return
	}

	var colors domain.BrandColors
	if req.BrandColors != nil {
		colors = *req.BrandColors
	} else {
		colors = domain.DefaultBrandColors()
	}

	updatedTenant, err := h.tenantService.UpdateBranding(ctx, service.UpdateTenantBrandingInput{
		TenantID:    tenantID,
		Name:        req.Name,
		LogoURL:     req.LogoURL,
		BrandColors: colors,
	})
	if err != nil {
		if errors.Is(err, domain.ErrValidation) {
			response.Error(w, http.StatusBadRequest, "VALIDATION_FAILED", "Dati di branding non validi (ragione sociale vuota o colori HEX errati)")
			return
		}
		if errors.Is(err, domain.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "TENANT_NOT_FOUND", "Organizzazione tenant non trovata")
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Errore durante l'aggiornamento del branding tenant")
		return
	}

	response.JSON(w, http.StatusOK, toTenantResponse(updatedTenant))
}

func toTenantResponse(t *domain.Tenant) TenantResponse {
	return TenantResponse{
		ID:          t.ID.String(),
		Name:        t.Name,
		Slug:        t.Slug,
		LogoURL:     t.LogoURL,
		BrandColors: t.BrandColors,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}


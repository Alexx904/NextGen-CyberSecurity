package handler

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"nextgen-cybersec/internal/delivery/http/middleware"
	"nextgen-cybersec/internal/delivery/http/response"
	"nextgen-cybersec/internal/domain"
	"nextgen-cybersec/internal/service"
)

type VirtualPatchingHandler struct {
	service *service.VirtualPatchingService
}

func NewVirtualPatchingHandler(s *service.VirtualPatchingService) *VirtualPatchingHandler {
	return &VirtualPatchingHandler{service: s}
}

func (h *VirtualPatchingHandler) Generate(w http.ResponseWriter, r *http.Request) {
	tenant, ok := middleware.GetTenantID(r.Context())
	if !ok || tenant == uuid.Nil {
		response.Error(w, 401, "UNAUTHORIZED", "Tenant non identificato")
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil || id == uuid.Nil {
		response.Error(w, 400, "VALIDATION_ERROR", "ID finding non valido")
		return
	}
	bundle, err := h.service.Generate(r.Context(), tenant, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			response.Error(w, 404, "NOT_FOUND", "Finding non trovato")
		} else {
			response.Error(w, 500, "INTERNAL_ERROR", "Generazione template non disponibile")
		}
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	response.JSON(w, http.StatusOK, bundle)
}

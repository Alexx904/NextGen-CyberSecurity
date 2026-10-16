package handler

import (
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
	"nextgen-cybersec/internal/delivery/http/middleware"
	"nextgen-cybersec/internal/delivery/http/response"
	"nextgen-cybersec/internal/domain"
	"nextgen-cybersec/internal/service"
)

type DefenseHandler struct{ service service.DefenseService }

func NewDefenseHandler(s service.DefenseService) *DefenseHandler           { return &DefenseHandler{service: s} }
func (h *DefenseHandler) Preview(w http.ResponseWriter, r *http.Request)   { h.handle(w, r, false) }
func (h *DefenseHandler) Remediate(w http.ResponseWriter, r *http.Request) { h.handle(w, r, true) }
func (h *DefenseHandler) handle(w http.ResponseWriter, r *http.Request, apply bool) {
	tenant, ok := middleware.GetTenantID(r.Context())
	if !ok {
		response.Error(w, 401, "UNAUTHORIZED", "Tenant non identificato")
		return
	}
	var req domain.RemediationRequest
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		response.Error(w, 400, "VALIDATION_ERROR", "Richiesta non valida")
		return
	}
	if err := dec.Decode(new(any)); err != io.EOF {
		response.Error(w, 400, "VALIDATION_ERROR", "Richiesta non valida")
		return
	}
	var result *domain.RemediationResult
	var err error
	if apply {
		result, err = h.service.Remediate(r.Context(), tenant, chi.URLParam(r, "id"), req)
	} else {
		result, err = h.service.Preview(r.Context(), tenant, chi.URLParam(r, "id"), req)
	}
	if err != nil {
		status, code := http.StatusBadGateway, "REMEDIATION_FAILED"
		switch {
		case errors.Is(err, domain.ErrValidation):
			status, code = 400, "VALIDATION_ERROR"
		case errors.Is(err, domain.ErrNotFound), errors.Is(err, domain.ErrSSHServerNotFound):
			status, code = 404, "NOT_FOUND"
		case errors.Is(err, domain.ErrSafetyTimerActive):
			status, code = 409, "SAFETY_TIMER_ACTIVE"
		}
		response.Error(w, status, code, "Stato mitigazione non confermato; verificare servizio, server e stato firewall")
		return
	}
	response.JSON(w, http.StatusOK, result)
}

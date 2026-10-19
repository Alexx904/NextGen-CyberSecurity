package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"nextgen-cybersec/internal/delivery/http/middleware"
	"nextgen-cybersec/internal/delivery/http/response"
	"nextgen-cybersec/internal/domain"
	"nextgen-cybersec/internal/service"
)

type AlertHandler struct{ service *service.AlertService }

func NewAlertHandler(s *service.AlertService) *AlertHandler { return &AlertHandler{service: s} }

func alertTenant(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	tenant, ok := middleware.GetTenantID(r.Context())
	if !ok || tenant == uuid.Nil {
		response.Error(w, 401, "UNAUTHORIZED", "Tenant non identificato")
		return uuid.Nil, false
	}
	w.Header().Set("Cache-Control", "no-store")
	return tenant, true
}
func alertID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil || id == uuid.Nil {
		response.Error(w, 400, "VALIDATION_ERROR", "ID canale non valido")
		return uuid.Nil, false
	}
	return id, true
}
func alertDecode(w http.ResponseWriter, r *http.Request, out any) bool {
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16384))
	decoder.DisallowUnknownFields()
	if decoder.Decode(out) != nil || decoder.Decode(new(any)) != io.EOF {
		response.Error(w, 400, "VALIDATION_ERROR", "Configurazione canale non valida")
		return false
	}
	return true
}
func alertError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrAlertEncryptionUnavailable):
		response.Error(w, 503, "ALERT_ENCRYPTION_UNAVAILABLE", "Configurare una master key non predefinita di almeno 32 byte prima di usare le notifiche")
	case errors.Is(err, domain.ErrValidation):
		response.Error(w, 400, "VALIDATION_ERROR", "Configurazione canale non valida")
	case errors.Is(err, domain.ErrNotFound):
		response.Error(w, 404, "NOT_FOUND", "Canale o azienda non trovato")
	case errors.Is(err, domain.ErrConflict):
		response.Error(w, 409, "STALE_CHANNEL", "Canale modificato: ricaricare prima di salvare")
	default:
		response.Error(w, 500, "INTERNAL_ERROR", "Operazione canale non disponibile")
	}
}
func (h *AlertHandler) List(w http.ResponseWriter, r *http.Request) {
	tenant, ok := alertTenant(w, r)
	if !ok {
		return
	}
	channels, err := h.service.List(r.Context(), tenant)
	if err != nil {
		alertError(w, err)
		return
	}
	response.JSON(w, 200, channels)
}
func (h *AlertHandler) Create(w http.ResponseWriter, r *http.Request) { h.save(w, r, true) }
func (h *AlertHandler) Update(w http.ResponseWriter, r *http.Request) { h.save(w, r, false) }
func (h *AlertHandler) save(w http.ResponseWriter, r *http.Request, create bool) {
	tenant, ok := alertTenant(w, r)
	if !ok {
		return
	}
	id := uuid.Nil
	if !create {
		id, ok = alertID(w, r)
		if !ok {
			return
		}
	}
	var req domain.AlertChannelRequest
	if !alertDecode(w, r, &req) {
		return
	}
	channel, err := h.service.Save(r.Context(), tenant, id, req)
	if err != nil {
		alertError(w, err)
		return
	}
	status := 200
	if create {
		status = 201
	}
	response.JSON(w, status, channel)
}
func (h *AlertHandler) Delete(w http.ResponseWriter, r *http.Request) {
	tenant, ok := alertTenant(w, r)
	if !ok {
		return
	}
	id, ok := alertID(w, r)
	if !ok {
		return
	}
	if err := h.service.Delete(r.Context(), tenant, id); err != nil {
		alertError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
func (h *AlertHandler) Test(w http.ResponseWriter, r *http.Request) {
	tenant, ok := alertTenant(w, r)
	if !ok {
		return
	}
	var req struct {
		Kind      domain.AlertChannelKind   `json:"kind"`
		Config    domain.AlertChannelConfig `json:"config"`
		Confirmed bool                      `json:"confirmed_external_send"`
	}
	if !alertDecode(w, r, &req) {
		return
	}
	result, err := h.service.Test(r.Context(), tenant, req.Kind, req.Config, req.Confirmed)
	if err != nil {
		alertError(w, err)
		return
	}
	if result.Code != "ACCEPTED" {
		response.Error(w, 502, result.Code, "Messaggio di prova non accettato; verificare la configurazione. Nessun retry automatico.")
		return
	}
	response.JSON(w, 200, map[string]any{"accepted": true, "code": result.Code, "message": "Provider ha accettato il messaggio. Verificare la ricezione nel canale."})
}

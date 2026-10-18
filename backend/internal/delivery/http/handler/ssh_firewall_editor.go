package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"nextgen-cybersec/internal/delivery/http/middleware"
	"nextgen-cybersec/internal/delivery/http/response"
	"nextgen-cybersec/internal/domain"
	"nextgen-cybersec/internal/service"
)

type FirewallPolicyRequest struct {
	Chain                domain.ChainName  `json:"chain"`
	Policy               domain.RuleAction `json:"policy"`
	ExpectedRevision     string            `json:"expected_revision"`
	SafetyTimeoutSeconds int               `json:"safety_timeout_seconds"`
}
type FirewallMoveRequest struct {
	Chain                domain.ChainName `json:"chain"`
	From                 int              `json:"from"`
	To                   int              `json:"to"`
	ExpectedRevision     string           `json:"expected_revision"`
	SafetyTimeoutSeconds int              `json:"safety_timeout_seconds"`
}

func firewallScope(w http.ResponseWriter, r *http.Request) (uuid.UUID, uuid.UUID, bool) {
	tenant, ok := middleware.GetTenantID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Tenant non identificato")
		return uuid.Nil, uuid.Nil, false
	}
	server, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil || server == uuid.Nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "Identificativo server non valido")
		return uuid.Nil, uuid.Nil, false
	}
	return tenant, server, true
}
func (h *SSHFirewallHandler) firewallEditor(w http.ResponseWriter) (service.FirewallRuleEditor, bool) {
	editor, ok := h.firewallService.(service.FirewallRuleEditor)
	if !ok {
		response.Error(w, http.StatusNotImplemented, "FIREWALL_EDITOR_UNAVAILABLE", "Operazioni firewall non disponibili")
	}
	return editor, ok
}
func writeFirewallCommonError(w http.ResponseWriter, err error) bool {
	switch {
	case errors.Is(err, domain.ErrSafetyTimerActive):
		response.Error(w, http.StatusConflict, "SAFETY_TIMER_ACTIVE", domain.ErrSafetyTimerActive.Error())
	case errors.Is(err, domain.ErrNoSafetyTimerActive):
		response.Error(w, http.StatusConflict, "NO_SAFETY_TIMER_ACTIVE", domain.ErrNoSafetyTimerActive.Error())
	case errors.Is(err, domain.ErrFirewallStaleState):
		response.Error(w, http.StatusConflict, "FIREWALL_STATE_CHANGED", domain.ErrFirewallStaleState.Error())
	case errors.Is(err, domain.ErrFirewallPrivilegesRequired):
		response.Error(w, http.StatusBadGateway, "FIREWALL_PRIVILEGES_REQUIRED", domain.ErrFirewallPrivilegesRequired.Error())
	case errors.Is(err, domain.ErrFirewallUnavailable):
		response.Error(w, http.StatusBadGateway, "FIREWALL_UNAVAILABLE", domain.ErrFirewallUnavailable.Error())
	default:
		return false
	}
	return true
}
func writeFirewallEditResult(w http.ResponseWriter, status *domain.FirewallStatus, err error) {
	if err == nil {
		response.JSON(w, http.StatusOK, status)
		return
	}
	if writeFirewallCommonError(w, err) {
		return
	}
	switch {
	case errors.Is(err, domain.ErrValidation), errors.Is(err, domain.ErrInvalidChainName), errors.Is(err, domain.ErrInvalidRuleAction):
		response.Error(w, http.StatusBadRequest, "VALIDATION_FAILED", err.Error())
	case errors.Is(err, domain.ErrSSHServerNotFound), errors.Is(err, domain.ErrNotFound):
		response.Error(w, http.StatusNotFound, "NOT_FOUND", "Server SSH non trovato")
	case errors.Is(err, domain.ErrSSHExecutionFailed):
		response.Error(w, http.StatusBadGateway, "SSH_EXECUTION_FAILED", "Modifica firewall remoto fallita; verificare lo stato")
	default:
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Errore durante la modifica firewall")
	}
}
func decodeFirewallEdit(w http.ResponseWriter, r *http.Request, target any) bool {
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 32*1024))
	if err := decoder.Decode(target); err != nil {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "Payload JSON non valido")
		return false
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "Payload JSON non valido")
		return false
	}
	return true
}
func (h *SSHFirewallHandler) SetFirewallPolicy(w http.ResponseWriter, r *http.Request) {
	tenant, server, ok := firewallScope(w, r)
	if !ok {
		return
	}
	var req FirewallPolicyRequest
	if !decodeFirewallEdit(w, r, &req) {
		return
	}
	editor, ok := h.firewallEditor(w)
	if !ok {
		return
	}
	status, err := editor.SetPolicy(r.Context(), tenant, server, req.Chain, req.Policy, req.ExpectedRevision, req.SafetyTimeoutSeconds)
	writeFirewallEditResult(w, status, err)
}
func (h *SSHFirewallHandler) ReplaceFirewallRule(w http.ResponseWriter, r *http.Request) {
	tenant, server, ok := firewallScope(w, r)
	if !ok {
		return
	}
	line, err := strconv.Atoi(chi.URLParam(r, "line"))
	if err != nil || line < 1 {
		response.Error(w, http.StatusBadRequest, "INVALID_LINE_NUMBER", "Numero riga non valido")
		return
	}
	var req AddFirewallRuleRequest
	if !decodeFirewallEdit(w, r, &req) {
		return
	}
	editor, ok := h.firewallEditor(w)
	if !ok {
		return
	}
	params := domain.FirewallRuleParams{Chain: req.Chain, Action: req.Action, Protocol: req.Protocol, Source: req.Source, Destination: req.Destination, Port: req.Port, InInterface: req.InInterface, OutInterface: req.OutInterface, Comment: req.Comment, Position: req.Position}
	status, err := editor.ReplaceRule(r.Context(), tenant, server, line, params, req.ExpectedRevision, req.SafetyTimeoutSeconds)
	writeFirewallEditResult(w, status, err)
}
func (h *SSHFirewallHandler) MoveFirewallRule(w http.ResponseWriter, r *http.Request) {
	tenant, server, ok := firewallScope(w, r)
	if !ok {
		return
	}
	var req FirewallMoveRequest
	if !decodeFirewallEdit(w, r, &req) {
		return
	}
	editor, ok := h.firewallEditor(w)
	if !ok {
		return
	}
	status, err := editor.MoveRule(r.Context(), tenant, server, req.Chain, req.From, req.To, req.ExpectedRevision, req.SafetyTimeoutSeconds)
	writeFirewallEditResult(w, status, err)
}

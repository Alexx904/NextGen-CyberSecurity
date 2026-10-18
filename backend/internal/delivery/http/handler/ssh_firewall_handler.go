package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"nextgen-cybersec/internal/delivery/http/middleware"
	"nextgen-cybersec/internal/delivery/http/response"
	"nextgen-cybersec/internal/domain"
	"nextgen-cybersec/internal/service"
)

// SSHServerResponse modella la rappresentazione JSON pubblica di un server SSH (secret sempre omesso)
type SSHServerResponse struct {
	HostKeyFingerprint   string     `json:"host_key_fingerprint"`
	ID                   string     `json:"id"`
	TenantID             string     `json:"tenant_id"`
	Name                 string     `json:"name"`
	Host                 string     `json:"host"`
	Port                 int        `json:"port"`
	Username             string     `json:"username"`
	AuthType             string     `json:"auth_type"`
	PublicKeyFingerprint *string    `json:"public_key_fingerprint,omitempty"`
	Status               string     `json:"status"`
	LastConnectedAt      *time.Time `json:"last_connected_at,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

// AddFirewallRuleRequest modella il payload JSON per l'inserimento di una nuova regola firewall
type AddFirewallRuleRequest struct {
	Chain                domain.ChainName  `json:"chain"`
	Action               domain.RuleAction `json:"action"`
	Protocol             string            `json:"protocol"`
	Source               string            `json:"source"`
	Destination          string            `json:"destination"`
	Port                 string            `json:"port"`
	InInterface          string            `json:"in_interface"`
	OutInterface         string            `json:"out_interface"`
	Comment              string            `json:"comment"`
	Position             int               `json:"position"`
	WithSafetyTimer      bool              `json:"with_safety_timer"`
	SafetyTimeoutSeconds int               `json:"safety_timeout_seconds"`
}

// ApplyFirewallRequest modella la richiesta di attivazione manuale del Safety Rollback Timer
type ApplyFirewallRequest struct {
	SafetyTimeoutSeconds int `json:"safety_timeout_seconds"`
}

// RollbackFirewallRequest modella la richiesta di ripristino ad uno snapshot firewall specifico
type RollbackFirewallRequest struct {
	BackupID *uuid.UUID `json:"backup_id,omitempty"`
}

func toSSHServerResponse(s *domain.SSHServer) SSHServerResponse {
	return SSHServerResponse{
		HostKeyFingerprint:   s.HostKeyFingerprint,
		ID:                   s.ID.String(),
		TenantID:             s.TenantID.String(),
		Name:                 s.Name,
		Host:                 s.Host,
		Port:                 s.Port,
		Username:             s.Username,
		AuthType:             string(s.AuthType),
		PublicKeyFingerprint: s.PublicKeyFingerprint,
		Status:               string(s.Status),
		LastConnectedAt:      s.LastConnectedAt,
		CreatedAt:            s.CreatedAt,
		UpdatedAt:            s.UpdatedAt,
	}
}

// SSHFirewallHandler gestisce gli endpoint REST per l'amministrazione dei server SSH e del firewall iptables
type SSHFirewallHandler struct {
	sshService      service.SSHService
	firewallService service.FirewallService
}

// NewSSHFirewallHandler crea una nuova istanza di SSHFirewallHandler
func NewSSHFirewallHandler(
	sshService service.SSHService,
	firewallService service.FirewallService,
) *SSHFirewallHandler {
	return &SSHFirewallHandler{
		sshService:      sshService,
		firewallService: firewallService,
	}
}

// ListServers gestisce GET /api/v1/ssh/servers per elencare tutti i server del tenant
func (h *SSHFirewallHandler) ListServers(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Tenant non identificato nel contesto")
		return
	}

	servers, err := h.sshService.ListServers(r.Context(), tenantID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Impossibile recuperare i server SSH")
		return
	}

	items := make([]SSHServerResponse, 0, len(servers))
	for _, s := range servers {
		items = append(items, toSSHServerResponse(s))
	}

	response.JSON(w, http.StatusOK, items)
}

// GetServer gestisce GET /api/v1/ssh/servers/{id} per recuperare i dettagli di un singolo server
func (h *SSHFirewallHandler) GetServer(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Tenant non identificato nel contesto")
		return
	}

	serverID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "Identificativo server non valido")
		return
	}

	server, err := h.sshService.GetServer(r.Context(), tenantID, serverID)
	if err != nil {
		if errors.Is(err, domain.ErrSSHServerNotFound) || errors.Is(err, domain.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "Server SSH non trovato")
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Errore durante il recupero del server SSH")
		return
	}

	response.JSON(w, http.StatusOK, toSSHServerResponse(server))
}

// CreateServer gestisce POST /api/v1/ssh/servers per registrare un nuovo server (cifratura AES-256-GCM)
func (h *SSHFirewallHandler) CreateServer(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Tenant non identificato nel contesto")
		return
	}

	var req service.CreateSSHServerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "Payload JSON non valido")
		return
	}

	server, err := h.sshService.CreateServer(r.Context(), tenantID, req)
	if err != nil {
		if errors.Is(err, domain.ErrSSHServerDuplicateName) {
			response.Error(w, http.StatusConflict, "DUPLICATE_NAME", "Esiste già un server registrato con questo nome")
			return
		}
		if errors.Is(err, domain.ErrEmptySecret) {
			response.Error(w, http.StatusBadRequest, "EMPTY_SECRET", "La credenziale di autenticazione SSH (password o chiave privata) non può essere vuota")
			return
		}
		if errors.Is(err, domain.ErrInvalidSSHAuthType) {
			response.Error(w, http.StatusBadRequest, "INVALID_AUTH_TYPE", "Tipo di autenticazione SSH non valido (richiesto KEY o PASSWORD)")
			return
		}
		if errors.Is(err, domain.ErrValidation) {
			response.Error(w, http.StatusBadRequest, "VALIDATION_FAILED", "Parametri di registrazione server non validi")
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Errore durante la registrazione del server SSH")
		return
	}

	response.JSON(w, http.StatusCreated, toSSHServerResponse(server))
}

// UpdateServer gestisce PUT /api/v1/ssh/servers/{id} per aggiornare metadati o rinnovare il secret
func (h *SSHFirewallHandler) UpdateServer(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Tenant non identificato nel contesto")
		return
	}

	serverID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "Identificativo server non valido")
		return
	}

	var req service.UpdateSSHServerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "Payload JSON non valido")
		return
	}

	server, err := h.sshService.UpdateServer(r.Context(), tenantID, serverID, req)
	if err != nil {
		if errors.Is(err, domain.ErrSSHServerNotFound) || errors.Is(err, domain.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "Server SSH non trovato")
			return
		}
		if errors.Is(err, domain.ErrSSHServerDuplicateName) {
			response.Error(w, http.StatusConflict, "DUPLICATE_NAME", "Esiste già un server registrato con questo nome")
			return
		}
		if errors.Is(err, domain.ErrValidation) {
			response.Error(w, http.StatusBadRequest, "VALIDATION_FAILED", "Parametri di aggiornamento server non validi")
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Errore durante l'aggiornamento del server SSH")
		return
	}

	response.JSON(w, http.StatusOK, toSSHServerResponse(server))
}

// DeleteServer gestisce DELETE /api/v1/ssh/servers/{id} per eliminare un server e i relativi snapshot
func (h *SSHFirewallHandler) DeleteServer(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Tenant non identificato nel contesto")
		return
	}

	serverID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "Identificativo server non valido")
		return
	}

	if err := h.sshService.DeleteServer(r.Context(), tenantID, serverID); err != nil {
		if errors.Is(err, domain.ErrSSHServerNotFound) || errors.Is(err, domain.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "Server SSH non trovato")
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Errore durante l'eliminazione del server SSH")
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "Server SSH eliminato con successo"})
}

// TestConnection gestisce POST /api/v1/ssh/servers/{id}/test-connection
func (h *SSHFirewallHandler) TestConnection(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Tenant non identificato nel contesto")
		return
	}

	serverID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "Identificativo server non valido")
		return
	}

	result, err := h.sshService.TestConnection(r.Context(), tenantID, serverID)
	if err != nil {
		if errors.Is(err, domain.ErrSSHServerNotFound) || errors.Is(err, domain.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "Server SSH non trovato")
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Errore durante il test di connettività SSH")
		return
	}

	response.JSON(w, http.StatusOK, result)
}

// ============================================================================
// Endpoint Firewall iptables (/api/v1/ssh/servers/{id}/firewall)
// ============================================================================

// GetFirewallStatus gestisce GET /api/v1/ssh/servers/{id}/firewall per recuperare catene e regole iptables
func (h *SSHFirewallHandler) GetFirewallStatus(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Tenant non identificato nel contesto")
		return
	}

	serverID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "Identificativo server non valido")
		return
	}

	status, err := h.firewallService.GetStatus(r.Context(), tenantID, serverID)
	if err != nil {
		if errors.Is(err, domain.ErrFirewallUnavailable) {
			response.Error(w, http.StatusBadGateway, "FIREWALL_UNAVAILABLE", domain.ErrFirewallUnavailable.Error())
			return
		}
		if errors.Is(err, domain.ErrFirewallPrivilegesRequired) {
			response.Error(w, http.StatusBadGateway, "FIREWALL_PRIVILEGES_REQUIRED", domain.ErrFirewallPrivilegesRequired.Error())
			return
		}
		if errors.Is(err, domain.ErrSSHServerNotFound) || errors.Is(err, domain.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "Server SSH non trovato")
			return
		}
		if errors.Is(err, domain.ErrSSHExecutionFailed) {
			response.Error(w, http.StatusBadGateway, "SSH_EXECUTION_FAILED", "Impossibile recuperare lo stato del firewall dal server remoto")
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Errore durante la lettura del firewall")
		return
	}

	response.JSON(w, http.StatusOK, status)
}

// AddFirewallRule gestisce POST /api/v1/ssh/servers/{id}/firewall/rules per inserire una nuova regola
func (h *SSHFirewallHandler) AddFirewallRule(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Tenant non identificato nel contesto")
		return
	}

	serverID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "Identificativo server non valido")
		return
	}

	var req AddFirewallRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "Payload JSON non valido")
		return
	}

	params := domain.FirewallRuleParams{
		Chain:        req.Chain,
		Action:       req.Action,
		Protocol:     req.Protocol,
		Source:       req.Source,
		Destination:  req.Destination,
		Port:         req.Port,
		InInterface:  req.InInterface,
		OutInterface: req.OutInterface,
		Comment:      req.Comment,
		Position:     req.Position,
	}

	timeoutSec := req.SafetyTimeoutSeconds
	if timeoutSec == 0 {
		timeoutSec = 60
	}

	status, err := h.firewallService.AddRule(r.Context(), tenantID, serverID, params, req.WithSafetyTimer, timeoutSec)
	if err != nil {
		if errors.Is(err, domain.ErrSSHServerNotFound) || errors.Is(err, domain.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "Server SSH non trovato")
			return
		}
		if errors.Is(err, domain.ErrInvalidChainName) {
			response.Error(w, http.StatusBadRequest, "INVALID_CHAIN", "Nome catena non valido (atteso INPUT, OUTPUT o FORWARD)")
			return
		}
		if errors.Is(err, domain.ErrInvalidRuleAction) {
			response.Error(w, http.StatusBadRequest, "INVALID_ACTION", "Azione regola non valida (atteso ACCEPT, DROP, REJECT o LOG)")
			return
		}
		if errors.Is(err, domain.ErrValidation) {
			response.Error(w, http.StatusBadRequest, "VALIDATION_FAILED", err.Error())
			return
		}
		if errors.Is(err, domain.ErrSSHExecutionFailed) {
			response.Error(w, http.StatusBadGateway, "SSH_EXECUTION_FAILED", "Applicazione regola firewall su host remoto fallita")
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Errore durante l'aggiunta della regola firewall")
		return
	}

	response.JSON(w, http.StatusOK, status)
}

// DeleteFirewallRule gestisce DELETE /api/v1/ssh/servers/{id}/firewall/rules/{line}?chain=INPUT
func (h *SSHFirewallHandler) DeleteFirewallRule(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Tenant non identificato nel contesto")
		return
	}

	serverID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "Identificativo server non valido")
		return
	}

	lineStr := chi.URLParam(r, "line")
	lineNum, err := strconv.Atoi(lineStr)
	if err != nil || lineNum <= 0 {
		response.Error(w, http.StatusBadRequest, "INVALID_LINE_NUMBER", "Numero riga regola non valido (intero positivo atteso)")
		return
	}

	chainStr := r.URL.Query().Get("chain")
	if chainStr == "" {
		chainStr = "INPUT"
	}
	chain := domain.ChainName(strings.ToUpper(strings.TrimSpace(chainStr)))
	if !chain.IsValid() {
		response.Error(w, http.StatusBadRequest, "INVALID_CHAIN", "Catena firewall non valida (atteso INPUT, OUTPUT o FORWARD)")
		return
	}

	withSafetyTimer := true
	if stStr := r.URL.Query().Get("safety_timer"); stStr != "" {
		if val, parseErr := strconv.ParseBool(stStr); parseErr == nil {
			withSafetyTimer = val
		}
	}

	timeoutSec := 60
	if toStr := r.URL.Query().Get("timeout"); toStr != "" {
		val, parseErr := strconv.Atoi(toStr)
		if parseErr != nil {
			response.Error(w, http.StatusBadRequest, "VALIDATION_FAILED", "Timeout non valido")
			return
		}
		timeoutSec = val
	}

	status, err := h.firewallService.DeleteRule(r.Context(), tenantID, serverID, chain, lineNum, withSafetyTimer, timeoutSec)
	if err != nil {
		if errors.Is(err, domain.ErrValidation) {
			response.Error(w, http.StatusBadRequest, "VALIDATION_FAILED", err.Error())
			return
		}
		if errors.Is(err, domain.ErrSSHServerNotFound) || errors.Is(err, domain.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "Server SSH non trovato")
			return
		}
		if errors.Is(err, domain.ErrSSHExecutionFailed) {
			response.Error(w, http.StatusBadGateway, "SSH_EXECUTION_FAILED", "Eliminazione regola su host remoto fallita")
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Errore durante l'eliminazione della regola firewall")
		return
	}

	response.JSON(w, http.StatusOK, status)
}

// ApplyWithSafetyTimer gestisce POST /api/v1/ssh/servers/{id}/firewall/apply per attivare il timer di rollback
func (h *SSHFirewallHandler) ApplyWithSafetyTimer(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Tenant non identificato nel contesto")
		return
	}

	serverID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "Identificativo server non valido")
		return
	}

	var req ApplyFirewallRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	timeoutSec := req.SafetyTimeoutSeconds
	if timeoutSec == 0 {
		timeoutSec = 60
	}

	status, err := h.firewallService.ApplyWithSafetyTimer(r.Context(), tenantID, serverID, timeoutSec)
	if err != nil {
		if errors.Is(err, domain.ErrValidation) {
			response.Error(w, http.StatusBadRequest, "VALIDATION_FAILED", err.Error())
			return
		}
		if errors.Is(err, domain.ErrSSHServerNotFound) || errors.Is(err, domain.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "Server SSH non trovato")
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Errore durante l'attivazione del safety timer")
		return
	}

	response.JSON(w, http.StatusOK, status)
}

// ConfirmFirewall gestisce POST /api/v1/ssh/servers/{id}/firewall/confirm per confermare le regole applicate
func (h *SSHFirewallHandler) ConfirmFirewall(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Tenant non identificato nel contesto")
		return
	}

	serverID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "Identificativo server non valido")
		return
	}

	if err := h.firewallService.Confirm(r.Context(), tenantID, serverID); err != nil {
		if errors.Is(err, domain.ErrSSHServerNotFound) || errors.Is(err, domain.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "Server SSH non trovato")
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Errore durante la conferma della configurazione firewall")
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{
		"message": "Configurazione firewall confermata con successo: Safety Timer disattivato",
	})
}

// RollbackFirewall gestisce POST /api/v1/ssh/servers/{id}/firewall/rollback per il ripristino manuale immediato
func (h *SSHFirewallHandler) RollbackFirewall(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Tenant non identificato nel contesto")
		return
	}

	serverID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "Identificativo server non valido")
		return
	}

	var req RollbackFirewallRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	if err := h.firewallService.Rollback(r.Context(), tenantID, serverID, req.BackupID); err != nil {
		if errors.Is(err, domain.ErrSSHServerNotFound) || errors.Is(err, domain.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "Server SSH o snapshot backup non trovato")
			return
		}
		if errors.Is(err, domain.ErrSSHExecutionFailed) {
			response.Error(w, http.StatusBadGateway, "SSH_EXECUTION_FAILED", "Rollback regole firewall su host remoto fallito")
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Errore durante il rollback delle regole firewall")
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{
		"message": "Rollback regole firewall completato con successo allo snapshot precedente",
	})
}

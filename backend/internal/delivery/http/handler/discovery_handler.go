package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"nextgen-cybersec/internal/delivery/http/middleware"
	"nextgen-cybersec/internal/delivery/http/response"
	"nextgen-cybersec/internal/domain"
	"nextgen-cybersec/internal/service"
)

// ScanSubnetRequest modella il payload JSON per avviare una scansione rapida della subnet
type ScanSubnetRequest struct {
	Subnet             string     `json:"subnet"`
	IncludeCommonPorts bool       `json:"include_common_ports"`
	SessionID          *uuid.UUID `json:"session_id,omitempty"`
}

// ImportDiscoveredDevice rappresenta un singolo dispositivo da importare come bersaglio
type ImportDiscoveredDevice struct {
	IP          string  `json:"ip"`
	Hostname    string  `json:"hostname,omitempty"`
	Vendor      string  `json:"vendor,omitempty"`
	Description *string `json:"description,omitempty"`
}

// ImportTargetsRequest modella la richiesta di importazione massiva di dispositivi
type ImportTargetsRequest struct {
	GroupID *uuid.UUID               `json:"group_id,omitempty"`
	Devices []ImportDiscoveredDevice `json:"devices"`
}

// ImportItemError dettaglio di errore per un dispositivo scartato durante l'importazione
type ImportItemError struct {
	IP     string `json:"ip"`
	Reason string `json:"reason"`
}

// ImportTargetsResponse modella il riepilogo dell'importazione massiva
type ImportTargetsResponse struct {
	GroupID       *string           `json:"group_id,omitempty"`
	TotalImported int               `json:"total_imported"`
	TotalSkipped  int               `json:"total_skipped"`
	Imported      []TargetResponse  `json:"imported"`
	Errors        []ImportItemError `json:"errors,omitempty"`
}

// DiscoveryHandler gestisce gli endpoint REST per la ricognizione rapida LAN e l'importazione dei target
type DiscoveryHandler struct {
	discoveryService domain.DiscoveryService
	targetService    service.TargetService
	groupService     service.TargetGroupService
	broadcaster      service.EventBroadcaster
}

// NewDiscoveryHandler crea una nuova istanza di DiscoveryHandler
func NewDiscoveryHandler(
	discoveryService domain.DiscoveryService,
	targetService service.TargetService,
	groupService service.TargetGroupService,
	broadcaster service.EventBroadcaster,
) *DiscoveryHandler {
	return &DiscoveryHandler{
		discoveryService: discoveryService,
		targetService:    targetService,
		groupService:     groupService,
		broadcaster:      broadcaster,
	}
}

// Scan gestisce POST /api/v1/network-discovery/scan per eseguire un ping/ARP sweep su subnet o range IP
func (h *DiscoveryHandler) Scan(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, ok := middleware.GetTenantID(ctx)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Tenant non identificato nel contesto")
		return
	}

	var req ScanSubnetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "Payload JSON non valido")
		return
	}

	cleanSubnet := strings.TrimSpace(req.Subnet)
	if cleanSubnet == "" {
		response.Error(w, http.StatusBadRequest, "INVALID_SUBNET", "Subnet o range IP obbligatorio")
		return
	}

	result, err := h.discoveryService.DiscoverSubnet(ctx, tenantID, domain.DiscoveryRequest{
		Subnet:             cleanSubnet,
		IncludeCommonPorts: req.IncludeCommonPorts,
		SessionID:          req.SessionID,
	})
	if err != nil {
		if errors.Is(err, domain.ErrTargetInvalid) {
			response.Error(w, http.StatusBadRequest, "INVALID_SUBNET", "Subnet non valida o tipologia non supportata per LAN discovery")
			return
		}
		if errors.Is(err, domain.ErrValidation) {
			response.Error(w, http.StatusBadRequest, "VALIDATION_FAILED", "Parametri di richiesta non validi")
			return
		}
		if errors.Is(err, domain.ErrScannerExecFailed) {
			response.Error(w, http.StatusBadGateway, "SCANNER_ERROR", "Esecuzione dello scanner di rete fallita")
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Errore durante la scansione della rete locale")
		return
	}

	response.JSON(w, http.StatusOK, result)
}

// GetResults gestisce GET /api/v1/network-discovery/results per recuperare l'ultimo report di discovery del tenant
func (h *DiscoveryHandler) GetResults(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, ok := middleware.GetTenantID(ctx)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Tenant non identificato nel contesto")
		return
	}

	result, err := h.discoveryService.GetLastResults(ctx, tenantID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "Nessun risultato di discovery recente disponibile per il tenant")
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Errore durante il recupero dei risultati di discovery")
		return
	}

	response.JSON(w, http.StatusOK, result)
}

// Stream gestisce GET /api/v1/network-discovery/stream?session_id={session_id} per lo streaming live SSE degli host scoperti
func (h *DiscoveryHandler) Stream(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, ok := middleware.GetTenantID(ctx)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Tenant non identificato nel contesto")
		return
	}

	sessionIDStr := strings.TrimSpace(r.URL.Query().Get("session_id"))
	if sessionIDStr == "" {
		response.Error(w, http.StatusBadRequest, "INVALID_SESSION_ID", "Parametro 'session_id' mancante")
		return
	}

	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil || sessionID == uuid.Nil {
		response.Error(w, http.StatusBadRequest, "INVALID_SESSION_ID", "Parametro 'session_id' non è un UUID valido")
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		response.Error(w, http.StatusInternalServerError, "STREAMING_UNSUPPORTED", "HTTP streaming non supportato dal server")
		return
	}

	// Impostazione header conformi a text/event-stream e disattivazione buffering per reverse proxy Nginx
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	// Handshake iniziale SSE
	if _, err := fmt.Fprintf(w, ": connected\n\n"); err != nil {
		return
	}
	flusher.Flush()

	// Se il broadcaster non è configurato, mantieni la connessione fino alla chiusura del context
	if h.broadcaster == nil {
		<-ctx.Done()
		return
	}

	// Replay dispositivi già censiti per questa sessione (NFR-REL-05)
	if existingResult, err := h.discoveryService.GetSessionResults(ctx, tenantID, sessionID); err == nil && existingResult != nil {
		for _, dev := range existingResult.Devices {
			devEvent := domain.SSEEvent{
				Event: domain.EventDeviceDiscovered,
				Data: domain.DeviceDiscoveredPayload{
					SessionID: sessionID,
					Device:    dev,
				},
			}
			if b, err := devEvent.FormatBytes(); err == nil {
				_, _ = w.Write(b)
			}
		}
		flusher.Flush()

		// Se la sessione era già conclusa
		if !existingResult.CompletedAt.IsZero() {
			compEvent := domain.SSEEvent{
				Event: domain.EventDiscoveryComplete,
				Data: domain.DiscoveryCompletePayload{
					SessionID:       sessionID,
					TotalDiscovered: existingResult.TotalDiscovered,
					DurationSeconds: existingResult.DurationSeconds,
				},
			}
			if b, err := compEvent.FormatBytes(); err == nil {
				_, _ = w.Write(b)
				flusher.Flush()
			}
			return
		}
	}

	topic := domain.DiscoveryTopic(sessionID)
	ch, unsubscribe := h.broadcaster.Subscribe(ctx, topic)
	defer unsubscribe()

	for {
		select {
		case <-ctx.Done():
			return
		case evt, ok := <-ch:
			if !ok {
				return
			}
			bytes, err := evt.FormatBytes()
			if err != nil {
				continue
			}
			if _, err := w.Write(bytes); err != nil {
				return
			}
			flusher.Flush()

			if evt.Event == domain.EventDiscoveryComplete || evt.Event == domain.EventDiscoveryError {
				return
			}
		}
	}
}

// ImportTargets gestisce POST /api/v1/network-discovery/import-targets per importare dispositivi come Target
func (h *DiscoveryHandler) ImportTargets(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, ok := middleware.GetTenantID(ctx)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Tenant non identificato nel contesto")
		return
	}

	var req ImportTargetsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "Payload JSON non valido")
		return
	}

	if len(req.Devices) == 0 {
		response.Error(w, http.StatusBadRequest, "EMPTY_DEVICE_LIST", "Nessun dispositivo specificato per l'importazione")
		return
	}

	// Se specificato un gruppo, verifica l'esistenza e l'appartenenza al tenant prima di procedere
	if req.GroupID != nil && *req.GroupID != uuid.Nil {
		if _, err := h.groupService.GetGroup(ctx, tenantID, *req.GroupID); err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				response.Error(w, http.StatusNotFound, "GROUP_NOT_FOUND", "Cartella di destinazione non trovata")
				return
			}
			response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Errore durante la verifica della cartella")
			return
		}
	}

	seenIPs := make(map[string]bool)
	imported := make([]TargetResponse, 0, len(req.Devices))
	var itemErrors []ImportItemError

	for _, dev := range req.Devices {
		cleanIP := strings.TrimSpace(dev.IP)
		if cleanIP == "" {
			itemErrors = append(itemErrors, ImportItemError{
				IP:     dev.IP,
				Reason: "Indirizzo IP vuoto",
			})
			continue
		}

		if seenIPs[cleanIP] {
			itemErrors = append(itemErrors, ImportItemError{
				IP:     cleanIP,
				Reason: "Indirizzo IP duplicato nella richiesta",
			})
			continue
		}
		seenIPs[cleanIP] = true

		desc := buildDeviceDescription(dev)

		target, err := h.targetService.CreateTarget(ctx, service.CreateTargetInput{
			TenantID:      tenantID,
			GroupID:       req.GroupID,
			Value:         cleanIP,
			Description:   &desc,
			IsBlacklisted: false,
		})
		if err != nil {
			reason := "Errore durante la registrazione del target"
			if errors.Is(err, domain.ErrConflict) {
				reason = "Target già registrato per l'organizzazione"
			} else if errors.Is(err, domain.ErrTargetBlacklisted) {
				reason = "Target presente nella blacklist di esclusione"
			} else if errors.Is(err, domain.ErrTargetInvalid) {
				reason = "Formato indirizzo IP non valido"
			}
			itemErrors = append(itemErrors, ImportItemError{
				IP:     cleanIP,
				Reason: reason,
			})
			continue
		}

		imported = append(imported, toTargetResponse(target))
	}

	var groupIDStr *string
	if req.GroupID != nil && *req.GroupID != uuid.Nil {
		s := req.GroupID.String()
		groupIDStr = &s
	}

	resp := ImportTargetsResponse{
		GroupID:       groupIDStr,
		TotalImported: len(imported),
		TotalSkipped:  len(itemErrors),
		Imported:      imported,
		Errors:        itemErrors,
	}

	response.JSON(w, http.StatusOK, resp)
}

func buildDeviceDescription(dev ImportDiscoveredDevice) string {
	if dev.Description != nil && strings.TrimSpace(*dev.Description) != "" {
		return strings.TrimSpace(*dev.Description)
	}

	vendor := strings.TrimSpace(dev.Vendor)
	hostname := strings.TrimSpace(dev.Hostname)

	if vendor != "" && hostname != "" {
		return fmt.Sprintf("Discovered LAN device (%s - %s)", vendor, hostname)
	}
	if vendor != "" {
		return fmt.Sprintf("Discovered LAN device (%s)", vendor)
	}
	if hostname != "" {
		return fmt.Sprintf("Discovered LAN device (%s)", hostname)
	}
	return "Discovered LAN device"
}

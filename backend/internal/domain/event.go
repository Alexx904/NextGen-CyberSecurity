package domain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Costanti per gli eventi Server-Sent Events (SSE)
const (
	// Eventi LAN Network Discovery
	EventDeviceDiscovered  = "device_discovered"
	EventDiscoveryProgress = "progress"
	EventDiscoveryComplete = "complete"
	EventDiscoveryError    = "error"

	// Eventi Vulnerability Scanning Nmap
	EventScanPhase        = "phase"
	EventScanProgress     = "progress"
	EventScanFinding      = "finding_discovered"
	EventScanStatusChange = "status_change"
	EventScanError        = "error"
)

// SSEEvent rappresenta il frame standard per lo streaming Server-Sent Events (W3C standard)
type SSEEvent struct {
	Event string `json:"event"`
	Data  any    `json:"data"`
	ID    string `json:"id,omitempty"`
	Retry int    `json:"retry,omitempty"`
}

// FormatBytes formatta l'evento nel wire protocol standard text/event-stream:
//
//	id: <id>\n (opzionale)
//	event: <event>\n
//	retry: <retry>\n (opzionale)
//	data: <json>\n\n
func (e SSEEvent) FormatBytes() ([]byte, error) {
	dataBytes, err := json.Marshal(e.Data)
	if err != nil {
		return nil, fmt.Errorf("impossibile serializzare i dati dell'evento SSE: %w", err)
	}

	var buf bytes.Buffer
	if e.ID != "" {
		buf.WriteString("id: ")
		buf.WriteString(e.ID)
		buf.WriteByte('\n')
	}
	if e.Event != "" {
		buf.WriteString("event: ")
		buf.WriteString(e.Event)
		buf.WriteByte('\n')
	}
	if e.Retry > 0 {
		fmt.Fprintf(&buf, "retry: %d\n", e.Retry)
	}
	buf.WriteString("data: ")
	buf.Write(dataBytes)
	buf.WriteString("\n\n")

	return buf.Bytes(), nil
}

// --- Payload Tipizzati per LAN Discovery ---

// DeviceDiscoveredPayload viene emesso non appena un host attivo risponde al probe
type DeviceDiscoveredPayload struct {
	SessionID uuid.UUID        `json:"session_id"`
	Device    DiscoveredDevice `json:"device"`
}

// DiscoveryProgressPayload riporta la progressione percentuale e l'IP corrente scansionato
type DiscoveryProgressPayload struct {
	SessionID    uuid.UUID `json:"session_id"`
	ScannedHosts int       `json:"scanned_hosts"`
	TotalHosts   int       `json:"total_hosts"`
	Percentage   float64   `json:"percentage"`
	CurrentIP    string    `json:"current_ip"`
}

// DiscoveryCompletePayload viene emesso alla conclusione dell'intera ricognizione LAN
type DiscoveryCompletePayload struct {
	SessionID       uuid.UUID `json:"session_id"`
	TotalDiscovered int       `json:"total_discovered"`
	DurationSeconds float64   `json:"duration_seconds"`
}

// DiscoveryErrorPayload segnala un'anomalia o interruzione durante la scansione LAN
type DiscoveryErrorPayload struct {
	SessionID uuid.UUID `json:"session_id"`
	Message   string    `json:"message"`
}

// --- Payload Tipizzati per Vulnerability Scans ---

// ScanPhasePayload notifica la transizione tra gli stadi di scansione (HOST_DISCOVERY, PORT_SCAN, ecc.)
type ScanPhasePayload struct {
	ScanID     uuid.UUID `json:"scan_id"`
	Phase      string    `json:"phase"`
	Message    string    `json:"message"`
	Percentage float64   `json:"percentage"`
}

// ScanProgressPayload notifica l'avanzamento percentuale stimato dal processo Nmap (--stats-every)
type ScanProgressPayload struct {
	ObservedAt       time.Time `json:"observed_at"`
	Phase            string    `json:"phase,omitempty"`
	PhasePercentage  *float64  `json:"phase_percentage,omitempty"`
	RemainingSeconds *float64  `json:"remaining_seconds,omitempty"`
	ScanID           uuid.UUID `json:"scan_id"`
	Percentage       float64   `json:"percentage"`
	Message          string    `json:"message"`
}

// ScanFindingPayload notifica una porta aperta o vulnerabilità persistita incrementalmente su DB
type ScanFindingPayload struct {
	ScanID  uuid.UUID `json:"scan_id"`
	Finding Finding   `json:"finding"`
}

// ScanStatusPayload notifica il cambio di stato della scansione (RUNNING, COMPLETED, FAILED)
type ScanStatusPayload struct {
	ScanID       uuid.UUID        `json:"scan_id"`
	Status       ScanStatus       `json:"status"`
	SummaryStats ScanSummaryStats `json:"summary_stats"`
}

// ScanErrorPayload segnala un errore bloccante del job di scansione
type ScanErrorPayload struct {
	ScanID  uuid.UUID `json:"scan_id"`
	Message string    `json:"message"`
}

// --- Helper per Routing dei Topic SSE ---

// DiscoveryTopic genera la chiave di instradamento per gli eventi di una sessione di network discovery
func DiscoveryTopic(sessionID uuid.UUID) string {
	return "discovery:" + sessionID.String()
}

// ScanTopic genera la chiave di instradamento per gli eventi di un job di scansione
func ScanTopic(scanID uuid.UUID) string {
	return "scan:" + scanID.String()
}

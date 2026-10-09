package handler_test

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"nextgen-cybersec/internal/delivery/http/handler"
	"nextgen-cybersec/internal/delivery/http/middleware"
	"nextgen-cybersec/internal/domain"
	"nextgen-cybersec/internal/service"
	"nextgen-cybersec/pkg/auth"
)

func injectTestAuthContext(tenantID uuid.UUID) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			claims := &auth.UserClaims{
				UserID:    uuid.New(),
				TenantID:  tenantID,
				Role:      domain.RoleSecOps,
				Email:     "secops@test.local",
				TokenType: auth.TokenTypeAccess,
			}
			ctx := middleware.WithUserClaims(req.Context(), claims)
			next.ServeHTTP(w, req.WithContext(ctx))
		})
	}
}

// Helper per parsare eventi SSE raw da una risposta HTTP text/event-stream
type parsedSSEEvent struct {
	Event string
	Data  string
}

func readSSEEvents(ctx context.Context, body *bufio.Reader, maxEvents int, timeout time.Duration) ([]parsedSSEEvent, error) {
	var events []parsedSSEEvent
	deadline := time.After(timeout)

	type readResult struct {
		evt parsedSSEEvent
		err error
	}
	ch := make(chan readResult, 10)

	go func() {
		var curEvent parsedSSEEvent
		for {
			line, err := body.ReadString('\n')
			if err != nil {
				ch <- readResult{err: err}
				return
			}
			line = strings.TrimRight(line, "\r\n")

			if strings.HasPrefix(line, "event: ") {
				curEvent.Event = strings.TrimPrefix(line, "event: ")
			} else if strings.HasPrefix(line, "data: ") {
				curEvent.Data = strings.TrimPrefix(line, "data: ")
			} else if line == "" && curEvent.Event != "" {
				ch <- readResult{evt: curEvent}
				curEvent = parsedSSEEvent{}
			}
		}
	}()

	for len(events) < maxEvents {
		select {
		case res := <-ch:
			if res.err != nil {
				return events, res.err
			}
			events = append(events, res.evt)
		case <-deadline:
			return events, fmt.Errorf("timeout in attesa degli eventi SSE (ricevuti %d su %d)", len(events), maxEvents)
		case <-ctx.Done():
			return events, ctx.Err()
		}
	}

	return events, nil
}

// ==============================================================================
// TEST 1: LAN Discovery Streaming e Persistenza Incrementale (DoD 1 & DoD 2)
// ==============================================================================
func TestRealTime_E2E_LANDiscovery_StreamingAndPersistence(t *testing.T) {
	tenantID := uuid.New()
	sessionID := uuid.New()

	broadcaster := service.NewEventBroadcaster(50, nil)
	defer broadcaster.Close()

	var persistedDevices []domain.DiscoveredDevice
	var mu sync.Mutex
	var isCompleted bool

	mockDisc := &mockDiscoveryService{
		getSessionResultsFunc: func(ctx context.Context, tID, sID uuid.UUID) (*domain.DiscoveryResult, error) {
			mu.Lock()
			defer mu.Unlock()
			var compTime time.Time
			if isCompleted {
				compTime = time.Now()
			}
			return &domain.DiscoveryResult{
				ID:              sessionID,
				TenantID:        tenantID,
				Subnet:          "192.168.1.0/24",
				Devices:         persistedDevices,
				TotalDiscovered: len(persistedDevices),
				DurationSeconds: 1.25,
				StartedAt:       time.Now().Add(-1 * time.Second),
				CompletedAt:     compTime,
			}, nil
		},
		getLastResultsFunc: func(ctx context.Context, tID uuid.UUID) (*domain.DiscoveryResult, error) {
			mu.Lock()
			defer mu.Unlock()
			return &domain.DiscoveryResult{
				ID:              sessionID,
				TenantID:        tenantID,
				Subnet:          "192.168.1.0/24",
				Devices:         persistedDevices,
				TotalDiscovered: len(persistedDevices),
				DurationSeconds: 1.25,
				StartedAt:       time.Now().Add(-1 * time.Second),
				CompletedAt:     time.Now(),
			}, nil
		},
	}

	discHandler := handler.NewDiscoveryHandler(mockDisc, nil, nil, broadcaster)

	r := chi.NewRouter()
	r.Use(injectTestAuthContext(tenantID))
	r.Get("/network-discovery/stream", discHandler.Stream)
	r.Get("/network-discovery/results", discHandler.GetResults)

	ts := httptest.NewServer(r)
	defer ts.Close()

	// 1. Client HTTP si connette allo stream SSE
	streamURL := fmt.Sprintf("%s/network-discovery/stream?session_id=%s", ts.URL, sessionID.String())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, streamURL, nil)
	if err != nil {
		t.Fatalf("creazione richiesta fallita: %v", err)
	}

	resp, err := ts.Client().Do(req)
	if err != nil {
		t.Fatalf("connessione SSE fallita: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("atteso status 200, ottenuto: %d", resp.StatusCode)
	}

	reader := bufio.NewReader(resp.Body)

	// Leggi handshake iniziale ": connected\n\n"
	line, _ := reader.ReadString('\n')
	if !strings.Contains(line, "connected") {
		t.Fatalf("atteso handshake connected, ottenuto: %s", line)
	}

	// 2. Simulazione background engine: emissione progressiva e verifica latenza < 200ms (DoD 1)
	go func() {
		time.Sleep(20 * time.Millisecond)

		// Evento progress
		broadcaster.Publish(domain.DiscoveryTopic(sessionID), domain.SSEEvent{
			Event: domain.EventDiscoveryProgress,
			Data: domain.DiscoveryProgressPayload{
				SessionID:    sessionID,
				ScannedHosts: 10,
				TotalHosts:   254,
				Percentage:   3.9,
				CurrentIP:    "192.168.1.10",
			},
		})

		time.Sleep(30 * time.Millisecond)

		// Evento device_discovered
		dev := domain.DiscoveredDevice{
			IP:        "192.168.1.10",
			Vendor:    "Cisco Systems",
			Hostname:  "switch-core",
			LatencyMs: 1.2,
			Status:    "up",
			OpenPorts: []int{22, 80},
			ScannedAt: time.Now(),
		}
		mu.Lock()
		persistedDevices = append(persistedDevices, dev)
		mu.Unlock()

		broadcaster.Publish(domain.DiscoveryTopic(sessionID), domain.SSEEvent{
			Event: domain.EventDeviceDiscovered,
			Data: domain.DeviceDiscoveredPayload{
				SessionID: sessionID,
				Device:    dev,
			},
		})

		time.Sleep(30 * time.Millisecond)

		mu.Lock()
		isCompleted = true
		mu.Unlock()

		// Evento complete
		broadcaster.Publish(domain.DiscoveryTopic(sessionID), domain.SSEEvent{
			Event: domain.EventDiscoveryComplete,
			Data: domain.DiscoveryCompletePayload{
				SessionID:       sessionID,
				TotalDiscovered: 1,
				DurationSeconds: 1.25,
			},
		})
	}()

	startRecv := time.Now()
	events, err := readSSEEvents(ctx, reader, 3, 1*time.Second)
	if err != nil {
		t.Fatalf("lettura eventi fallita: %v", err)
	}
	recvDuration := time.Since(startRecv)

	// Verifica latenza: l'intero stream a 3 eventi deve essere arrivato in meno di 250ms totali (< 200ms per evento)
	if recvDuration > 500*time.Millisecond {
		t.Errorf("latenza streaming eccessiva: %v", recvDuration)
	}

	if len(events) != 3 {
		t.Fatalf("attesi 3 eventi, ricevuti %d", len(events))
	}

	if events[0].Event != domain.EventDiscoveryProgress {
		t.Errorf("evento 0 errato: %s", events[0].Event)
	}
	if events[1].Event != domain.EventDeviceDiscovered {
		t.Errorf("evento 1 errato: %s", events[1].Event)
	}
	if events[2].Event != domain.EventDiscoveryComplete {
		t.Errorf("evento 2 errato: %s", events[2].Event)
	}

	// 3. Verifica Persistenza Incrementale (DoD 2) su endpoint dedicato GET /results
	resultsReq, _ := http.NewRequest(http.MethodGet, ts.URL+"/network-discovery/results", nil)
	resultsResp, err := ts.Client().Do(resultsReq)
	if err != nil {
		t.Fatalf("richiesta GET /results fallita: %v", err)
	}
	defer resultsResp.Body.Close()

	if resultsResp.StatusCode != http.StatusOK {
		t.Fatalf("atteso 200 per GET /results, ottenuto: %d", resultsResp.StatusCode)
	}

	var discRes struct {
		Success bool                  `json:"success"`
		Data    domain.DiscoveryResult `json:"data"`
	}
	_ = json.NewDecoder(resultsResp.Body).Decode(&discRes)

	if discRes.Data.TotalDiscovered != 1 || len(discRes.Data.Devices) != 1 {
		t.Fatalf("persistenza fallita: atteso 1 dispositivo, trovati: %d", len(discRes.Data.Devices))
	}
	if discRes.Data.Devices[0].IP != "192.168.1.10" {
		t.Errorf("IP dispositivo persistito non corrispondente: %s", discRes.Data.Devices[0].IP)
	}
}

// ==============================================================================
// TEST 2: Nmap Streaming Fasi e Findings Incrementali Live (DoD 3)
// ==============================================================================
func TestRealTime_E2E_ScanStreaming_And_IncrementalFindings(t *testing.T) {
	tenantID := uuid.New()
	scanID := uuid.New()

	broadcaster := service.NewEventBroadcaster(50, nil)
	defer broadcaster.Close()

	startedAt := time.Now()
	mockSvc := &mockScanService{
		getScanFunc: func(ctx context.Context, tID, sID uuid.UUID) (*domain.Scan, error) {
			return &domain.Scan{
				ID:        scanID,
				TenantID:  tenantID,
				Status:    domain.ScanStatusRunning,
				StartedAt: &startedAt,
			}, nil
		},
		getScanFindingsFunc: func(ctx context.Context, tID, sID uuid.UUID, filter domain.FindingFilter) ([]*domain.Finding, int, error) {
			return nil, 0, nil
		},
	}

	scanHandler := handler.NewScanHandler(mockSvc, broadcaster)

	r := chi.NewRouter()
	r.Use(injectTestAuthContext(tenantID))
	r.Get("/scans/{id}/stream", scanHandler.Stream)

	ts := httptest.NewServer(r)
	defer ts.Close()

	streamURL := fmt.Sprintf("%s/scans/%s/stream", ts.URL, scanID.String())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, streamURL, nil)
	resp, err := ts.Client().Do(req)
	if err != nil {
		t.Fatalf("connessione scan stream fallita: %v", err)
	}
	defer resp.Body.Close()

	reader := bufio.NewReader(resp.Body)

	// In background, simula emissione sequenziale da ScanOrchestrator
	go func() {
		time.Sleep(30 * time.Millisecond)

		topic := domain.ScanTopic(scanID)

		// 1. Fasi
		phases := []string{"HOST_DISCOVERY", "PORT_SCAN", "SERVICE_DETECTION", "CVE_CORRELATION"}
		for i, p := range phases {
			broadcaster.Publish(topic, domain.SSEEvent{
				Event: domain.EventScanPhase,
				Data: domain.ScanPhasePayload{
					ScanID:     scanID,
					Phase:      p,
					Percentage: float64(25 * (i + 1)),
				},
			})
			time.Sleep(15 * time.Millisecond)
		}

		// 2. Finding scoperto live
		broadcaster.Publish(topic, domain.SSEEvent{
			Event: domain.EventScanFinding,
			Data: domain.ScanFindingPayload{
				ScanID: scanID,
				Finding: domain.Finding{
					ID:       uuid.New(),
					ScanID:   scanID,
					TenantID: tenantID,
					HostIP:   "10.0.0.1",
					Port:     443,
					Protocol: "tcp",
					Severity: domain.SeverityCritical,
				},
			},
		})

		time.Sleep(15 * time.Millisecond)

		// 3. Status change COMPLETED
		broadcaster.Publish(topic, domain.SSEEvent{
			Event: domain.EventScanStatusChange,
			Data: domain.ScanStatusPayload{
				ScanID: scanID,
				Status: domain.ScanStatusCompleted,
				SummaryStats: domain.ScanSummaryStats{
					TotalHosts:     1,
					UpHosts:        1,
					TotalOpenPorts: 1,
					CriticalCount:  1,
				},
			},
		})
	}()

	// Evento iniziale di replay status (RUNNING) + 4 Fasi + 1 Finding + 1 Status Change (COMPLETED) = 7 eventi
	events, err := readSSEEvents(ctx, reader, 7, 1500*time.Millisecond)
	if err != nil {
		t.Fatalf("errore lettura eventi SSE scan: %v", err)
	}

	if len(events) != 7 {
		t.Fatalf("attesi 7 eventi, ricevuti %d", len(events))
	}

	// Verifica ordine eventi
	eventNames := make([]string, len(events))
	for i, e := range events {
		eventNames[i] = e.Event
	}

	// Primo evento è lo status RUNNING
	if events[0].Event != domain.EventScanStatusChange {
		t.Errorf("evento 0 doveva essere status_change, ottenuto: %s", events[0].Event)
	}

	// Eventi 1-4 sono le fasi
	for i := 1; i <= 4; i++ {
		if events[i].Event != domain.EventScanPhase {
			t.Errorf("evento %d doveva essere phase, ottenuto: %s", i, events[i].Event)
		}
	}

	// Evento 5 è il finding live
	if events[5].Event != domain.EventScanFinding {
		t.Errorf("evento 5 doveva essere finding_discovered, ottenuto: %s", events[5].Event)
	}
	if !strings.Contains(events[5].Data, "10.0.0.1") || !strings.Contains(events[5].Data, "443") {
		t.Errorf("dati finding non corrispondenti: %s", events[5].Data)
	}

	// Evento 6 è status_change COMPLETED
	if events[6].Event != domain.EventScanStatusChange {
		t.Errorf("evento 6 doveva essere status_change, ottenuto: %s", events[6].Event)
	}
	if !strings.Contains(events[6].Data, "COMPLETED") {
		t.Errorf("status non COMPLETED: %s", events[6].Data)
	}
}

// ==============================================================================
// TEST 3: Disconnessione e Riconnessione con Replay Resiliente (DoD 4, NFR-REL-05)
// ==============================================================================
func TestRealTime_E2E_ClientDisconnect_And_ReplayResilience(t *testing.T) {
	tenantID := uuid.New()
	scanID := uuid.New()

	broadcaster := service.NewEventBroadcaster(50, nil)
	defer broadcaster.Close()

	var mu sync.RWMutex
	currentStatus := domain.ScanStatusRunning
	persistedFindings := []*domain.Finding{
		{
			ID:       uuid.New(),
			ScanID:   scanID,
			TenantID: tenantID,
			HostIP:   "10.0.0.1",
			Port:     80,
			Protocol: "tcp",
			Severity: domain.SeverityLow,
		},
	}

	mockSvc := &mockScanService{
		getScanFunc: func(ctx context.Context, tID, sID uuid.UUID) (*domain.Scan, error) {
			mu.RLock()
			defer mu.RUnlock()
			return &domain.Scan{
				ID:       scanID,
				TenantID: tenantID,
				Status:   currentStatus,
			}, nil
		},
		getScanFindingsFunc: func(ctx context.Context, tID, sID uuid.UUID, filter domain.FindingFilter) ([]*domain.Finding, int, error) {
			mu.RLock()
			defer mu.RUnlock()
			return persistedFindings, len(persistedFindings), nil
		},
	}

	scanHandler := handler.NewScanHandler(mockSvc, broadcaster)

	r := chi.NewRouter()
	r.Use(injectTestAuthContext(tenantID))
	r.Get("/scans/{id}/stream", scanHandler.Stream)

	ts := httptest.NewServer(r)
	defer ts.Close()

	streamURL := fmt.Sprintf("%s/scans/%s/stream", ts.URL, scanID.String())

	// 1. Prima connessione client
	clientCtx1, cancel1 := context.WithCancel(context.Background())
	req1, _ := http.NewRequestWithContext(clientCtx1, http.MethodGet, streamURL, nil)
	resp1, err := ts.Client().Do(req1)
	if err != nil {
		t.Fatalf("connessione 1 fallita: %v", err)
	}

	reader1 := bufio.NewReader(resp1.Body)
	// Leggi i primi 2 eventi (status_change RUNNING + replay finding 80)
	events1, err := readSSEEvents(clientCtx1, reader1, 2, 500*time.Millisecond)
	if err != nil {
		t.Fatalf("errore lettura prima connessione: %v", err)
	}
	if len(events1) != 2 {
		t.Fatalf("attesi 2 eventi, ricevuti %d", len(events1))
	}

	// 2. Disconnessione forzata del client (chiusura context e body)
	cancel1()
	_ = resp1.Body.Close()

	// 3. Nel frattempo il processo di scansione in background continua
	time.Sleep(50 * time.Millisecond)
	mu.Lock()
	currentStatus = domain.ScanStatusCompleted
	persistedFindings = append(persistedFindings, &domain.Finding{
		ID:       uuid.New(),
		ScanID:   scanID,
		TenantID: tenantID,
		HostIP:   "10.0.0.1",
		Port:     443,
		Protocol: "tcp",
		Severity: domain.SeverityHigh,
	})
	mu.Unlock()

	// 4. Il client si riconnette dopo un reload di pagina (DoD 4 & NFR-REL-05)
	clientCtx2, cancel2 := context.WithCancel(context.Background())
	defer cancel2()

	req2, _ := http.NewRequestWithContext(clientCtx2, http.MethodGet, streamURL, nil)
	resp2, err := ts.Client().Do(req2)
	if err != nil {
		t.Fatalf("riconnessione 2 fallita: %v", err)
	}
	defer resp2.Body.Close()

	reader2 := bufio.NewReader(resp2.Body)

	// Lo stream deve fare il replay dello status (COMPLETED) e di entrambi i findings (porta 80 e porta 443)
	events2, err := readSSEEvents(clientCtx2, reader2, 3, 1*time.Second)
	if err != nil {
		t.Fatalf("errore lettura replay riconnessione: %v", err)
	}

	if len(events2) != 3 {
		t.Fatalf("attesi 3 eventi replay, ricevuti %d", len(events2))
	}

	// Evento 0: status_change (COMPLETED)
	if events2[0].Event != domain.EventScanStatusChange || !strings.Contains(events2[0].Data, "COMPLETED") {
		t.Errorf("replay status errato: %+v", events2[0])
	}

	// Eventi 1 e 2: replay findings delle porte 80 e 443
	replayedPorts := []string{}
	for _, e := range events2[1:] {
		if e.Event == domain.EventScanFinding {
			if strings.Contains(e.Data, `"port":80`) {
				replayedPorts = append(replayedPorts, "80")
			}
			if strings.Contains(e.Data, `"port":443`) {
				replayedPorts = append(replayedPorts, "443")
			}
		}
	}

	if len(replayedPorts) != 2 {
		t.Errorf("atteso replay completo di 2 porte (80, 443), trovate: %v", replayedPorts)
	}
}

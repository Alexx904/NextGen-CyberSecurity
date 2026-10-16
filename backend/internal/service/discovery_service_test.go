package service

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"

	"nextgen-cybersec/internal/domain"
	"nextgen-cybersec/internal/scanner"
)

// mockScanRunner simula il runner Nmap restituendo XML prefabbricato
type mockScanRunner struct {
	runScanFn func(ctx context.Context, target string, profile scanner.ProfileConfig) (*scanner.ScanExecutionResult, error)
}

func (m *mockScanRunner) RunScan(ctx context.Context, target string, profile scanner.ProfileConfig) (*scanner.ScanExecutionResult, error) {
	if m.runScanFn != nil {
		return m.runScanFn(ctx, target, profile)
	}
	return nil, errors.New("runScanFn non configurata nel mock")
}

var sampleDiscoveryXML = []byte(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE nmaprun>
<nmaprun scanner="nmap" args="nmap -sn -PR -T4 192.168.1.0/24" start="1710000000" version="7.94">
  <host starttime="1710000000" endtime="1710000001">
    <status state="up" reason="arp-response" reason_ttl="0"/>
    <address addr="192.168.1.10" addrtype="ipv4"/>
    <address addr="B8:27:EB:12:34:56" addrtype="mac"/>
    <hostnames>
      <hostname name="raspi-server.local" type="PTR"/>
    </hostnames>
    <times srtt="1250" rttvar="500" to="100000"/>
  </host>
  <host starttime="1710000001" endtime="1710000002">
    <status state="up" reason="arp-response" reason_ttl="0"/>
    <address addr="192.168.1.20" addrtype="ipv4"/>
    <address addr="00:14:22:AA:BB:CC" addrtype="mac"/>
    <times srtt="800" rttvar="300" to="100000"/>
  </host>
  <host starttime="1710000002" endtime="1710000003">
    <status state="down" reason="no-response"/>
    <address addr="192.168.1.50" addrtype="ipv4"/>
  </host>
  <runstats>
    <finished time="1710000005" elapsed="2.50" summary="Nmap done: 3 hosts scanned, 2 up" exit="success"/>
    <hosts up="2" down="1" total="3"/>
  </runstats>
</nmaprun>`)

func TestDiscoveryService_RespectsCommonPortsOption(t *testing.T) {
	for _, enabled := range []bool{false, true} {
		t.Run(fmt.Sprintf("include_ports_%v", enabled), func(t *testing.T) {
			var enrichments atomic.Int32
			runner := &mockScanRunner{runScanFn: func(ctx context.Context, target string, profile scanner.ProfileConfig) (*scanner.ScanExecutionResult, error) {
				if profile.Name == "LAN Host Enrichment" {
					enrichments.Add(1)
				}
				return &scanner.ScanExecutionResult{RawXML: sampleDiscoveryXML}, nil
			}}
			svc := NewDiscoveryService(runner, nil, nil, nil)
			_, err := svc.DiscoverSubnet(context.Background(), uuid.New(), domain.DiscoveryRequest{Subnet: "192.168.1.0/24", IncludeCommonPorts: enabled})
			if err != nil {
				t.Fatal(err)
			}
			want := int32(0)
			if enabled {
				want = 2
			}
			if got := enrichments.Load(); got != want {
				t.Fatalf("port scans = %d, want %d", got, want)
			}
		})
	}
}

func TestDiscoveryService_DiscoverSubnet_Success(t *testing.T) {
	runner := &mockScanRunner{
		runScanFn: func(ctx context.Context, target string, profile scanner.ProfileConfig) (*scanner.ScanExecutionResult, error) {
			if target == "192.168.1.0/24" {
				return &scanner.ScanExecutionResult{
					RawXML:   sampleDiscoveryXML,
					ExitCode: 0,
					Duration: 2500 * time.Millisecond,
				}, nil
			}
			// Stage 2 targeted host enrichment
			return &scanner.ScanExecutionResult{
				RawXML:   sampleDiscoveryXML,
				ExitCode: 0,
				Duration: 500 * time.Millisecond,
			}, nil
		},
	}

	ouiDB := scanner.NewOUIDatabase()
	svc := NewDiscoveryService(runner, ouiDB, nil, nil)

	tenantID := uuid.New()
	req := domain.DiscoveryRequest{
		Subnet: "192.168.1.0/24",
	}

	res, err := svc.DiscoverSubnet(context.Background(), tenantID, req)
	if err != nil {
		t.Fatalf("DiscoverSubnet fallito inaspettatamente: %v", err)
	}

	if res == nil {
		t.Fatal("risultato nullo inatteso")
	}

	if res.TenantID != tenantID {
		t.Errorf("atteso tenantID %s, ricevuto %s", tenantID, res.TenantID)
	}

	if res.TotalDiscovered != 2 {
		t.Errorf("attesi 2 host scoperti (escluso host down), ricevuti %d", res.TotalDiscovered)
	}

	// Verifica primo dispositivo (Raspberry Pi con hostname e latenza)
	dev1 := res.Devices[0]
	if dev1.IP != "192.168.1.10" {
		t.Errorf("dev1: atteso IP 192.168.1.10, ricevuto '%s'", dev1.IP)
	}
	if dev1.MAC != "B8:27:EB:12:34:56" {
		t.Errorf("dev1: atteso MAC B8:27:EB:12:34:56, ricevuto '%s'", dev1.MAC)
	}
	if dev1.Vendor != "Raspberry Pi Foundation" {
		t.Errorf("dev1: atteso Vendor 'Raspberry Pi Foundation', ricevuto '%s'", dev1.Vendor)
	}
	if dev1.Hostname != "raspi-server.local" {
		t.Errorf("dev1: atteso Hostname 'raspi-server.local', ricevuto '%s'", dev1.Hostname)
	}
	if dev1.LatencyMs != 1.25 {
		t.Errorf("dev1: attesa LatencyMs 1.25, ricevuta %f", dev1.LatencyMs)
	}
	if dev1.Status != "up" {
		t.Errorf("dev1: atteso Status 'up', ricevuto '%s'", dev1.Status)
	}

	// Verifica secondo dispositivo (Dell senza hostname esplicito in XML)
	dev2 := res.Devices[1]
	if dev2.IP != "192.168.1.20" {
		t.Errorf("dev2: atteso IP 192.168.1.20, ricevuto '%s'", dev2.IP)
	}
	if dev2.Vendor != "Dell" {
		t.Errorf("dev2: atteso Vendor 'Dell', ricevuto '%s'", dev2.Vendor)
	}
	if dev2.LatencyMs != 0.8 {
		t.Errorf("dev2: attesa LatencyMs 0.8, ricevuta %f", dev2.LatencyMs)
	}

	// Verifica cache tramite GetLastResults
	cached, err := svc.GetLastResults(context.Background(), tenantID)
	if err != nil {
		t.Fatalf("GetLastResults fallito inaspettatamente: %v", err)
	}
	if cached.ID != res.ID {
		t.Errorf("ID cache non corrispondente: atteso %s, ricevuto %s", res.ID, cached.ID)
	}
}

func TestDiscoveryService_DiscoverSubnet_ValidationErrors(t *testing.T) {
	runner := &mockScanRunner{}
	svc := NewDiscoveryService(runner, nil, nil, nil)
	validTenant := uuid.New()

	tests := []struct {
		name     string
		tenantID uuid.UUID
		subnet   string
		errCheck func(error) bool
	}{
		{
			name:     "Nil tenant ID",
			tenantID: uuid.Nil,
			subnet:   "192.168.1.0/24",
			errCheck: func(err error) bool { return errors.Is(err, domain.ErrValidation) },
		},
		{
			name:     "Invalid target format",
			tenantID: validTenant,
			subnet:   "invalid---range",
			errCheck: func(err error) bool { return errors.Is(err, domain.ErrTargetInvalid) },
		},
		{
			name:     "FQDN target rejected for LAN discovery",
			tenantID: validTenant,
			subnet:   "scanme.nmap.org",
			errCheck: func(err error) bool { return errors.Is(err, domain.ErrTargetInvalid) },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.DiscoverSubnet(context.Background(), tt.tenantID, domain.DiscoveryRequest{Subnet: tt.subnet})
			if err == nil {
				t.Fatal("atteso errore di validazione, ricevuto nil")
			}
			if !tt.errCheck(err) {
				t.Fatalf("tipo di errore inatteso: %v", err)
			}
		})
	}
}

func TestDiscoveryService_GetLastResults_NotFound(t *testing.T) {
	runner := &mockScanRunner{}
	svc := NewDiscoveryService(runner, nil, nil, nil)

	_, err := svc.GetLastResults(context.Background(), uuid.New())
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("atteso domain.ErrNotFound per tenant non presente in cache, ricevuto: %v", err)
	}
}

func TestDiscoveryService_LookupVendor(t *testing.T) {
	runner := &mockScanRunner{}
	svc := NewDiscoveryService(runner, nil, nil, nil)

	if vendor := svc.LookupVendor("00:50:56:11:22:33"); vendor != "VMware" {
		t.Errorf("atteso VMware, ricevuto '%s'", vendor)
	}
	if vendor := svc.LookupVendor("invalid"); vendor != "Unknown" {
		t.Errorf("atteso Unknown, ricevuto '%s'", vendor)
	}
}

func TestDiscoveryService_ResolveHostnamesConcurrently(t *testing.T) {
	runner := &mockScanRunner{}
	svcObj := NewDiscoveryService(runner, nil, nil, nil)
	svc, ok := svcObj.(*discoveryService)
	if !ok {
		t.Fatal("tipo di discoveryService non atteso")
	}

	hosts := []scanner.ParsedHost{
		{IP: "127.0.0.1", PrimaryHostname: ""},
		{IP: "127.0.0.2", PrimaryHostname: "pre-resolved.local"},
		{IP: "192.0.2.1", PrimaryHostname: ""}, // TEST-NET-1 (non risolvibile, andrà in timeout veloce)
		{IP: "192.0.2.2", PrimaryHostname: ""},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	start := time.Now()
	res := svc.resolveHostnamesConcurrently(ctx, hosts, 4)
	elapsed := time.Since(start)

	if elapsed > 3*time.Second {
		t.Errorf("risoluzione concorrente ha superato il timeout di sicurezza: %v", elapsed)
	}

	if res["127.0.0.2"] != "pre-resolved.local" {
		t.Errorf("atteso pre-resolved.local, ottenuto '%s'", res["127.0.0.2"])
	}

	// 127.0.0.1 su molti sistemi risolve in localhost
	t.Logf("Risoluzione 127.0.0.1: '%s' (completata in %v)", res["127.0.0.1"], elapsed)
}

func TestDiscoveryService_FilterNetworkAndBroadcast(t *testing.T) {
	xmlWithNetworkAndBroadcast := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE nmaprun>
<nmaprun scanner="nmap" args="nmap -sn 192.168.1.0/24" start="1710000000" version="7.94">
  <host starttime="1710000000" endtime="1710000001">
    <status state="up" reason="echo-reply"/>
    <address addr="192.168.1.0" addrtype="ipv4"/>
  </host>
  <host starttime="1710000001" endtime="1710000002">
    <status state="up" reason="echo-reply"/>
    <address addr="192.168.1.1" addrtype="ipv4"/>
    <address addr="98:BA:5F:95:4E:50" addrtype="mac"/>
  </host>
  <host starttime="1710000002" endtime="1710000003">
    <status state="up" reason="echo-reply"/>
    <address addr="192.168.1.255" addrtype="ipv4"/>
  </host>
  <runstats>
    <finished time="1710000004" elapsed="1.50" summary="done" exit="success"/>
    <hosts up="3" down="0" total="3"/>
  </runstats>
</nmaprun>`)

	runner := &mockScanRunner{
		runScanFn: func(ctx context.Context, target string, profile scanner.ProfileConfig) (*scanner.ScanExecutionResult, error) {
			return &scanner.ScanExecutionResult{
				RawXML:   xmlWithNetworkAndBroadcast,
				ExitCode: 0,
				Duration: 1500 * time.Millisecond,
			}, nil
		},
	}

	ouiDB := scanner.NewOUIDatabase()
	svc := NewDiscoveryService(runner, ouiDB, nil, nil)

	tenantID := uuid.New()
	res, err := svc.DiscoverSubnet(context.Background(), tenantID, domain.DiscoveryRequest{
		Subnet: "192.168.1.0/24",
	})
	if err != nil {
		t.Fatalf("DiscoverSubnet fallito: %v", err)
	}

	// Devono essere scartati sia 192.168.1.0 (network) sia 192.168.1.255 (broadcast)
	if res.TotalDiscovered != 1 {
		t.Fatalf("atteso 1 solo host attivo (esclusi .0 e .255), ottenuti %d", res.TotalDiscovered)
	}
	if res.Devices[0].IP != "192.168.1.1" {
		t.Errorf("atteso IP 192.168.1.1, ottenuto '%s'", res.Devices[0].IP)
	}
}

func TestDiscoveryService_HeuristicVendorFallback(t *testing.T) {
	xmlWithServiceAndNoMAC := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE nmaprun>
<nmaprun scanner="nmap" args="nmap -sn 10.0.0.1" start="1710000000" version="7.94">
  <host starttime="1710000000" endtime="1710000001">
    <status state="up" reason="syn-ack"/>
    <address addr="10.0.0.1" addrtype="ipv4"/>
    <ports>
      <port protocol="tcp" portid="80">
        <state state="open"/>
        <service name="http" product="TP-LINK Web Server" method="probed"/>
      </port>
    </ports>
  </host>
  <runstats>
    <finished time="1710000002" elapsed="1.0" summary="done" exit="success"/>
    <hosts up="1" down="0" total="1"/>
  </runstats>
</nmaprun>`)

	runner := &mockScanRunner{
		runScanFn: func(ctx context.Context, target string, profile scanner.ProfileConfig) (*scanner.ScanExecutionResult, error) {
			return &scanner.ScanExecutionResult{
				RawXML:   xmlWithServiceAndNoMAC,
				ExitCode: 0,
				Duration: 1000 * time.Millisecond,
			}, nil
		},
	}

	svc := NewDiscoveryService(runner, nil, nil, nil)
	tenantID := uuid.New()
	res, err := svc.DiscoverSubnet(context.Background(), tenantID, domain.DiscoveryRequest{
		Subnet: "10.0.0.1",
	})
	if err != nil {
		t.Fatalf("DiscoverSubnet fallito: %v", err)
	}

	if len(res.Devices) != 1 {
		t.Fatalf("atteso 1 dispositivo, ottenuti %d", len(res.Devices))
	}
	dev := res.Devices[0]
	if dev.Vendor != "TP-Link" {
		t.Errorf("atteso Vendor euristico 'TP-Link', ottenuto '%s'", dev.Vendor)
	}
	if len(dev.OpenPorts) != 1 || dev.OpenPorts[0] != 80 {
		t.Errorf("attesa porta 80 aperta, ottenute %v", dev.OpenPorts)
	}
}

func TestDiscoveryService_LiveStreamingEvents(t *testing.T) {
	runner := &mockScanRunner{
		runScanFn: func(ctx context.Context, target string, profile scanner.ProfileConfig) (*scanner.ScanExecutionResult, error) {
			return &scanner.ScanExecutionResult{
				RawXML:   sampleDiscoveryXML,
				ExitCode: 0,
				Duration: 500 * time.Millisecond,
			}, nil
		},
	}

	broadcaster := NewEventBroadcaster(32, nil)
	defer broadcaster.Close()

	ouiDB := scanner.NewOUIDatabase()
	svc := NewDiscoveryService(runner, ouiDB, nil, broadcaster)

	sessionID := uuid.New()
	topic := domain.DiscoveryTopic(sessionID)
	ch, unsub := broadcaster.Subscribe(context.Background(), topic)
	defer unsub()

	tenantID := uuid.New()
	req := domain.DiscoveryRequest{
		Subnet:    "192.168.1.0/24",
		SessionID: &sessionID,
	}

	res, err := svc.DiscoverSubnet(context.Background(), tenantID, req)
	if err != nil {
		t.Fatalf("DiscoverSubnet fallito: %v", err)
	}
	if res.TotalDiscovered != 2 {
		t.Fatalf("attesi 2 dispositivi, trovati %d", res.TotalDiscovered)
	}

	// Verifica ricezione eventi SSE dal canale
	receivedEvents := make([]domain.SSEEvent, 0)
collectLoop:
	for {
		select {
		case evt, ok := <-ch:
			if !ok {
				break collectLoop
			}
			receivedEvents = append(receivedEvents, evt)
			if evt.Event == domain.EventDiscoveryComplete {
				break collectLoop
			}
		case <-time.After(500 * time.Millisecond):
			t.Fatalf("timeout in attesa eventi SSE, ricevuti finora: %d", len(receivedEvents))
		}
	}

	// Verifichiamo la presenza di progress, device_discovered e complete
	hasStartProgress := false
	hasMidProgress := false
	deviceDiscoveredCount := 0
	hasComplete := false

	for _, evt := range receivedEvents {
		switch evt.Event {
		case domain.EventDiscoveryProgress:
			payload, ok := evt.Data.(domain.DiscoveryProgressPayload)
			if ok && payload.Percentage == 0 {
				hasStartProgress = true
			}
			if ok && payload.Percentage == 50 {
				hasMidProgress = true
			}
		case domain.EventDeviceDiscovered:
			deviceDiscoveredCount++
		case domain.EventDiscoveryComplete:
			hasComplete = true
		}
	}

	if !hasStartProgress {
		t.Error("mancato evento progresso iniziale 0%")
	}
	if !hasMidProgress {
		t.Error("mancato evento progresso intermedio 50%")
	}
	if deviceDiscoveredCount != 2 {
		t.Errorf("attesi 2 eventi device_discovered, ricevuti %d", deviceDiscoveredCount)
	}
	if !hasComplete {
		t.Error("mancato evento finale discovery complete")
	}
}

func TestDiscoveryService_ErrorEventEmission(t *testing.T) {
	runner := &mockScanRunner{
		runScanFn: func(ctx context.Context, target string, profile scanner.ProfileConfig) (*scanner.ScanExecutionResult, error) {
			return nil, errors.New("nmap network unreachable")
		},
	}

	broadcaster := NewEventBroadcaster(10, nil)
	defer broadcaster.Close()

	svc := NewDiscoveryService(runner, nil, nil, broadcaster)

	sessionID := uuid.New()
	topic := domain.DiscoveryTopic(sessionID)
	ch, unsub := broadcaster.Subscribe(context.Background(), topic)
	defer unsub()

	tenantID := uuid.New()
	_, err := svc.DiscoverSubnet(context.Background(), tenantID, domain.DiscoveryRequest{
		Subnet:    "192.168.1.0/24",
		SessionID: &sessionID,
	})
	if err == nil {
		t.Fatal("atteso errore di scansione, ottenuto nil")
	}

	// Verifica ricezione evento di errore via SSE
	select {
	case evt := <-ch:
		// Il primo evento è progress 0%
		if evt.Event != domain.EventDiscoveryProgress {
			t.Fatalf("atteso evento iniziale progress, ottenuto %s", evt.Event)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timeout in attesa evento progresso iniziale")
	}

	select {
	case evt := <-ch:
		if evt.Event != domain.EventDiscoveryError {
			t.Fatalf("atteso evento error, ottenuto %s", evt.Event)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timeout in attesa evento discovery error")
	}
}

func TestDiscoveryService_GetSessionResults(t *testing.T) {
	runner := &mockScanRunner{
		runScanFn: func(ctx context.Context, target string, profile scanner.ProfileConfig) (*scanner.ScanExecutionResult, error) {
			return &scanner.ScanExecutionResult{
				RawXML:   sampleDiscoveryXML,
				ExitCode: 0,
			}, nil
		},
	}

	svc := NewDiscoveryService(runner, nil, nil, nil)
	tenantID := uuid.New()
	sessionID := uuid.New()

	// Sessione non esistente
	_, err := svc.GetSessionResults(context.Background(), tenantID, sessionID)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("atteso ErrNotFound per sessione inesistente, ottenuto: %v", err)
	}

	// Parametri non validi
	_, err = svc.GetSessionResults(context.Background(), uuid.Nil, sessionID)
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("atteso ErrValidation per tenant Nil, ottenuto: %v", err)
	}

	// Esegui scansione con sessionID
	_, err = svc.DiscoverSubnet(context.Background(), tenantID, domain.DiscoveryRequest{
		Subnet:    "192.168.1.0/24",
		SessionID: &sessionID,
	})
	if err != nil {
		t.Fatalf("DiscoverSubnet fallito: %v", err)
	}

	// Recupero sessione esistente
	res, err := svc.GetSessionResults(context.Background(), tenantID, sessionID)
	if err != nil {
		t.Fatalf("GetSessionResults fallito per sessione esistente: %v", err)
	}
	if res.ID != sessionID {
		t.Errorf("atteso ID sessione %s, ottenuto %s", sessionID, res.ID)
	}
	if len(res.Devices) != 2 {
		t.Errorf("attesi 2 dispositivi salvati nella sessione, trovati %d", len(res.Devices))
	}
}

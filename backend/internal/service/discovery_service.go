package service

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"nextgen-cybersec/internal/domain"
	"nextgen-cybersec/internal/scanner"
	"nextgen-cybersec/pkg/validator"
)

type cachedHostInfo struct {
	Hostname string
	Vendor   string
	MAC      string
}

type discoveryService struct {
	runner      scanner.Runner
	ouiDB       scanner.OUIDatabase
	arpCache    scanner.ARPCache
	logger      *slog.Logger
	broadcaster EventBroadcaster

	mu             sync.RWMutex
	results        map[uuid.UUID]*domain.DiscoveryResult
	sessionResults map[uuid.UUID]*domain.DiscoveryResult
	hostCache      map[string]cachedHostInfo
}

// NewDiscoveryService inizializza il servizio per la ricognizione veloce della rete locale
func NewDiscoveryService(runner scanner.Runner, ouiDB scanner.OUIDatabase, logger *slog.Logger, broadcaster EventBroadcaster) domain.DiscoveryService {
	if logger == nil {
		logger = slog.Default()
	}
	if ouiDB == nil {
		ouiDB = scanner.NewOUIDatabase()
	}

	return &discoveryService{
		runner:         runner,
		ouiDB:          ouiDB,
		arpCache:       scanner.NewARPCache(),
		logger:         logger,
		broadcaster:    broadcaster,
		results:        make(map[uuid.UUID]*domain.DiscoveryResult),
		sessionResults: make(map[uuid.UUID]*domain.DiscoveryResult),
		hostCache:      make(map[string]cachedHostInfo),
	}
}

// DiscoverSubnet avvia la scansione rapida della subnet (Ping sweep sicuro a 2 stadi) e arricchisce i risultati con OUI, porte, hostname e banner
func (s *discoveryService) DiscoverSubnet(ctx context.Context, tenantID uuid.UUID, req domain.DiscoveryRequest) (*domain.DiscoveryResult, error) {
	// 1. Controllo validità tenant (NFR-SEC-03)
	if tenantID == uuid.Nil {
		return nil, fmt.Errorf("%w: tenant ID non valido", domain.ErrValidation)
	}

	// 2. Validazione sintattica e di sicurezza del bersaglio/subnet (NFR-SEC-01)
	cleanSubnet := strings.TrimSpace(req.Subnet)
	targetInfo, err := validator.ValidateTarget(cleanSubnet)
	if err != nil {
		return nil, fmt.Errorf("%w: subnet '%s' non valida", domain.ErrTargetInvalid, cleanSubnet)
	}

	// 2.1 Verifica perimetro di sicurezza interno e deny-list (FR-01.4)
	if err := validator.CheckTargetScope(cleanSubnet, false); err != nil {
		return nil, fmt.Errorf("perimetro subnet non consentito: %w", err)
	}

	// Solo IP, CIDR o Range sono ammessi per la LAN discovery
	if targetInfo.Type != domain.TargetTypeIP && targetInfo.Type != domain.TargetTypeCIDR && targetInfo.Type != domain.TargetTypeRange {
		return nil, fmt.Errorf("%w: tipologia '%s' non supportata per LAN discovery (richiesto IP, CIDR o Range)", domain.ErrTargetInvalid, targetInfo.Type)
	}

	// 3. Selezione profilo Stage 1: Fast Ping Sweep affidabile (ICMP + TCP SYN + UDP, senza TCP ACK)
	defaultProfiles := scanner.GetDefaultProfiles()
	profile, ok := defaultProfiles[scanner.ProfileLANDiscovery]
	if !ok {
		profile = scanner.ProfileConfig{
			Name:        string(scanner.ProfileLANDiscovery),
			Args:        []string{"-sn", "-PE", "-PS21,22,80,443,445,8080,8443,3389,53,8000", "-PU53,137,161,5353", "-T4", "-n"},
			Timeout:     2 * time.Minute,
			Description: "Ricognizione LAN ultra-veloce con ping sweep affidabile e probe ICMP/SYN/UDP su porte standard",
		}
	}

	sessionID := uuid.New()
	if req.SessionID != nil && *req.SessionID != uuid.Nil {
		sessionID = *req.SessionID
	}
	topic := domain.DiscoveryTopic(sessionID)

	s.logger.Info("avvio scansione LAN discovery (Stage 1 Sweep)",
		slog.String("tenant_id", tenantID.String()),
		slog.String("session_id", sessionID.String()),
		slog.String("subnet", targetInfo.CleanValue),
		slog.Bool("include_ports", req.IncludeCommonPorts),
	)

	startTime := time.Now()

	// Inizializza il record della sessione per persistenza incrementale
	s.mu.Lock()
	s.sessionResults[sessionID] = &domain.DiscoveryResult{
		ID:              sessionID,
		TenantID:        tenantID,
		Subnet:          targetInfo.CleanValue,
		Devices:         make([]domain.DiscoveredDevice, 0),
		TotalDiscovered: 0,
		StartedAt:       startTime,
	}
	s.mu.Unlock()

	// Notifica avvio scansione (0%)
	s.publishEvent(topic, domain.SSEEvent{
		Event: domain.EventDiscoveryProgress,
		Data: domain.DiscoveryProgressPayload{
			SessionID:    sessionID,
			ScannedHosts: 0,
			TotalHosts:   254,
			Percentage:   0.0,
			CurrentIP:    "",
		},
	})

	// 4. Esecuzione Nmap Stage 1
	execResult, err := s.runner.RunScan(ctx, targetInfo.CleanValue, profile)
	if err != nil {
		s.logger.Error("esecuzione scansione LAN discovery fallita",
			slog.String("subnet", targetInfo.CleanValue),
			slog.String("error", err.Error()),
		)
		s.publishEvent(topic, domain.SSEEvent{
			Event: domain.EventDiscoveryError,
			Data: domain.DiscoveryErrorPayload{
				SessionID: sessionID,
				Message:   err.Error(),
			},
		})
		return nil, fmt.Errorf("%w: %v", domain.ErrScannerExecFailed, err)
	}

	// 5. Parsing dell'output XML in streaming
	report, err := scanner.ParseNmapXMLBytes(execResult.RawXML)
	if err != nil {
		s.logger.Error("parsing output XML scansione LAN fallito",
			slog.String("subnet", targetInfo.CleanValue),
			slog.String("error", err.Error()),
		)
		s.publishEvent(topic, domain.SSEEvent{
			Event: domain.EventDiscoveryError,
			Data: domain.DiscoveryErrorPayload{
				SessionID: sessionID,
				Message:   err.Error(),
			},
		})
		return nil, fmt.Errorf("errore parsing risultati scansione LAN: %w", err)
	}

	// 6. Filtraggio rigoroso: solo host attivi ('up'), escludendo network e broadcast IP
	rawUpHosts := report.GetUpHosts()
	upHosts := make([]scanner.ParsedHost, 0, len(rawUpHosts))
	for _, h := range rawUpHosts {
		if isNetworkOrBroadcast(h.IP, targetInfo.CleanValue) {
			continue
		}
		upHosts = append(upHosts, h)
	}

	// Notifica avanzamento al termine dello Stage 1 (50%)
	s.publishEvent(topic, domain.SSEEvent{
		Event: domain.EventDiscoveryProgress,
		Data: domain.DiscoveryProgressPayload{
			SessionID:    sessionID,
			ScannedHosts: len(upHosts),
			TotalHosts:   len(rawUpHosts),
			Percentage:   50.0,
			CurrentIP:    "",
		},
	})

	// 7. Aggiorna tabella ARP locale
	_ = s.arpCache.Refresh()

	// 8. Stage 2: Arricchimento mirato porte e servizi per gli host attivi trovati
	enrichedHostsMap := make(map[string]scanner.ParsedHost)
	if len(upHosts) > 0 && len(upHosts) <= 64 {
		targetedProfile := scanner.ProfileConfig{
			Name:        "LAN Host Enrichment",
			Args:        []string{"-T4", "-F", "-n"},
			Timeout:     15 * time.Second,
			Description: "Scansione rapida porte comuni per host attivi",
		}

		var enrichWg sync.WaitGroup
		var enrichMu sync.Mutex
		sem := make(chan struct{}, 8) // Massimo 8 scansioni concorrenti

		for _, h := range upHosts {
			ip := h.IP
			enrichWg.Add(1)
			go func(targetIP string) {
				defer enrichWg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()

				itemCtx, itemCancel := context.WithTimeout(ctx, 20*time.Second)
				defer itemCancel()

				enrichExec, enrichErr := s.runner.RunScan(itemCtx, targetIP, targetedProfile)
				if enrichErr == nil && enrichExec != nil && len(enrichExec.RawXML) > 0 {
					enrichReport, pErr := scanner.ParseNmapXMLBytes(enrichExec.RawXML)
					if pErr == nil && enrichReport != nil {
						for _, eh := range enrichReport.GetUpHosts() {
							if eh.IP == targetIP {
								enrichMu.Lock()
								enrichedHostsMap[targetIP] = eh
								enrichMu.Unlock()
								break
							}
						}
					}
				}
			}(ip)
		}
		enrichWg.Wait()
	}

	// 9. Risoluzione concorrente hostname DNS
	resolvedHostnames := s.resolveHostnamesConcurrently(ctx, upHosts, 25)

	// 10. Assemblaggio finale dispositivi con arricchimento HTTP e Vendor Heuristic
	devices := make([]domain.DiscoveredDevice, 0, len(upHosts))
	scanTimestamp := report.EndTime
	if scanTimestamp.IsZero() {
		scanTimestamp = time.Now()
	}

	for idx, host := range upHosts {
		// Se disponibile, fondi le informazioni arricchite dello Stage 2
		currentHost := host
		if eh, exists := enrichedHostsMap[host.IP]; exists {
			if len(eh.Ports) > 0 {
				currentHost.Ports = eh.Ports
			}
			if eh.OS != nil && currentHost.OS == nil {
				currentHost.OS = eh.OS
			}
			if eh.PrimaryHostname != "" && currentHost.PrimaryHostname == "" {
				currentHost.PrimaryHostname = eh.PrimaryHostname
			}
			if eh.Vendor != "" && currentHost.Vendor == "" {
				currentHost.Vendor = eh.Vendor
			}
			if eh.MAC != "" && currentHost.MAC == "" {
				currentHost.MAC = eh.MAC
			}
		}

		// Risoluzione MAC: controlla Nmap L2 o fallback su ARPCache
		mac := currentHost.MAC
		if mac == "" {
			if cachedMAC := s.arpCache.Lookup(currentHost.IP); cachedMAC != "" {
				mac = cachedMAC
			}
		}

		// Estrazione porte aperte
		var openPorts []int
		var serviceNames []string
		var rawBanners []string
		for _, p := range currentHost.Ports {
			if strings.EqualFold(p.State, "open") {
				openPorts = append(openPorts, p.PortID)
				if p.Service != nil {
					if p.Service.Name != "" {
						serviceNames = append(serviceNames, p.Service.Name)
					}
					if p.Service.RawBanner != "" {
						rawBanners = append(rawBanners, p.Service.RawBanner)
					}
				}
			}
		}

		// Ispezione porte Web per estrazione titolo HTML e server header
		httpTitle, serverHeader := scanner.GrabHTTPInfo(ctx, currentHost.IP, openPorts)

		// Risoluzione Vendor:
		// a) Da MAC Address tramite IEEE OUI Database
		vendor := "Unknown"
		if mac != "" {
			ouiVendor := s.ouiDB.Lookup(mac)
			if ouiVendor != "Unknown" && ouiVendor != "" {
				vendor = ouiVendor
			}
		}

		// b) Fallback euristico se il MAC non è noto o non ha prodotto vendor
		if vendor == "Unknown" || vendor == "" {
			osName := ""
			if currentHost.OS != nil {
				osName = currentHost.OS.Name
			}
			hint := scanner.DeviceHint{
				Hostname:     currentHost.PrimaryHostname,
				OSName:       osName,
				Services:     serviceNames,
				RawBanners:   rawBanners,
				HTTPTitle:    httpTitle,
				ServerHeader: serverHeader,
			}
			inferred := scanner.InferVendor(hint)
			if inferred != "Unknown" {
				vendor = inferred
			} else if currentHost.Vendor != "" {
				vendor = currentHost.Vendor
			}
		}

		// Risoluzione Hostname:
		baseHostname := currentHost.PrimaryHostname
		if baseHostname == "" {
			baseHostname = resolvedHostnames[currentHost.IP]
		}
		hostname := scanner.CleanDeviceHostname(baseHostname, httpTitle, vendor)

		// Verifica e arricchimento da cache cronologica (NFR-REL: consistenza tra scansioni successive dello stesso IP)
		hostKey := fmt.Sprintf("%s:%s", tenantID.String(), currentHost.IP)
		s.mu.RLock()
		cached, hasCached := s.hostCache[hostKey]
		s.mu.RUnlock()

		if hasCached {
			if (mac == "" || mac == "Unknown") && cached.MAC != "" {
				mac = cached.MAC
			}
			if (vendor == "" || vendor == "Unknown") && cached.Vendor != "" && cached.Vendor != "Unknown" {
				vendor = cached.Vendor
			}
			if (hostname == "" || hostname == "Dispositivo LAN") && cached.Hostname != "" && cached.Hostname != "Dispositivo LAN" {
				hostname = cached.Hostname
			}
		}

		// Aggiorna cache con i dati più recenti e dettagliati
		if (hostname != "" && hostname != "Dispositivo LAN") || (vendor != "" && vendor != "Unknown") || mac != "" {
			s.mu.Lock()
			currentCache := s.hostCache[hostKey]
			if hostname != "" && hostname != "Dispositivo LAN" {
				currentCache.Hostname = hostname
			}
			if vendor != "" && vendor != "Unknown" {
				currentCache.Vendor = vendor
			}
			if mac != "" {
				currentCache.MAC = mac
			}
			s.hostCache[hostKey] = currentCache
			s.mu.Unlock()
		}

		device := domain.DiscoveredDevice{
			IP:        currentHost.IP,
			MAC:       mac,
			Vendor:    vendor,
			Hostname:  hostname,
			LatencyMs: currentHost.LatencyMs,
			Status:    "up",
			OpenPorts: openPorts,
			ScannedAt: scanTimestamp,
		}
		devices = append(devices, device)

		// Salvataggio incrementale nel contesto di sessione
		s.mu.Lock()
		if sRes, ok := s.sessionResults[sessionID]; ok {
			sRes.Devices = append(sRes.Devices, device)
			sRes.TotalDiscovered = len(sRes.Devices)
		}
		s.mu.Unlock()

		// Notifica push istantanea SSE del dispositivo identificato
		s.publishEvent(topic, domain.SSEEvent{
			Event: domain.EventDeviceDiscovered,
			Data: domain.DeviceDiscoveredPayload{
				SessionID: sessionID,
				Device:    device,
			},
		})

		// Avanzamento percentuale progressivo tra 50% e 99%
		progressPct := 50.0 + (float64(idx+1)/float64(len(upHosts)))*49.0
		s.publishEvent(topic, domain.SSEEvent{
			Event: domain.EventDiscoveryProgress,
			Data: domain.DiscoveryProgressPayload{
				SessionID:    sessionID,
				ScannedHosts: idx + 1,
				TotalHosts:   len(upHosts),
				Percentage:   progressPct,
				CurrentIP:    device.IP,
			},
		})
	}

	durationSec := report.Duration
	if durationSec <= 0 {
		durationSec = time.Since(startTime).Seconds()
	}

	resultStarted := report.StartTime
	if resultStarted.IsZero() {
		resultStarted = startTime
	}

	discoveryResult := &domain.DiscoveryResult{
		ID:              sessionID,
		TenantID:        tenantID,
		Subnet:          targetInfo.CleanValue,
		Devices:         devices,
		TotalDiscovered: len(devices),
		DurationSeconds: durationSec,
		StartedAt:       resultStarted,
		CompletedAt:     scanTimestamp,
	}

	// 11. Salvataggio in cache in-memory per il tenant e la sessione
	s.mu.Lock()
	s.results[tenantID] = discoveryResult
	s.sessionResults[sessionID] = discoveryResult
	s.mu.Unlock()

	// Notifica evento di completamento finale SSE
	s.publishEvent(topic, domain.SSEEvent{
		Event: domain.EventDiscoveryComplete,
		Data: domain.DiscoveryCompletePayload{
			SessionID:       sessionID,
			TotalDiscovered: len(devices),
			DurationSeconds: durationSec,
		},
	})

	s.logger.Info("completata scansione LAN discovery arricchita",
		slog.String("tenant_id", tenantID.String()),
		slog.String("subnet", targetInfo.CleanValue),
		slog.Int("total_discovered", len(devices)),
		slog.Float64("duration_seconds", durationSec),
	)

	return discoveryResult, nil
}

// isNetworkOrBroadcast verifica se un IP è l'indirizzo di rete o di broadcast per la subnet
func isNetworkOrBroadcast(ipStr string, cleanTarget string) bool {
	ip := net.ParseIP(ipStr).To4()
	if ip == nil {
		return false
	}

	_, ipNet, err := net.ParseCIDR(cleanTarget)
	if err == nil && ipNet != nil && ipNet.IP.To4() != nil {
		if ip.Equal(ipNet.IP) {
			return true
		}
		mask := ipNet.Mask
		broadcast := make(net.IP, len(ipNet.IP.To4()))
		for i := 0; i < len(broadcast); i++ {
			broadcast[i] = ipNet.IP.To4()[i] | ^mask[i]
		}
		if ip.Equal(broadcast) {
			return true
		}
	}
	return false
}

// GetLastResults recupera l'ultimo set di risultati di ricognizione memorizzati per il tenant
func (s *discoveryService) GetLastResults(ctx context.Context, tenantID uuid.UUID) (*domain.DiscoveryResult, error) {
	if tenantID == uuid.Nil {
		return nil, fmt.Errorf("%w: tenant ID non valido", domain.ErrValidation)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	res, ok := s.results[tenantID]
	if !ok {
		return nil, domain.ErrNotFound
	}

	return res, nil
}

// GetSessionResults recupera i risultati parziali o completi di una specifica sessione di scansione
func (s *discoveryService) GetSessionResults(ctx context.Context, tenantID, sessionID uuid.UUID) (*domain.DiscoveryResult, error) {
	if tenantID == uuid.Nil || sessionID == uuid.Nil {
		return nil, fmt.Errorf("%w: parametri identificativi non validi", domain.ErrValidation)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	res, ok := s.sessionResults[sessionID]
	if !ok || res.TenantID != tenantID {
		return nil, domain.ErrNotFound
	}

	// Restituisce una copia shallow sicura
	copied := *res
	copied.Devices = make([]domain.DiscoveredDevice, len(res.Devices))
	copy(copied.Devices, res.Devices)

	return &copied, nil
}

func (s *discoveryService) publishEvent(topic string, event domain.SSEEvent) {
	if s.broadcaster != nil {
		s.broadcaster.Publish(topic, event)
	}
}

// LookupVendor espone direttamente il matching OUI MAC -> Vendor hardware
func (s *discoveryService) LookupVendor(mac string) string {
	return s.ouiDB.Lookup(mac)
}

// resolveHostnameFallback tenta la risoluzione inversa DNS dell'IP con un timeout breve (300ms)
func (s *discoveryService) resolveHostnameFallback(ctx context.Context, ip string) string {
	if ip == "" {
		return ""
	}

	lookupCtx, cancel := context.WithTimeout(ctx, 300*time.Millisecond)
	defer cancel()

	resolver := &net.Resolver{}
	names, err := resolver.LookupAddr(lookupCtx, ip)
	if err != nil || len(names) == 0 {
		return ""
	}

	return strings.TrimSuffix(names[0], ".")
}

// resolveHostnamesConcurrently risolve in parallelo gli hostname degli host attivi privi di PrimaryHostname
func (s *discoveryService) resolveHostnamesConcurrently(ctx context.Context, hosts []scanner.ParsedHost, maxWorkers int) map[string]string {
	resolved := make(map[string]string)
	if len(hosts) == 0 {
		return resolved
	}

	var ipsToResolve []string
	for _, h := range hosts {
		if h.IP != "" && h.PrimaryHostname == "" {
			if _, exists := resolved[h.IP]; !exists {
				resolved[h.IP] = ""
				ipsToResolve = append(ipsToResolve, h.IP)
			}
		} else if h.IP != "" && h.PrimaryHostname != "" {
			resolved[h.IP] = h.PrimaryHostname
		}
	}

	if len(ipsToResolve) == 0 {
		return resolved
	}

	if maxWorkers <= 0 {
		maxWorkers = 25
	}
	if maxWorkers > len(ipsToResolve) {
		maxWorkers = len(ipsToResolve)
	}

	poolCtx, poolCancel := context.WithTimeout(ctx, 8*time.Second)
	defer poolCancel()

	type hostResult struct {
		ip       string
		hostname string
	}

	jobs := make(chan string, len(ipsToResolve))
	results := make(chan hostResult, len(ipsToResolve))

	var wg sync.WaitGroup
	for w := 0; w < maxWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			resolver := &net.Resolver{}
			for ip := range jobs {
				select {
				case <-poolCtx.Done():
					return
				default:
				}

				itemCtx, itemCancel := context.WithTimeout(poolCtx, 1200*time.Millisecond)
				names, err := resolver.LookupAddr(itemCtx, ip)
				itemCancel()

				hostname := ""
				if err == nil && len(names) > 0 {
					hostname = strings.TrimSuffix(names[0], ".")
				}

				select {
				case results <- hostResult{ip: ip, hostname: hostname}:
				case <-poolCtx.Done():
					return
				}
			}
		}()
	}

	for _, ip := range ipsToResolve {
		jobs <- ip
	}
	close(jobs)

	go func() {
		wg.Wait()
		close(results)
	}()

	for res := range results {
		if res.hostname != "" {
			resolved[res.ip] = res.hostname
		}
	}

	return resolved
}

package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"

	"nextgen-cybersec/internal/domain"
	"nextgen-cybersec/internal/scanner"
	"nextgen-cybersec/pkg/validator"
)

// ScanJob rappresenta un'unità di lavoro da processare nel pool asincrono
type ScanJob struct {
	ScanID    uuid.UUID `json:"scan_id"`
	TenantID  uuid.UUID `json:"tenant_id"`
	CreatedAt time.Time `json:"created_at"`
}

// OrchestratorConfig definisce i parametri operativi del pool di worker (NFR-PERF-01)
type OrchestratorConfig struct {
	MaxWorkers int // Numero massimo di scansioni eseguite in parallelo (es. 2-4)
	QueueSize  int // Capacità del buffer in-memory per le scansioni in attesa
}

// DefaultOrchestratorConfig restituisce una configurazione predefinita sicura per preservare le risorse di rete
func DefaultOrchestratorConfig() OrchestratorConfig {
	return OrchestratorConfig{
		MaxWorkers: 2,
		QueueSize:  100,
	}
}

// ScanOrchestrator coordina l'esecuzione asincrona delle scansioni di rete
type ScanOrchestrator struct {
	cfg         OrchestratorConfig
	scanRepo    domain.ScanRepository
	profileRepo domain.ScanProfileRepository
	targetRepo  domain.TargetRepository
	runner      scanner.Runner
	logger      *slog.Logger

	findingRepo    domain.FindingRepository
	correlator     Correlator
	broadcaster    EventBroadcaster
	jobs           chan *ScanJob
	activeScans    map[uuid.UUID]context.CancelFunc
	progressByScan map[uuid.UUID]domain.ScanProgressPayload
	activeMu       sync.RWMutex

	workersWg sync.WaitGroup
	ctx       context.Context
	cancel    context.CancelFunc
	running   atomic.Bool
}

// SetFindingRepo configura il repository per salvare i risultati dei servizi scoperti
func (o *ScanOrchestrator) SetFindingRepo(repo domain.FindingRepository) {
	o.findingRepo = repo
}

// SetCorrelator configura il motore di correlazione vulnerabilità CVE
func (o *ScanOrchestrator) SetCorrelator(correlator Correlator) {
	o.correlator = correlator
}

// SetEventBroadcaster configura il servizio broker di streaming SSE in tempo reale
func (o *ScanOrchestrator) SetEventBroadcaster(broadcaster EventBroadcaster) {
	o.broadcaster = broadcaster
}

func (o *ScanOrchestrator) publishEvent(topic string, event domain.SSEEvent) {
	o.activeMu.Lock()
	switch payload := event.Data.(type) {
	case domain.ScanPhasePayload:
		if o.progressByScan == nil {
			o.progressByScan = make(map[uuid.UUID]domain.ScanProgressPayload)
		}
		o.progressByScan[payload.ScanID] = domain.ScanProgressPayload{ScanID: payload.ScanID, Phase: payload.Phase, Percentage: payload.Percentage, Message: payload.Message, ObservedAt: time.Now().UTC()}
	case domain.ScanProgressPayload:
		if o.progressByScan == nil {
			o.progressByScan = make(map[uuid.UUID]domain.ScanProgressPayload)
		}
		o.progressByScan[payload.ScanID] = payload
	case domain.ScanStatusPayload:
		if payload.Status != domain.ScanStatusRunning && payload.Status != domain.ScanStatusQueued {
			delete(o.progressByScan, payload.ScanID)
		}
	}
	o.activeMu.Unlock()
	if o.broadcaster != nil {
		o.broadcaster.Publish(topic, event)
	}
}

// NewScanOrchestrator istanzia un nuovo orchestratore con worker pool
func NewScanOrchestrator(
	cfg OrchestratorConfig,
	scanRepo domain.ScanRepository,
	profileRepo domain.ScanProfileRepository,
	targetRepo domain.TargetRepository,
	runner scanner.Runner,
	logger *slog.Logger,
) *ScanOrchestrator {
	if cfg.MaxWorkers <= 0 {
		cfg.MaxWorkers = 2
	}
	if cfg.QueueSize <= 0 {
		cfg.QueueSize = 100
	}
	if logger == nil {
		logger = slog.Default()
	}

	return &ScanOrchestrator{
		cfg:            cfg,
		scanRepo:       scanRepo,
		profileRepo:    profileRepo,
		targetRepo:     targetRepo,
		runner:         runner,
		logger:         logger,
		jobs:           make(chan *ScanJob, cfg.QueueSize),
		activeScans:    make(map[uuid.UUID]context.CancelFunc),
		progressByScan: make(map[uuid.UUID]domain.ScanProgressPayload),
	}
}

// Start avvia i worker concorrenti in background
func (o *ScanOrchestrator) Start(parentCtx context.Context) {
	if o.running.Swap(true) {
		return // Già avviato
	}

	o.ctx, o.cancel = context.WithCancel(parentCtx)

	o.logger.Info("Avvio Scan Orchestrator Worker Pool",
		"max_workers", o.cfg.MaxWorkers,
		"queue_size", o.cfg.QueueSize,
	)

	for i := 1; i <= o.cfg.MaxWorkers; i++ {
		o.workersWg.Add(1)
		go o.worker(i)
	}
}

// Stop arresta gracefulmente il worker pool attendendo il completamento delle scansioni in corso
func (o *ScanOrchestrator) Stop() {
	if !o.running.Swap(false) {
		return
	}

	o.logger.Info("Arresto Scan Orchestrator: invio segnale di terminazione ai worker...")
	if o.cancel != nil {
		o.cancel()
	}

	// Annullamento forzato di tutte le scansioni attive rimaste
	o.activeMu.Lock()
	for scanID, cancelFunc := range o.activeScans {
		o.logger.Warn("Arresto forzato scansione attiva per shutdown", "scan_id", scanID)
		cancelFunc()
	}
	o.activeMu.Unlock()

	o.workersWg.Wait()
	o.logger.Info("Scan Orchestrator arrestato con successo")
}

// Enqueue accoda una scansione già persistita su DB per l'elaborazione asincrona
func (o *ScanOrchestrator) Enqueue(tenantID, scanID uuid.UUID) error {
	if !o.running.Load() {
		return errors.New("orchestrator non avviato")
	}

	job := &ScanJob{
		ScanID:    scanID,
		TenantID:  tenantID,
		CreatedAt: time.Now(),
	}

	select {
	case o.jobs <- job:
		o.logger.Info("Scansione accodata con successo nel worker pool",
			"scan_id", scanID,
			"tenant_id", tenantID,
		)
		return nil
	default:
		o.logger.Error("Coda scansioni satura, impossibile accodare",
			"scan_id", scanID,
			"queue_capacity", o.cfg.QueueSize,
		)
		return errors.New("coda scansioni satura")
	}
}

// CancelScan richiede l'annullamento immediato di una scansione in esecuzione (NFR-REL-01)
func (o *ScanOrchestrator) CancelScan(tenantID, scanID uuid.UUID) error {
	o.activeMu.RLock()
	cancelFunc, exists := o.activeScans[scanID]
	o.activeMu.RUnlock()

	if exists && cancelFunc != nil {
		o.logger.Info("Invocazione cancel context per scansione attiva",
			"scan_id", scanID,
			"tenant_id", tenantID,
		)
		cancelFunc()
		return nil
	}

	// Se la scansione non è in activeScans, potrebbe essere ancora in stato QUEUED su DB
	scan, err := o.scanRepo.GetByID(context.Background(), tenantID, scanID)
	if err != nil {
		return err
	}

	if !scan.CanCancel() {
		return domain.ErrScanCannotBeCancelled
	}

	if err := scan.MarkCancelled(); err != nil {
		return err
	}

	return o.scanRepo.Update(context.Background(), scan)
}

// registerActiveScan registra la funzione di cancel della scansione corrente
func (o *ScanOrchestrator) registerActiveScan(scanID uuid.UUID, cancelFunc context.CancelFunc) {
	o.activeMu.Lock()
	defer o.activeMu.Unlock()
	o.activeScans[scanID] = cancelFunc
}

// unregisterActiveScan deregistra la scansione completata o interrotta
func (o *ScanOrchestrator) unregisterActiveScan(scanID uuid.UUID) {
	o.activeMu.Lock()
	defer o.activeMu.Unlock()
	delete(o.activeScans, scanID)
	delete(o.progressByScan, scanID)
}

// GetProgress is consulted only after the service has verified tenant ownership.
func (o *ScanOrchestrator) GetProgress(scanID uuid.UUID) *domain.ScanProgressPayload {
	o.activeMu.RLock()
	defer o.activeMu.RUnlock()
	progress, ok := o.progressByScan[scanID]
	if !ok {
		return nil
	}
	return &progress
}

// worker gestisce il ciclo di vita del singolo thread consumatore
func (o *ScanOrchestrator) worker(workerID int) {
	defer o.workersWg.Done()

	o.logger.Debug("Worker avviato", "worker_id", workerID)

	for {
		select {
		case <-o.ctx.Done():
			o.logger.Debug("Worker terminato su context done", "worker_id", workerID)
			return

		case job, ok := <-o.jobs:
			if !ok {
				return
			}
			o.processJob(job, workerID)
		}
	}
}

// processJob esegue l'intera sequenza di scansione e le transizioni atomiche di stato
func (o *ScanOrchestrator) processJob(job *ScanJob, workerID int) {
	ctx := o.ctx

	o.logger.Info("Presa in carico job di scansione",
		"worker_id", workerID,
		"scan_id", job.ScanID,
		"tenant_id", job.TenantID,
	)

	topic := domain.ScanTopic(job.ScanID)

	// 1. Recupero record scansione dal database
	scan, err := o.scanRepo.GetByID(ctx, job.TenantID, job.ScanID)
	if err != nil {
		o.logger.Error("Recupero record scansione fallito", "scan_id", job.ScanID, "error", err)
		return
	}

	// Se la scansione è stata cancellata mentre era in coda
	if scan.Status == domain.ScanStatusCancelled {
		o.logger.Info("Scansione già annullata durante l'attesa in coda", "scan_id", job.ScanID)
		return
	}

	// 2. Recupero del target e verifica blacklist
	target, err := o.targetRepo.GetByID(ctx, job.TenantID, scan.TargetID)
	if err != nil {
		o.failScan(scan, fmt.Sprintf("target non trovato: %v", err))
		return
	}

	if target.IsBlacklisted {
		o.failScan(scan, domain.ErrTargetBlacklisted.Error())
		return
	}

	isBlacklisted, err := o.targetRepo.IsBlacklisted(ctx, job.TenantID, target.Value)
	if err != nil || isBlacklisted {
		o.failScan(scan, domain.ErrTargetBlacklisted.Error())
		return
	}

	// 2.1 Verifica perimetro di sicurezza interno e deny-list (FR-01.4)
	if err := validator.CheckTargetScope(target.Value, false); err != nil {
		o.failScan(scan, fmt.Sprintf("bersaglio non consentito dalle policy di sicurezza: %v", err))
		return
	}

	// 3. Recupero del profilo di scansione
	profile, err := o.profileRepo.GetByID(ctx, scan.ProfileID)
	if err != nil {
		o.failScan(scan, fmt.Sprintf("profilo di scansione non trovato: %v", err))
		return
	}

	profileArgs, err := scanner.ParseProfileArgs(profile.NmapArgs)
	if err != nil {
		o.failScan(scan, fmt.Sprintf("argomenti del profilo non validi: %v", err))
		return
	}

	// 4. Transizione atomica di stato: QUEUED -> RUNNING
	if err := scan.MarkRunning(); err != nil {
		o.logger.Error("Transizione a RUNNING non valida", "scan_id", scan.ID, "error", err)
		return
	}
	if err := o.scanRepo.Update(ctx, scan); err != nil {
		o.logger.Error("Aggiornamento stato RUNNING su DB fallito", "scan_id", scan.ID, "error", err)
		return
	}

	// Notifica avvio scansione e prima fase (HOST_DISCOVERY)
	o.publishEvent(topic, domain.SSEEvent{
		Event: domain.EventScanStatusChange,
		Data: domain.ScanStatusPayload{
			ScanID:       scan.ID,
			Status:       domain.ScanStatusRunning,
			SummaryStats: scan.SummaryStats,
		},
	})
	o.publishEvent(topic, domain.SSEEvent{
		Event: domain.EventScanPhase,
		Data: domain.ScanPhasePayload{
			ScanID:     scan.ID,
			Phase:      "HOST_DISCOVERY",
			Message:    "Verifica e censimento host attivi",
			Percentage: 0,
		},
	})

	// 5. Configurazione context di scansione con timeout e registrazione per cancellation
	timeout := time.Duration(profile.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 60 * time.Minute
	}

	scanCtx, scanCancel := context.WithTimeout(ctx, timeout)
	defer scanCancel()

	o.registerActiveScan(scan.ID, scanCancel)
	defer o.unregisterActiveScan(scan.ID)

	profileConfig := scanner.ProfileConfig{
		Name:    profile.Name,
		Args:    profileArgs,
		Timeout: timeout,
	}
	progressStarted := time.Now()
	totalProgress := 0.0
	profileConfig.OnProgress = func(progress scanner.TaskProgress) {
		phase, percentage, remaining := estimateScanProgress(progress, time.Since(progressStarted).Seconds(), totalProgress)
		if phase == "" {
			return
		}
		totalProgress = percentage
		payload := domain.ScanProgressPayload{
			ScanID: scan.ID, Phase: phase, PhasePercentage: progress.Percent, Percentage: percentage,
			RemainingSeconds: remaining, Message: "Statistiche Nmap: " + progress.Task, ObservedAt: time.Now().UTC(),
		}
		o.publishEvent(topic, domain.SSEEvent{Event: domain.EventScanProgress, Data: payload})
	}

	// Notifica avvio scansione porte
	o.publishEvent(topic, domain.SSEEvent{
		Event: domain.EventScanPhase,
		Data: domain.ScanPhasePayload{
			ScanID:     scan.ID,
			Phase:      "PORT_SCAN",
			Message:    fmt.Sprintf("Scansione porte con profilo %s in corso", profile.Name),
			Percentage: 0,
		},
	})

	// Persist each completed XML host while Nmap is still running. The final
	// parse supports legacy runners without replaying hosts already committed.
	var findings []*domain.Finding
	seenHosts := make(map[string]bool)
	profileConfig.OnHost = func(h scanner.ParsedHost) error {
		if h.Status != "up" || seenHosts[h.IP] {
			return nil
		}
		now := time.Now().UTC()
		for _, p := range h.Ports {
			isRelevantPort := strings.EqualFold(p.State, "open") ||
				strings.EqualFold(p.State, "unfiltered") ||
				strings.EqualFold(p.State, "open|filtered")

			if isRelevantPort {
				var portFindings []*domain.Finding
				if o.correlator != nil {
					pFindings, err := o.correlator.CorrelatePort(scanCtx, scan.ID, scan.TenantID, h, p)
					if err != nil {
						o.logger.Warn("Errore durante correlazione CVE porta", "port", p.PortID, "error", err)
					}
					portFindings = pFindings
				} else {
					portState := p.State
					if portState == "" {
						portState = "open"
					}
					f := &domain.Finding{
						ID:              uuid.New(),
						ScanID:          scan.ID,
						TenantID:        scan.TenantID,
						HostIP:          h.IP,
						Port:            p.PortID,
						Protocol:        p.Protocol,
						PortState:       portState,
						Severity:        domain.SeverityInfo,
						Confidence:      domain.ConfidenceConfirmed,
						DetectionMethod: domain.DetectionMethodPortState,
						CreatedAt:       now,
					}
					if h.PrimaryHostname != "" {
						hn := h.PrimaryHostname
						f.Hostname = &hn
					}
					if p.Service != nil {
						f.DetectionMethod = domain.DetectionMethodBanner
						if p.Service.Name != "" {
							sName := p.Service.Name
							f.ServiceName = &sName
						}
						if p.Service.Version != "" {
							sVer := p.Service.Version
							f.ServiceVersion = &sVer
						}
						if p.Service.RawBanner != "" {
							sBanner := p.Service.RawBanner
							f.RawBanner = &sBanner
						}
						if len(p.Service.CPEs) > 0 {
							cpe := p.Service.CPEs[0]
							f.CPE = &cpe
						}
					}
					portFindings = []*domain.Finding{f}
				}

				if len(portFindings) > 0 {
					findings = append(findings, portFindings...)

					// Salvataggio incrementale immediato su PostgreSQL (NFR-REL-02)
					if o.findingRepo != nil {
						if err := o.findingRepo.CreateBatch(scanCtx, portFindings); err != nil {
							return fmt.Errorf("unable to persist scan findings")
						}
					}

					// Notifica push istantanea per ciascun finding (SSE)
					for _, pf := range portFindings {
						o.publishEvent(topic, domain.SSEEvent{
							Event: domain.EventScanFinding,
							Data: domain.ScanFindingPayload{
								ScanID:  scan.ID,
								Finding: *pf,
							},
						})
					}
				}
			}
		}

		// Correlazione vulnerabilità script a livello di host (es. SMB EternalBlue, SSL, ecc.) (FR-17)
		if o.correlator != nil && len(h.HostScripts) > 0 {
			hostFindings, err := o.correlator.CorrelateHostScripts(scanCtx, scan.ID, scan.TenantID, h)
			if err != nil {
				o.logger.Warn("Errore durante correlazione script host Nmap", "host", h.IP, "error", err)
			}
			if len(hostFindings) > 0 {
				findings = append(findings, hostFindings...)
				if o.findingRepo != nil {
					if err := o.findingRepo.CreateBatch(scanCtx, hostFindings); err != nil {
						return fmt.Errorf("unable to persist host findings")
					}
				}
				for _, hf := range hostFindings {
					o.publishEvent(topic, domain.SSEEvent{
						Event: domain.EventScanFinding,
						Data: domain.ScanFindingPayload{
							ScanID:  scan.ID,
							Finding: *hf,
						},
					})
				}
			}
		}
		seenHosts[h.IP] = true
		return nil
	}

	// 6. Esecuzione del comando Nmap tramite il Runner
	execResult, runErr := o.runner.RunScan(scanCtx, target.Value, profileConfig)

	// 7. Valutazione esito dell'esecuzione
	if scanCtx.Err() != nil {
		if errors.Is(scanCtx.Err(), context.Canceled) {
			o.logger.Warn("Scansione annullata con successo durante l'esecuzione", "scan_id", scan.ID)
			_ = scan.MarkCancelled()
			_ = o.scanRepo.Update(context.Background(), scan)
			o.publishEvent(topic, domain.SSEEvent{
				Event: domain.EventScanStatusChange,
				Data: domain.ScanStatusPayload{
					ScanID: scan.ID,
					Status: domain.ScanStatusCancelled,
				},
			})
			return
		}
		if errors.Is(scanCtx.Err(), context.DeadlineExceeded) {
			o.logger.Error("Scansione interrotta per superamento timeout", "scan_id", scan.ID)
			o.failScan(scan, fmt.Sprintf("timeout di scansione superato (%v)", timeout))
			return
		}
	}

	if runErr != nil {
		o.logger.Error("Esecuzione processo Nmap fallita", "scan_id", scan.ID, "error", runErr)
		o.failScan(scan, runErr.Error())
		return
	}

	// 8. Parsing XML dell'output in streaming
	report, parseErr := scanner.ParseNmapXMLBytes(execResult.RawXML)
	if parseErr != nil {
		o.logger.Error("Parsing XML Nmap fallito", "scan_id", scan.ID, "error", parseErr)
		o.failScan(scan, fmt.Sprintf("parsing output XML fallito: %v", parseErr))
		return
	}

	// Notifica fase SERVICE_DETECTION
	o.publishEvent(topic, domain.SSEEvent{
		Event: domain.EventScanPhase,
		Data: domain.ScanPhasePayload{
			ScanID:     scan.ID,
			Phase:      "SERVICE_DETECTION",
			Message:    "Rilevamento servizi ed estrazione banner",
			Percentage: 90.0,
		},
	})

	// Notifica fase CVE_CORRELATION
	o.publishEvent(topic, domain.SSEEvent{
		Event: domain.EventScanPhase,
		Data: domain.ScanPhasePayload{
			ScanID:     scan.ID,
			Phase:      "CVE_CORRELATION",
			Message:    "Correlazione vulnerabilità CVE/CVSS su feed locale",
			Percentage: 92.0,
		},
	})

	// Process hosts only when a legacy runner did not stream them.
	for _, h := range report.GetUpHosts() {
		if err := profileConfig.OnHost(h); err != nil {
			o.failScan(scan, err.Error())
			return
		}
	}

	// 10. Calcolo delle statistiche riassuntive e transizione atomica: RUNNING -> COMPLETED
	summaryStats := report.ToSummaryStats()
	for _, f := range findings {
		switch f.Severity {
		case domain.SeverityCritical:
			summaryStats.CriticalCount++
		case domain.SeverityHigh:
			summaryStats.HighCount++
		case domain.SeverityMedium:
			summaryStats.MediumCount++
		case domain.SeverityLow:
			summaryStats.LowCount++
		case domain.SeverityInfo:
			summaryStats.InfoCount++
		}
	}
	if err := scan.MarkCompleted(summaryStats); err != nil {
		o.logger.Error("Transizione a COMPLETED fallita", "scan_id", scan.ID, "error", err)
		return
	}

	if err := o.scanRepo.Update(context.Background(), scan); err != nil {
		o.logger.Error("Salvataggio stato COMPLETED su DB fallito", "scan_id", scan.ID, "error", err)
		return
	}

	// Notifica finale completamento scansione (SSE)
	o.publishEvent(topic, domain.SSEEvent{
		Event: domain.EventScanStatusChange,
		Data: domain.ScanStatusPayload{
			ScanID:       scan.ID,
			Status:       domain.ScanStatusCompleted,
			SummaryStats: summaryStats,
		},
	})

	o.logger.Info("Scansione completata con successo",
		"scan_id", scan.ID,
		"up_hosts", summaryStats.UpHosts,
		"down_hosts", summaryStats.DownHosts,
		"open_ports", summaryStats.TotalOpenPorts,
		"duration", summaryStats.DurationSeconds,
	)
}

// failScan imposta lo stato a FAILED, persiste l'errore su DB ed emette eventi SSE
func (o *ScanOrchestrator) failScan(scan *domain.Scan, errMsg string) {
	topic := domain.ScanTopic(scan.ID)
	o.publishEvent(topic, domain.SSEEvent{
		Event: domain.EventScanError,
		Data: domain.ScanErrorPayload{
			ScanID:  scan.ID,
			Message: errMsg,
		},
	})

	if err := scan.MarkFailed(errMsg); err != nil {
		o.logger.Error("Transizione a FAILED non valida", "scan_id", scan.ID, "error", err)
		return
	}
	if err := o.scanRepo.Update(context.Background(), scan); err != nil {
		o.logger.Error("Salvataggio stato FAILED su DB fallito", "scan_id", scan.ID, "error", err)
	}

	o.publishEvent(topic, domain.SSEEvent{
		Event: domain.EventScanStatusChange,
		Data: domain.ScanStatusPayload{
			ScanID: scan.ID,
			Status: domain.ScanStatusFailed,
		},
	})
}

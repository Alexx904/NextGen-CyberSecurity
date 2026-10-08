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

	findingRepo domain.FindingRepository
	correlator  Correlator
	jobs        chan *ScanJob
	activeScans map[uuid.UUID]context.CancelFunc
	activeMu    sync.RWMutex

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
		cfg:         cfg,
		scanRepo:    scanRepo,
		profileRepo: profileRepo,
		targetRepo:  targetRepo,
		runner:      runner,
		logger:      logger,
		jobs:        make(chan *ScanJob, cfg.QueueSize),
		activeScans: make(map[uuid.UUID]context.CancelFunc),
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

	// 6. Esecuzione del comando Nmap tramite il Runner
	execResult, runErr := o.runner.RunScan(scanCtx, target.Value, profileConfig)

	// 7. Valutazione esito dell'esecuzione
	if scanCtx.Err() != nil {
		if errors.Is(scanCtx.Err(), context.Canceled) {
			o.logger.Warn("Scansione annullata con successo durante l'esecuzione", "scan_id", scan.ID)
			_ = scan.MarkCancelled()
			_ = o.scanRepo.Update(context.Background(), scan)
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

	// 9. Persistenza dei finding rilevati su DB
	var findings []*domain.Finding
	now := time.Now().UTC()
	for _, h := range report.Hosts {
		for _, p := range h.Ports {
			if strings.EqualFold(p.State, "open") {
				if o.correlator != nil {
					portFindings, err := o.correlator.CorrelatePort(ctx, scan.ID, scan.TenantID, h, p)
					if err != nil {
						o.logger.Warn("Errore durante correlazione CVE porta", "port", p.PortID, "error", err)
					}
					findings = append(findings, portFindings...)
				} else {
					f := &domain.Finding{
						ID:        uuid.New(),
						ScanID:    scan.ID,
						TenantID:  scan.TenantID,
						HostIP:    h.IP,
						Port:      p.PortID,
						Protocol:  p.Protocol,
						Severity:  domain.SeverityInfo,
						CreatedAt: now,
					}
					if h.PrimaryHostname != "" {
						hn := h.PrimaryHostname
						f.Hostname = &hn
					}
					if p.Service != nil {
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
					}
					findings = append(findings, f)
				}
			}
		}
	}

	if o.findingRepo != nil && len(findings) > 0 {
		if err := o.findingRepo.CreateBatch(context.Background(), findings); err != nil {
			o.logger.Error("Salvataggio batch findings su DB fallito", "scan_id", scan.ID, "error", err)
		} else {
			o.logger.Info("Findings registrati con successo", "scan_id", scan.ID, "count", len(findings))
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

	o.logger.Info("Scansione completata con successo",
		"scan_id", scan.ID,
		"up_hosts", summaryStats.UpHosts,
		"open_ports", summaryStats.TotalOpenPorts,
		"duration", summaryStats.DurationSeconds,
	)
}

// failScan imposta lo stato a FAILED e persiste l'errore su DB
func (o *ScanOrchestrator) failScan(scan *domain.Scan, errMsg string) {
	if err := scan.MarkFailed(errMsg); err != nil {
		o.logger.Error("Transizione a FAILED non valida", "scan_id", scan.ID, "error", err)
		return
	}
	if err := o.scanRepo.Update(context.Background(), scan); err != nil {
		o.logger.Error("Salvataggio stato FAILED su DB fallito", "scan_id", scan.ID, "error", err)
	}
}

package scanner

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"time"

	"nextgen-cybersec/internal/domain"
)

// ScanExecutionResult racchiude il risultato dell'invocazione a basso livello del processo Nmap
type ScanExecutionResult struct {
	RawXML   []byte        `json:"-"`
	Stdout   string        `json:"stdout,omitempty"`
	Stderr   string        `json:"stderr,omitempty"`
	ExitCode int           `json:"exit_code"`
	Duration time.Duration `json:"duration"`
}

// Runner definisce l'interfaccia di esecuzione per disaccoppiare Nmap e abilitare Clean Architecture e Mocking
type Runner interface {
	RunScan(ctx context.Context, target string, profile ProfileConfig) (*ScanExecutionResult, error)
}

// CommandFactory astrae exec.CommandContext per testabilità unitaria
type CommandFactory func(ctx context.Context, name string, args ...string) *exec.Cmd

// Option consente la personalizzazione fluida di NmapRunner
type Option func(*NmapRunner)

// WithCommandFactory sovrascrive la factory di creazione comandi di sistema (utile per mock)
func WithCommandFactory(factory CommandFactory) Option {
	return func(r *NmapRunner) {
		if factory != nil {
			r.cmdFactory = factory
		}
	}
}

// WithCancelSignal configura il segnale da inviare all'annullamento del contesto
func WithCancelSignal(sig os.Signal) Option {
	return func(r *NmapRunner) {
		if sig != nil {
			r.cancelSignal = sig
		}
	}
}

// WithWaitDelay configura il periodo di tolleranza prima del SIGKILL forzato
func WithWaitDelay(delay time.Duration) Option {
	return func(r *NmapRunner) {
		if delay > 0 {
			r.waitDelay = delay
		}
	}
}

// NmapRunner esegue il binario Nmap in modo sicuro, isolato da shell e con propagazione graceful dei segnali
type NmapRunner struct {
	binaryPath   string
	logger       *slog.Logger
	cmdFactory   CommandFactory
	cancelSignal os.Signal
	waitDelay    time.Duration
}

// NewNmapRunner inizializza una nuova istanza di NmapRunner con valori di default sicuri
func NewNmapRunner(binaryPath string, logger *slog.Logger, opts ...Option) *NmapRunner {
	if binaryPath == "" {
		binaryPath = "nmap"
	}
	if logger == nil {
		logger = slog.Default()
	}

	runner := &NmapRunner{
		binaryPath:   binaryPath,
		logger:       logger,
		cmdFactory:   exec.CommandContext,
		cancelSignal: defaultCancelSignal,
		waitDelay:    3 * time.Second, // Timeout grace period post-SIGTERM prima di SIGKILL
	}

	for _, opt := range opts {
		opt(runner)
	}

	return runner
}

// RunScan esegue la scansione sul target fornito utilizzando il profilo specificato
func (r *NmapRunner) RunScan(ctx context.Context, target string, profile ProfileConfig) (*ScanExecutionResult, error) {
	// 1. Verifica preliminare del contesto
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	start := time.Now()

	// 2. Costruzione e validazione rigorosa degli argomenti (NFR-SEC-01)
	args, err := BuildSafeArguments(profile.Args, target)
	if err != nil {
		r.logger.Error("Validazione argomenti di scansione fallita",
			"target", target,
			"profile", profile.Name,
			"error", err,
		)
		return nil, err
	}

	r.logger.Info("Esecuzione comando Nmap",
		"binary", r.binaryPath,
		"profile", profile.Name,
		"target", target,
		"args", strings.Join(args, " "),
	)

	// 3. Creazione del processo senza shell di sistema (exec.CommandContext diretto)
	cmd := r.cmdFactory(ctx, r.binaryPath, args...)

	// 4. Configurazione Graceful Cancellation: propagazione SIGTERM e WaitDelay (NFR-REL-01)
	cmd.Cancel = func() error {
		if cmd.Process != nil {
			r.logger.Warn("Annullamento scansione: propagazione segnale di arresto al processo",
				"signal", r.cancelSignal.String(),
				"pid", cmd.Process.Pid,
			)
			return cmd.Process.Signal(r.cancelSignal)
		}
		return nil
	}
	cmd.WaitDelay = r.waitDelay

	var stdoutBuf bytes.Buffer
	var stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	// 5. Esecuzione del processo
	runErr := cmd.Run()
	duration := time.Since(start)

	exitCode := 0
	if runErr != nil {
		var exitErr *exec.ExitError
		if errors.As(runErr, &exitErr) {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1
		}
	}

	result := &ScanExecutionResult{
		RawXML:   stdoutBuf.Bytes(),
		Stdout:   stdoutBuf.String(),
		Stderr:   stderrBuf.String(),
		ExitCode: exitCode,
		Duration: duration,
	}

	// 6. Gestione esito del contesto
	if ctx.Err() != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			r.logger.Error("Scansione interrotta per timeout",
				"target", target,
				"duration", duration,
			)
			return result, fmt.Errorf("%w: timeout dopo %v", domain.ErrScanTimeout, duration)
		}
		if errors.Is(ctx.Err(), context.Canceled) {
			r.logger.Warn("Scansione annullata su richiesta utente",
				"target", target,
				"duration", duration,
			)
			return result, context.Canceled
		}
		return result, ctx.Err()
	}

	// 7. Gestione errore del processo Nmap
	if runErr != nil {
		r.logger.Error("Esecuzione Nmap terminata con errore",
			"exit_code", exitCode,
			"stderr", stderrBuf.String(),
			"duration", duration,
		)
		return result, fmt.Errorf("%w: exit code %d, stderr: %s", domain.ErrScannerExecFailed, exitCode, strings.TrimSpace(stderrBuf.String()))
	}

	r.logger.Info("Scansione Nmap completata con successo",
		"target", target,
		"duration", duration,
		"raw_xml_bytes", len(result.RawXML),
	)

	return result, nil
}


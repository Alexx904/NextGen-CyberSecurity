package scanner

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"nextgen-cybersec/internal/domain"
)

// TestHelperProcess simula l'esecuzione del processo di sistema durante i test unitari
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_TEST_HELPER_PROCESS") != "1" {
		return
	}

	mode := os.Getenv("HELPER_MODE")
	switch mode {
	case "progress":
		if !strings.Contains(strings.Join(os.Args, " "), "--stats-every 5s -- 192.168.1.10") {
			os.Exit(2)
		}
		fmt.Print(`<nmaprun><taskprogress task="SYN Stealth Scan" percent="25" remaining="90"/></nmaprun>`)
		os.Exit(0)
	case "success":
		fmt.Print(`<?xml version="1.0" encoding="UTF-8"?><nmaprun scanner="nmap" args="nmap -oX - 192.168.1.1"></nmaprun>`)
		os.Exit(0)
	case "error":
		fmt.Fprint(os.Stderr, "QUITTING! Failed to resolve target host")
		os.Exit(2)
	case "sigkill":
		fmt.Fprint(os.Stderr, "Killed by signal 9 (SIGKILL)")
		os.Exit(137)
	case "truncated_xml":
		fmt.Print(`<?xml version="1.0" encoding="UTF-8"?><nmaprun scanner="nmap" args="nmap -oX - 192.168.1.1"><host><status state="up"/><address addr="192.168.1.1" addrtype="ipv4"/><ports><port protocol="tcp" portid="80"><state state="open"/>`)
		os.Exit(0)
	case "corrupted_xml":
		fmt.Print("\x00\x01\x02\xFF\xFE\xFD<not-an-xml-doc><<<>>>")
		os.Exit(0)
	case "large_stderr":
		for i := 0; i < 500; i++ {
			fmt.Fprintln(os.Stderr, "Nmap warning: dropped probe packet detected on interface eth0")
		}
		os.Exit(1)
	case "sleep":
		// Rimane in attesa per testare context cancellation o timeout
		time.Sleep(10 * time.Second)
		os.Exit(0)
	default:
		os.Exit(1)
	}
}

// createMockCommandFactory genera una factory che invoca TestHelperProcess come sub-processo
func createMockCommandFactory(mode string) CommandFactory {
	return func(ctx context.Context, name string, args ...string) *exec.Cmd {
		cs := []string{"-test.run=TestHelperProcess", "--"}
		cs = append(cs, args...)
		cmd := exec.CommandContext(ctx, os.Args[0], cs...)
		cmd.Env = append(os.Environ(),
			"GO_TEST_HELPER_PROCESS=1",
			"HELPER_MODE="+mode,
		)
		return cmd
	}
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestGetDefaultProfiles(t *testing.T) {
	profiles := GetDefaultProfiles()

	expectedProfiles := []ProfileType{
		ProfileFastDiscovery,
		ProfileStandardScan,
		ProfileFullScan,
		ProfileLANDiscovery,
	}

	for _, expected := range expectedProfiles {
		p, exists := profiles[expected]
		if !exists {
			t.Errorf("profilo atteso %s mancante nei profili di default", expected)
		}
		if len(p.Args) == 0 {
			t.Errorf("profilo %s non definisce argomenti", expected)
		}
		if p.Timeout <= 0 {
			t.Errorf("profilo %s ha timeout nullo o non valido", expected)
		}
	}
}

func TestBuildSafeArguments_Success(t *testing.T) {
	tests := []struct {
		name          string
		profileArgs   []string
		target        string
		expectedLast  string
		shouldContain []string
	}{
		{
			name:          "Fast Discovery con IP",
			profileArgs:   []string{"-T4", "-F"},
			target:        "192.168.1.10",
			expectedLast:  "192.168.1.10",
			shouldContain: []string{"-oX", "-", "-T4", "-F"},
		},
		{
			name:          "LAN Discovery con ARP ping e CIDR",
			profileArgs:   []string{"-sn", "-PR", "-T4"},
			target:        "192.168.1.0/24",
			expectedLast:  "192.168.1.0/24",
			shouldContain: []string{"-oX", "-", "-sn", "-PR", "-T4"},
		},
		{
			name:          "Standard Service Scan con CIDR",
			profileArgs:   []string{"-sV", "-T4"},
			target:        "10.0.0.0/24",
			expectedLast:  "10.0.0.0/24",
			shouldContain: []string{"-oX", "-", "-sV", "-T4"},
		},
		{
			name:          "Full Deep Scan con FQDN",
			profileArgs:   []string{"-p-", "-sV", "-O", "-T4"},
			target:        "scanme.nmap.org",
			expectedLast:  "scanme.nmap.org",
			shouldContain: []string{"-oX", "-", "-p-", "-sV", "-O", "-T4"},
		},
		{
			name:          "Porte specifiche composte",
			profileArgs:   []string{"-p", "80,443", "-sV"},
			target:        "172.16.0.5",
			expectedLast:  "172.16.0.5",
			shouldContain: []string{"-p80,443", "-sV"},
		},
		{
			name:          "IPv6 auto-flag -6 injection",
			profileArgs:   []string{"-T4", "-F"},
			target:        "2001:db8::1",
			expectedLast:  "2001:db8::1",
			shouldContain: []string{"-6", "-T4", "-F"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args, err := BuildSafeArguments(tt.profileArgs, tt.target)
			if err != nil {
				t.Fatalf("BuildSafeArguments ha restituito errore inatteso: %v", err)
			}

			if len(args) < 3 {
				t.Fatalf("slice argomenti troppo corta: %v", args)
			}

			// Verifica che i primi due argomenti forzino sempre output XML su stdout
			if args[0] != "-oX" || args[1] != "-" {
				t.Errorf("i primi argomenti devono essere '-oX -', trovati: %s %s", args[0], args[1])
			}

			// Verifica che l'ultimo argomento sia il target
			if args[len(args)-1] != tt.expectedLast {
				t.Errorf("l'ultimo argomento deve essere il target %s, trovato: %s", tt.expectedLast, args[len(args)-1])
			}

			// Verifica presenza dei flag richiesti
			joined := strings.Join(args, " ")
			for _, exp := range tt.shouldContain {
				if !strings.Contains(joined, exp) {
					t.Errorf("argomenti %v non contengono '%s'", args, exp)
				}
			}
		})
	}
}

func TestBuildSafeArguments_Rejections(t *testing.T) {
	tests := []struct {
		name        string
		profileArgs []string
		target      string
		wantErr     bool
		targetErr   bool
	}{
		{
			name:        "Bersaglio non valido (flag injection)",
			profileArgs: []string{"-T4"},
			target:      "-sV",
			wantErr:     true,
			targetErr:   true,
		},
		{
			name:        "Bersaglio con metacaratteri shell",
			profileArgs: []string{"-T4"},
			target:      "192.168.1.1; cat /etc/passwd",
			wantErr:     true,
			targetErr:   true,
		},
		{
			name:        "Flag vietato --script arbitrario",
			profileArgs: []string{"-T4", "--script", "malicious.nse"},
			target:      "192.168.1.1",
			wantErr:     true,
			targetErr:   false,
		},
		{
			name:        "Flag vietato -oN output redirection",
			profileArgs: []string{"-T4", "-oN", "/tmp/out.txt"},
			target:      "192.168.1.1",
			wantErr:     true,
			targetErr:   false,
		},
		{
			name:        "Flag vietato -iL file input",
			profileArgs: []string{"-iL", "/etc/targets.txt"},
			target:      "192.168.1.1",
			wantErr:     true,
			targetErr:   false,
		},
		{
			name:        "Flag vietato --interactive",
			profileArgs: []string{"--interactive"},
			target:      "192.168.1.1",
			wantErr:     true,
			targetErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := BuildSafeArguments(tt.profileArgs, tt.target)
			if (err != nil) != tt.wantErr {
				t.Fatalf("BuildSafeArguments() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && !tt.targetErr {
				if !errors.Is(err, domain.ErrDisallowedScanOption) {
					t.Errorf("atteso ErrDisallowedScanOption, ottenuto: %v", err)
				}
			}
		})
	}
}

func TestNmapRunner_RunScan_Success(t *testing.T) {
	factory := createMockCommandFactory("success")
	runner := NewNmapRunner("nmap", testLogger(), WithCommandFactory(factory))

	profile := ProfileConfig{
		Name:    "Test Fast",
		Args:    []string{"-T4", "-F"},
		Timeout: 5 * time.Second,
	}

	ctx := context.Background()
	result, err := runner.RunScan(ctx, "192.168.1.1", profile)

	if err != nil {
		t.Fatalf("RunScan ha restituito un errore inatteso: %v", err)
	}

	if result == nil {
		t.Fatal("il risultato restituito è nil")
	}

	if result.ExitCode != 0 {
		t.Errorf("atteso ExitCode 0, ottenuto %d", result.ExitCode)
	}

	if !strings.Contains(result.Stdout, "<nmaprun") {
		t.Errorf("stdout non contiene il tag XML atteso: %s", result.Stdout)
	}

	if len(result.RawXML) == 0 {
		t.Error("RawXML è vuoto")
	}
}

func TestNmapRunner_RunScan_ProcessError(t *testing.T) {
	factory := createMockCommandFactory("error")
	runner := NewNmapRunner("nmap", testLogger(), WithCommandFactory(factory))

	profile := ProfileConfig{
		Name:    "Test Error Profile",
		Args:    []string{"-T4"},
		Timeout: 5 * time.Second,
	}

	ctx := context.Background()
	result, err := runner.RunScan(ctx, "192.168.1.1", profile)

	if err == nil {
		t.Fatal("atteso errore di esecuzione processo, ma err è nil")
	}

	if !errors.Is(err, domain.ErrScannerExecFailed) {
		t.Errorf("atteso ErrScannerExecFailed, ottenuto %v", err)
	}

	if result == nil {
		t.Fatal("il risultato non dovrebbe essere nil anche in caso di errore")
	}

	if result.ExitCode != 2 {
		t.Errorf("atteso exit code 2, ottenuto %d", result.ExitCode)
	}

	if !strings.Contains(result.Stderr, "Failed to resolve target host") {
		t.Errorf("stderr atteso non trovato nel risultato: %s", result.Stderr)
	}
}

func TestNmapRunner_RunScan_ContextCancellation(t *testing.T) {
	factory := createMockCommandFactory("sleep")
	runner := NewNmapRunner("nmap", testLogger(),
		WithCommandFactory(factory),
		WithWaitDelay(1*time.Second),
	)

	profile := ProfileConfig{
		Name:    "Slow Scan",
		Args:    []string{"-T4"},
		Timeout: 10 * time.Second,
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Annullamento immediato dopo 150ms
	time.AfterFunc(150*time.Millisecond, cancel)

	start := time.Now()
	_, err := runner.RunScan(ctx, "192.168.1.1", profile)
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("atteso context.Canceled, ottenuto nil")
	}

	if !errors.Is(err, context.Canceled) {
		t.Errorf("atteso context.Canceled, ottenuto %v", err)
	}

	// Verifica che il processo si sia interrotto rapidamente senza aspettare i 10s dello sleep
	if elapsed > 3*time.Second {
		t.Errorf("il processo ha impiegato troppo tempo per arrestarsi (%v)", elapsed)
	}
}

func TestNmapRunner_RunScan_Timeout(t *testing.T) {
	factory := createMockCommandFactory("sleep")
	runner := NewNmapRunner("nmap", testLogger(),
		WithCommandFactory(factory),
		WithWaitDelay(1*time.Second),
	)

	profile := ProfileConfig{
		Name:    "Timeout Scan",
		Args:    []string{"-T4"},
		Timeout: 200 * time.Millisecond,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	_, err := runner.RunScan(ctx, "192.168.1.1", profile)

	if err == nil {
		t.Fatal("atteso ErrScanTimeout, ottenuto nil")
	}

	if !errors.Is(err, domain.ErrScanTimeout) {
		t.Errorf("atteso ErrScanTimeout, ottenuto %v", err)
	}
}

func TestNmapRunner_RunScan_InvalidTargetBeforeExec(t *testing.T) {
	// Se il target non è valido, il processo non deve nemmeno essere avviato
	invoked := false
	customFactory := func(ctx context.Context, name string, args ...string) *exec.Cmd {
		invoked = true
		return exec.Command("true")
	}

	runner := NewNmapRunner("nmap", testLogger(), WithCommandFactory(customFactory))
	profile := ProfileConfig{Name: "Fast", Args: []string{"-T4"}}

	_, err := runner.RunScan(context.Background(), "-dangerous-flag", profile)
	if err == nil {
		t.Fatal("atteso errore per target non valido, ottenuto nil")
	}

	if invoked {
		t.Error("il comando di sistema non doveva essere istanziato per un target non valido")
	}
}

func TestNmapRunner_SuddenTermination_SIGKILL(t *testing.T) {
	// Simula la terminazione forzata del processo (es. OOM killer, kill -9 / SIGKILL esterno)
	factory := createMockCommandFactory("sigkill")
	runner := NewNmapRunner("nmap", testLogger(), WithCommandFactory(factory))

	profile := ProfileConfig{
		Name:    "Abrupt Kill Profile",
		Args:    []string{"-T4"},
		Timeout: 5 * time.Second,
	}

	result, err := runner.RunScan(context.Background(), "192.168.1.50", profile)

	if err == nil {
		t.Fatal("atteso errore di terminazione processo, ma err è nil")
	}

	if !errors.Is(err, domain.ErrScannerExecFailed) {
		t.Errorf("atteso ErrScannerExecFailed, ottenuto: %v", err)
	}

	if result == nil {
		t.Fatal("il risultato non deve essere nil anche in caso di kill violento")
	}

	if result.ExitCode != 137 {
		t.Errorf("atteso exit code 137 (SIGKILL), ottenuto %d", result.ExitCode)
	}

	if !strings.Contains(result.Stderr, "SIGKILL") {
		t.Errorf("stderr atteso non trovato: %s", result.Stderr)
	}
}

func TestNmapRunner_TruncatedXMLOutput_GracefulFallback(t *testing.T) {
	// Simula output XML troncato a metà a causa di crash o interruzione improvvisa
	factory := createMockCommandFactory("truncated_xml")
	runner := NewNmapRunner("nmap", testLogger(), WithCommandFactory(factory))

	profile := ProfileConfig{
		Name:    "Truncated XML Profile",
		Args:    []string{"-T4"},
		Timeout: 5 * time.Second,
	}

	result, err := runner.RunScan(context.Background(), "192.168.1.1", profile)
	if err != nil {
		t.Fatalf("RunScan del processo terminato con exit 0 non dovrebbe dare runErr, ottenuto: %v", err)
	}

	if len(result.RawXML) == 0 {
		t.Fatal("RawXML non deve essere vuoto")
	}

	// Verifica che il parser gestisca elegantemente l'XML troncato restituendo errore senza andare in panic
	report, parseErr := ParseNmapXMLBytes(result.RawXML)
	if parseErr == nil {
		t.Errorf("atteso errore di parsing per XML troncato, ottenuto report: %+v", report)
	}
}

func TestNmapRunner_CorruptedXMLOutput_GracefulFallback(t *testing.T) {
	// Simula output binario/corrotto non XML
	factory := createMockCommandFactory("corrupted_xml")
	runner := NewNmapRunner("nmap", testLogger(), WithCommandFactory(factory))

	profile := ProfileConfig{
		Name:    "Corrupted XML Profile",
		Args:    []string{"-T4"},
		Timeout: 5 * time.Second,
	}

	result, err := runner.RunScan(context.Background(), "192.168.1.1", profile)
	if err != nil {
		t.Fatalf("RunScan non dovrebbe dare runErr su exit 0, ottenuto: %v", err)
	}

	// Verifica che ParseNmapXMLBytes fallisca in modo pulito e sicuro
	_, parseErr := ParseNmapXMLBytes(result.RawXML)
	if parseErr == nil {
		t.Fatal("atteso parseErr per XML corrotto, ottenuto nil")
	}
}

func TestNmapRunner_LargeStderrBuffer_NoDeadlock(t *testing.T) {
	// Verifica che un volume ingente di stderr non causi stallo/deadlock sui buffer di pipe
	factory := createMockCommandFactory("large_stderr")
	runner := NewNmapRunner("nmap", testLogger(), WithCommandFactory(factory))

	profile := ProfileConfig{
		Name:    "Large Stderr Profile",
		Args:    []string{"-T4"},
		Timeout: 5 * time.Second,
	}

	done := make(chan struct{})
	var result *ScanExecutionResult
	var err error

	go func() {
		result, err = runner.RunScan(context.Background(), "192.168.1.1", profile)
		close(done)
	}()

	select {
	case <-done:
		if err == nil {
			t.Fatal("atteso errore di exit code, ottenuto nil")
		}
		if result == nil || len(result.Stderr) < 1000 {
			t.Errorf("atteso buffer stderr completo, lunghezza: %d", len(result.Stderr))
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timeout: deadlock rilevato nella lettura del buffer stderr di Nmap")
	}
}

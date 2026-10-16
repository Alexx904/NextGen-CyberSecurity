package scanner

import (
	"context"
	"strings"
	"testing"
)

func TestProgressWriterFragmentedAndInvalid(t *testing.T) {
	var received []TaskProgress
	w := &progressWriter{onProgress: func(p TaskProgress) { received = append(received, p) }}
	xml := `<nmaprun><taskprogress task="SYN Stealth Scan" percent="50" remaining="12"/><taskprogress task="NSE" percent="10"/><taskprogress task="NSE" percent="NaN"/><taskprogress task="NSE" percent="101"/><taskprogress task="NSE"/><taskprogress task="NSE" percent="40" remaining="-1"/></nmaprun>`
	for i := 0; i < len(xml); i += 7 {
		end := i + 7
		if end > len(xml) {
			end = len(xml)
		}
		if n, err := w.Write([]byte(xml[i:end])); n != end-i || err != nil {
			t.Fatal(n, err)
		}
	}
	if len(received) != 3 || *received[0].Percent != 50 || *received[0].Remaining != 12 || received[1].Remaining != nil || received[2].Remaining != nil {
		t.Fatalf("unexpected telemetry: %+v", received)
	}
	w.Write([]byte(strings.Repeat("x", 1000000)))
	if len(w.pending) > 8192 {
		t.Fatal("unbounded progress buffer")
	}
	w.Write([]byte("<taskprogress " + strings.Repeat("x", 1000000)))
	if len(w.pending) > 8192 {
		t.Fatal("unbounded malformed fragment")
	}
}

func TestRunnerEmitsTelemetryAndPreservesXML(t *testing.T) {
	var received []TaskProgress
	runner := NewNmapRunner("nmap", testLogger(), WithCommandFactory(createMockCommandFactory("progress")))
	result, err := runner.RunScan(context.Background(), "192.168.1.10", ProfileConfig{Args: []string{"-sS"}, OnProgress: func(p TaskProgress) { received = append(received, p) }})
	if err != nil {
		t.Fatal(err)
	}
	if len(received) != 1 || *received[0].Percent != 25 {
		t.Fatalf("missing progress: %+v", received)
	}
	if !strings.Contains(string(result.RawXML), "taskprogress") {
		t.Fatal("XML was altered")
	}
}

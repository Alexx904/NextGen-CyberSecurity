package scanner

import (
	"context"
	"errors"
	"io"
	"os/exec"
	"testing"
	"time"
)

func TestNmapRunnerHostArrivesBeforeProcessExit(t *testing.T) {
	input, ack := io.Pipe()
	defer input.Close()
	defer ack.Close()
	factory := createMockCommandFactory("live_hosts")
	runner := NewNmapRunner("fixture", testLogger(), WithCommandFactory(func(c context.Context, n string, a ...string) *exec.Cmd {
		cmd := factory(c, n, a...)
		cmd.Stdin = input
		return cmd
	}))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	hosts := 0
	profile := ProfileConfig{Args: []string{"-sn"}, OnHost: func(h ParsedHost) error {
		if h.IP != "192.168.1.10" || h.Status != "up" {
			t.Errorf("incorrect host: %+v", h)
		}
		hosts++
		// The child cannot finish until this live callback acknowledges its host.
		_, err := ack.Write([]byte{'x'})
		_ = ack.Close()
		return err
	}}
	result, err := runner.RunScan(ctx, "192.168.1.10", profile)
	if err != nil || hosts != 1 {
		t.Fatalf("host not delivered live: hosts=%d err=%v", hosts, err)
	}
	if _, err := ParseNmapXMLBytes(result.RawXML); err != nil {
		t.Fatal("streaming altered complete XML", err)
	}
}

func TestNmapRunnerHostCallbackFailureStopsProcess(t *testing.T) {
	input, ack := io.Pipe()
	defer input.Close()
	defer ack.Close()
	factory := createMockCommandFactory("live_hosts")
	runner := NewNmapRunner("fixture", testLogger(), WithCommandFactory(func(c context.Context, n string, a ...string) *exec.Cmd {
		cmd := factory(c, n, a...)
		cmd.Stdin = input
		return cmd
	}))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	want := errors.New("persistence unavailable")
	_, err := runner.RunScan(ctx, "192.168.1.10", ProfileConfig{Args: []string{"-sn"}, OnHost: func(ParsedHost) error { _ = ack.Close(); return want }})
	if !errors.Is(err, want) {
		t.Fatalf("callback error not preserved: %v", err)
	}
}

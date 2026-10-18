package service

import (
	"context"
	_ "embed"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"nextgen-cybersec/internal/domain"
)

//go:embed firewall_remote.sh
var remoteFirewallScript string

type remoteFirewallState struct {
	ID, TenantID, ServerID uuid.UUID
	ExpiresAt              time.Time
	Timeout                int
}

func remoteFirewallCommand(action string, tenantID, serverID, token uuid.UUID, timeout int, args ...string) string {
	words := []string{"sh", "-c", quoteFirewallArgument(remoteFirewallScript), "nextgen-firewall", quoteFirewallArgument(action), tenantID.String(), serverID.String(), token.String(), strconv.Itoa(timeout)}
	for _, arg := range args {
		words = append(words, quoteFirewallArgument(arg))
	}
	return strings.Join(words, " ")
}

func (s *firewallService) remoteCall(ctx context.Context, tenantID, serverID, token uuid.UUID, timeout int, action, input string, args ...string) (string, error) {
	if tenantID == uuid.Nil || serverID == uuid.Nil {
		return "", domain.ErrValidation
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	command := remoteFirewallCommand(action, tenantID, serverID, token, timeout, args...)
	var output, stderr string
	var code int
	var err error
	if input != "" {
		executor, ok := s.executor.(SSHInputExecutor)
		if !ok {
			return "", fmt.Errorf("%w: SSH transport does not support snapshot input", domain.ErrSSHExecutionFailed)
		}
		output, stderr, code, err = executor.ExecuteCommandWithInput(ctx, tenantID, serverID, command, input)
	} else {
		output, stderr, code, err = s.executor.ExecuteCommand(ctx, tenantID, serverID, command)
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	diagnostic := strings.ToLower(stderr)
	for _, message := range []string{"a password is required", "not in the sudoers", "not allowed to execute", "must have a tty", "must be root", "permission denied"} {
		if strings.Contains(diagnostic, message) {
			return "", domain.ErrFirewallPrivilegesRequired
		}
	}
	switch code {
	case 73:
		return "", domain.ErrSafetyTimerActive
	case 75:
		return "", domain.ErrNoSafetyTimerActive
	case 76:
		return "", domain.ErrFirewallStaleState
	case 126:
		return "", domain.ErrFirewallPrivilegesRequired
	case 127:
		return "", domain.ErrFirewallUnavailable
	}
	if err != nil || code != 0 {
		return "", fmt.Errorf("%w: remote firewall %s failed (exit %d)", domain.ErrSSHExecutionFailed, action, code)
	}
	return output, nil
}

func parseRemoteFirewallState(output string) (*remoteFirewallState, string, error) {
	header, body, ok := strings.Cut(output, "\n")
	if !ok {
		return nil, "", domain.ErrSSHExecutionFailed
	}
	if strings.HasPrefix(body, "REVISION ") {
		if _, err := remoteFirewallRevision(output); err != nil {
			return nil, "", err
		}
		_, body, _ = strings.Cut(body, "\n")
	}
	if header == "NONE" {
		if !strings.HasPrefix(body, "RULES\n") {
			return nil, "", domain.ErrSSHExecutionFailed
		}
		return nil, strings.TrimPrefix(body, "RULES\n"), nil
	}
	fields := strings.Fields(header)
	if len(fields) != 6 || fields[0] != "PENDING" || !strings.HasPrefix(body, "RULES\n") {
		return nil, "", domain.ErrSSHExecutionFailed
	}
	id, e1 := uuid.Parse(fields[1])
	tenant, e2 := uuid.Parse(fields[2])
	server, e3 := uuid.Parse(fields[3])
	expiry, e4 := strconv.ParseInt(fields[4], 10, 64)
	timeout, e5 := strconv.Atoi(fields[5])
	if e1 != nil || e2 != nil || e3 != nil || e4 != nil || e5 != nil || id == uuid.Nil || tenant == uuid.Nil || server == uuid.Nil || expiry <= 0 || timeout < 10 || timeout > 300 {
		return nil, "", domain.ErrSSHExecutionFailed
	}
	return &remoteFirewallState{ID: id, TenantID: tenant, ServerID: server, ExpiresAt: time.Unix(expiry, 0).UTC(), Timeout: timeout}, strings.TrimPrefix(body, "RULES\n"), nil
}

func validFirewallRevision(value string) bool {
	if len(value) != 64 || strings.ToLower(value) != value {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func remoteFirewallRevision(output string) (string, error) {
	_, body, _ := strings.Cut(output, "\n")
	if !strings.HasPrefix(body, "REVISION ") {
		return "", nil
	}
	line, _, _ := strings.Cut(body, "\n")
	value := strings.TrimPrefix(line, "REVISION ")
	if !validFirewallRevision(value) {
		return "", domain.ErrSSHExecutionFailed
	}
	return value, nil
}

func parseRemoteFirewallBackup(output string, token uuid.UUID) (string, time.Time, error) {
	header, body, ok := strings.Cut(output, "\n")
	fields := strings.Fields(header)
	if !ok || len(fields) != 4 || fields[0] != "BACKUP" || fields[1] != token.String() || strings.TrimSpace(body) == "" {
		return "", time.Time{}, domain.ErrSSHExecutionFailed
	}
	expiry, err := strconv.ParseInt(fields[2], 10, 64)
	if err != nil || expiry <= time.Now().Unix() {
		return "", time.Time{}, domain.ErrNoSafetyTimerActive
	}
	return body, time.Unix(expiry, 0), nil
}

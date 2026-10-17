package service_test

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"nextgen-cybersec/internal/domain"
	"nextgen-cybersec/internal/service"
	"strings"
	"testing"
)

func TestSSHClientLabelIndependentAndBackwardCompatible(t *testing.T) {
	repo := newMockSSHServerRepo()
	svc := service.NewSSHService(repo, strings.Repeat("t", 32))
	tenant := uuid.New()
	ctx := context.Background()
	server, err := svc.CreateServer(ctx, tenant, service.CreateSSHServerRequest{Name: "Test", ClientName: "  Client A  ", Host: "192.0.2.1", Port: 22, Username: "test", AuthType: domain.SSHAuthTypePassword, RawSecret: "synthetic-test"})
	if err != nil {
		t.Fatal(err)
	}
	if server.ClientName != "Client A" {
		t.Fatal("client not normalized")
	}
	req := service.UpdateSSHServerRequest{Name: server.Name, Host: server.Host, Port: 22, Username: server.Username, AuthType: server.AuthType}
	updated, err := svc.UpdateServer(ctx, tenant, server.ID, req)
	if err != nil {
		t.Fatal(err)
	}
	if updated.ClientName != "Client A" {
		t.Fatal("legacy update removed client")
	}
	empty := ""
	req.ClientName = &empty
	updated, err = svc.UpdateServer(ctx, tenant, server.ID, req)
	if err != nil {
		t.Fatal(err)
	}
	if updated.ClientName != "" {
		t.Fatal("explicit clearing failed")
	}
	long := strings.Repeat("a", 161)
	req.ClientName = &long
	if _, err = svc.UpdateServer(ctx, tenant, server.ID, req); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("oversized client accepted: %v", err)
	}
}

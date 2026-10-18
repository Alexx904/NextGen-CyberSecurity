package handler_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"nextgen-cybersec/internal/delivery/http/handler"
	"nextgen-cybersec/internal/domain"
	"nextgen-cybersec/internal/service"
)

func TestDiscoveryReplaySubscribesBeforeReadingSnapshot(t *testing.T) {
	tenant, session := uuid.New(), uuid.New()
	b := service.NewEventBroadcaster(8, nil)
	defer b.Close()
	mock := &mockDiscoveryService{getSessionResultsFunc: func(context.Context, uuid.UUID, uuid.UUID) (*domain.DiscoveryResult, error) {
		if b.SubscriberCount(domain.TenantDiscoveryTopic(tenant, session)) != 1 {
			t.Fatal("snapshot read before subscription")
		}
		if b.SubscriberCount(domain.TenantDiscoveryTopic(uuid.New(), session)) != 0 {
			t.Fatal("tenant topic collision")
		}
		b.Publish(domain.TenantDiscoveryTopic(tenant, session), domain.SSEEvent{Event: domain.EventDiscoveryComplete, Data: domain.DiscoveryCompletePayload{SessionID: session, TotalDiscovered: 1}})
		return &domain.DiscoveryResult{ID: session, TenantID: tenant, Devices: []domain.DiscoveredDevice{{IP: "192.168.1.10"}}}, nil
	}}
	ctx, cancel := context.WithTimeout(setupDiscoveryContext(tenant), time.Second)
	defer cancel()
	req := httptest.NewRequest(http.MethodGet, "/network-discovery/stream?session_id="+session.String(), nil).WithContext(ctx)
	w := httptest.NewRecorder()
	handler.NewDiscoveryHandler(mock, nil, nil, b).Stream(w, req)
	if !strings.Contains(w.Body.String(), "event: device_discovered") || !strings.Contains(w.Body.String(), "event: complete") {
		t.Fatalf("replay/live completion lost: %s", w.Body.String())
	}
	if b.SubscriberCount(domain.TenantDiscoveryTopic(tenant, session)) != 0 {
		t.Fatal("subscription leaked")
	}
}

func TestDiscoveryReplayStorageFailureIsVisible(t *testing.T) {
	tenant, session := uuid.New(), uuid.New()
	b := service.NewEventBroadcaster(8, nil)
	defer b.Close()
	mock := &mockDiscoveryService{getSessionResultsFunc: func(context.Context, uuid.UUID, uuid.UUID) (*domain.DiscoveryResult, error) {
		return nil, errors.New("test database unavailable")
	}}
	req := httptest.NewRequest(http.MethodGet, "/network-discovery/stream?session_id="+session.String(), nil).WithContext(setupDiscoveryContext(tenant))
	w := httptest.NewRecorder()
	handler.NewDiscoveryHandler(mock, nil, nil, b).Stream(w, req)
	if !strings.Contains(w.Body.String(), "event: error") || strings.Contains(w.Body.String(), "test database unavailable") {
		t.Fatalf("unsafe replay failure: %s", w.Body.String())
	}
}

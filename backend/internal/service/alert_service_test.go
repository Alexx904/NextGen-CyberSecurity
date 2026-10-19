package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"nextgen-cybersec/internal/domain"
)

type alertMemoryRepo struct {
	domain.AlertRepository
	channel *domain.AlertChannel
	job     *domain.AlertJob
	code    string
	retry   bool
	delay   time.Duration
}

func (r *alertMemoryRepo) GetChannel(_ context.Context, tenant, id uuid.UUID) (*domain.AlertChannel, error) {
	if r.channel == nil || r.channel.ID != id || r.channel.TenantID != tenant {
		return nil, domain.ErrNotFound
	}
	copy := *r.channel
	return &copy, nil
}
func (r *alertMemoryRepo) SaveChannel(_ context.Context, c *domain.AlertChannel, _ bool) error {
	c.UpdatedAt = time.Now()
	copy := *c
	r.channel = &copy
	return nil
}
func (r *alertMemoryRepo) ClaimJob(context.Context) (*domain.AlertJob, error) { return r.job, nil }
func (r *alertMemoryRepo) CompleteJob(_ context.Context, _ *domain.AlertJob, code string, retry bool, delay time.Duration) error {
	r.code, r.retry, r.delay = code, retry, delay
	r.job = nil
	return nil
}

type alertTestDispatcher struct {
	calls  int
	result AlertSendResult
}

func (d *alertTestDispatcher) Send(context.Context, domain.AlertChannelKind, domain.AlertChannelConfig, domain.AlertEvent) AlertSendResult {
	d.calls++
	return d.result
}

func TestAlertEncryptedConfigurationAndStaleUpdate(t *testing.T) {
	repo := &alertMemoryRepo{}
	dispatcher := &alertTestDispatcher{result: AlertSendResult{Code: "ACCEPTED"}}
	s := NewAlertService(repo, dispatcher, []byte("test-only-master-key-32-bytes-or-longer"), nil)
	tenant := uuid.New()
	config := domain.AlertChannelConfig{Endpoint: "https://example.com/test-only", HMACSecret: strings.Repeat("synthetic-key", 3)}
	req := domain.AlertChannelRequest{Name: "QA", Kind: domain.AlertWebhook, Enabled: true, Events: []string{"finding.critical"}, Config: &config}
	c, err := s.Save(context.Background(), tenant, uuid.Nil, req)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(c.EncryptedConfig, config.HMACSecret) {
		t.Fatal("plaintext persisted")
	}
	public, _ := json.Marshal(c)
	if strings.Contains(string(public), "config") || strings.Contains(string(public), "example.com") {
		t.Fatal("configuration leaked in response")
	}
	plain, err := s.decrypt(*c)
	if err != nil || plain.Endpoint != config.Endpoint {
		t.Fatal("decrypt", err)
	}
	foreign := *c
	foreign.TenantID = uuid.New()
	if _, err = s.decrypt(foreign); err == nil {
		t.Fatal("cross-tenant decrypt")
	}
	foreign = *c
	foreign.ID = uuid.New()
	if _, err = s.decrypt(foreign); err == nil {
		t.Fatal("channel-bound ciphertext accepted elsewhere")
	}
	if _, err = s.Save(context.Background(), tenant, c.ID, req); !errors.Is(err, domain.ErrConflict) {
		t.Fatal("stale update accepted")
	}
	req.ExpectedUpdatedAt = &c.UpdatedAt
	req.Config = nil
	updated, err := s.Save(context.Background(), tenant, c.ID, req)
	if err != nil || updated.EncryptedConfig != c.EncryptedConfig {
		t.Fatal("omitted configuration lost", err)
	}
	if _, err = s.Test(context.Background(), tenant, domain.AlertWebhook, config, false); !errors.Is(err, domain.ErrValidation) || dispatcher.calls != 0 {
		t.Fatal("external test without acknowledgement")
	}
	if _, err = s.Test(context.Background(), tenant, domain.AlertWebhook, config, true); err != nil || dispatcher.calls != 1 {
		t.Fatal("explicit test", err)
	}
}

func TestAlertWorkerRechecksScopeAndPersistsRetry(t *testing.T) {
	tenant := uuid.New()
	repo := &alertMemoryRepo{}
	dispatcher := &alertTestDispatcher{result: AlertSendResult{Code: "REMOTE_RETRY", Retry: true, RetryAfter: time.Minute}}
	s := NewAlertService(repo, dispatcher, []byte("test-only-master-key-32-bytes-or-longer"), nil)
	config := domain.AlertChannelConfig{Endpoint: "https://example.com/test", HMACSecret: strings.Repeat("synthetic-key", 3)}
	c, err := s.Save(context.Background(), tenant, uuid.Nil, domain.AlertChannelRequest{Name: "QA", Kind: domain.AlertWebhook, Enabled: true, Events: []string{"finding.critical"}, Config: &config})
	if err != nil {
		t.Fatal(err)
	}
	job := func() *domain.AlertJob {
		return &domain.AlertJob{ID: uuid.New(), Channel: *c, Attempts: 1, Event: domain.AlertEvent{ID: uuid.New(), TenantID: tenant, Kind: "finding.critical"}}
	}
	repo.job = job()
	if worked, err := s.ProcessOne(context.Background()); !worked || err != nil || dispatcher.calls != 1 || !repo.retry || repo.delay != time.Minute {
		t.Fatalf("retry not persisted: %s %v", repo.code, err)
	}
	repo.job = job()
	repo.job.Event.TenantID = uuid.New()
	if _, err = s.ProcessOne(context.Background()); err != nil || dispatcher.calls != 1 || repo.code != "CHANNEL_CHANGED" {
		t.Fatal("foreign event dispatched")
	}
	repo.job = job()
	repo.channel.UpdatedAt = repo.channel.UpdatedAt.Add(time.Second)
	if _, err = s.ProcessOne(context.Background()); err != nil || dispatcher.calls != 1 {
		t.Fatal("changed destination dispatched")
	}
}

func TestAlertDefaultEncryptionKeyRefused(t *testing.T) {
	for _, key := range []string{"", "short", "nextgen-master-ssh-encryption-key-32b"} {
		dispatcher := &alertTestDispatcher{}
		s := NewAlertService(&alertMemoryRepo{}, dispatcher, []byte(key), nil)
		config := domain.AlertChannelConfig{Endpoint: "https://example.com/test", HMACSecret: strings.Repeat("synthetic-key", 3)}
		if _, err := s.Test(context.Background(), uuid.New(), domain.AlertWebhook, config, true); !errors.Is(err, domain.ErrAlertEncryptionUnavailable) || dispatcher.calls != 0 {
			t.Fatal("default or short key allowed external test")
		}
		if _, err := s.encrypt(domain.AlertChannel{TenantID: uuid.New(), ID: uuid.New(), Kind: domain.AlertWebhook}, config); !errors.Is(err, domain.ErrAlertEncryptionUnavailable) {
			t.Fatal("default or short encryption key accepted")
		}
	}
}

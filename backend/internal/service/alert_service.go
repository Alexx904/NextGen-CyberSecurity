package service

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"nextgen-cybersec/internal/domain"
	"nextgen-cybersec/pkg/crypto"
)

var alertEventKinds = map[string]bool{"finding.high": true, "finding.critical": true, "host.discovered": true, "host.changed": true, "port.discovered": true, "scan.completed": true, "scan.failed": true}

type AlertService struct {
	repo       domain.AlertRepository
	dispatcher AlertDispatcher
	master     []byte
	logger     *slog.Logger
}

func NewAlertService(repo domain.AlertRepository, dispatcher AlertDispatcher, master []byte, logger *slog.Logger) *AlertService {
	if logger == nil {
		logger = slog.Default()
	}
	return &AlertService{repo: repo, dispatcher: dispatcher, master: append([]byte(nil), master...), logger: logger}
}
func (s *AlertService) List(ctx context.Context, tenant uuid.UUID) ([]domain.AlertChannel, error) {
	if tenant == uuid.Nil {
		return nil, domain.ErrValidation
	}
	return s.repo.ListChannels(ctx, tenant)
}
func validateAlertRequest(req domain.AlertChannelRequest) error {
	if len(strings.TrimSpace(req.Name)) == 0 || len(req.Name) > 160 || len(req.Events) == 0 || len(req.Events) > 7 || (req.CompanyID != nil && *req.CompanyID == uuid.Nil) {
		return domain.ErrValidation
	}
	seen := map[string]bool{}
	for _, kind := range req.Events {
		if !alertEventKinds[kind] || seen[kind] {
			return domain.ErrValidation
		}
		seen[kind] = true
	}
	switch req.Kind {
	case domain.AlertSlack, domain.AlertTeams, domain.AlertTelegram, domain.AlertWebhook, domain.AlertSMTP:
	default:
		return domain.ErrValidation
	}
	return nil
}
func alertAAD(c domain.AlertChannel) []byte {
	return []byte("nextgen-alert-channel:" + c.TenantID.String() + ":" + c.ID.String() + ":" + string(c.Kind))
}
func (s *AlertService) usableKey() bool {
	return len(s.master) >= 32 && string(s.master) != "nextgen-master-ssh-encryption-key-32b"
}
func (s *AlertService) encrypt(c domain.AlertChannel, config domain.AlertChannelConfig) (string, error) {
	if !s.usableKey() {
		return "", domain.ErrAlertEncryptionUnavailable
	}
	key, err := crypto.DeriveTenantKey(s.master, c.TenantID)
	if err != nil {
		return "", err
	}
	data, err := json.Marshal(config)
	if err != nil {
		return "", err
	}
	return crypto.EncryptAES256GCM(data, key, alertAAD(c))
}
func (s *AlertService) decrypt(c domain.AlertChannel) (domain.AlertChannelConfig, error) {
	var config domain.AlertChannelConfig
	if !s.usableKey() {
		return config, domain.ErrAlertEncryptionUnavailable
	}
	key, err := crypto.DeriveTenantKey(s.master, c.TenantID)
	if err != nil {
		return config, err
	}
	data, err := crypto.DecryptAES256GCM(c.EncryptedConfig, key, alertAAD(c))
	if err != nil {
		return config, err
	}
	err = json.Unmarshal(data, &config)
	return config, err
}
func (s *AlertService) Save(ctx context.Context, tenant, id uuid.UUID, req domain.AlertChannelRequest) (*domain.AlertChannel, error) {
	if tenant == uuid.Nil || validateAlertRequest(req) != nil {
		return nil, domain.ErrValidation
	}
	if req.Enabled && !s.usableKey() {
		return nil, domain.ErrAlertEncryptionUnavailable
	}
	create := id == uuid.Nil
	c := &domain.AlertChannel{ID: id, TenantID: tenant, CompanyID: req.CompanyID, Name: strings.TrimSpace(req.Name), Kind: req.Kind, Enabled: req.Enabled, Events: append([]string(nil), req.Events...)}
	sort.Strings(c.Events)
	if create {
		c.ID = uuid.New()
	} else {
		previous, err := s.repo.GetChannel(ctx, tenant, id)
		if err != nil {
			return nil, err
		}
		if previous == nil || previous.TenantID != tenant || previous.ID != id {
			return nil, domain.ErrNotFound
		}
		if req.ExpectedUpdatedAt == nil || !previous.UpdatedAt.Equal(*req.ExpectedUpdatedAt) {
			return nil, domain.ErrConflict
		}
		if previous.Kind != req.Kind && req.Config == nil {
			return nil, domain.ErrValidation
		}
		c.EncryptedConfig, c.UpdatedAt = previous.EncryptedConfig, previous.UpdatedAt
	}
	if req.Config != nil {
		if ValidateAlertConfig(req.Kind, *req.Config) != nil {
			return nil, domain.ErrValidation
		}
		cipher, err := s.encrypt(*c, *req.Config)
		if err != nil {
			return nil, err
		}
		c.EncryptedConfig = cipher
	} else if create {
		return nil, domain.ErrValidation
	}
	if err := s.repo.SaveChannel(ctx, c, create); err != nil {
		return nil, err
	}
	return c, nil
}
func (s *AlertService) Delete(ctx context.Context, tenant, id uuid.UUID) error {
	if tenant == uuid.Nil || id == uuid.Nil {
		return domain.ErrValidation
	}
	return s.repo.DeleteChannel(ctx, tenant, id)
}

func (s *AlertService) Test(ctx context.Context, tenant uuid.UUID, kind domain.AlertChannelKind, config domain.AlertChannelConfig, confirmed bool) (AlertSendResult, error) {
	if tenant == uuid.Nil || !confirmed || ValidateAlertConfig(kind, config) != nil {
		return AlertSendResult{}, domain.ErrValidation
	}
	if !s.usableKey() {
		return AlertSendResult{}, domain.ErrAlertEncryptionUnavailable
	}
	ctx, cancel := context.WithTimeout(ctx, 12*time.Second)
	defer cancel()
	event := domain.AlertEvent{ID: uuid.New(), TenantID: uuid.Nil, Kind: "channel.test", EntityID: "synthetic-test-only", OccurredAt: time.Now().UTC()}
	return s.dispatcher.Send(ctx, kind, config, event), nil
}

func alertRetryDelay(job *domain.AlertJob, result AlertSendResult) time.Duration {
	attempt := job.Attempts
	if attempt < 1 {
		attempt = 1
	}
	if attempt > 6 {
		attempt = 6
	}
	delay := 5 * time.Second * time.Duration(1<<uint(attempt-1))
	delay += time.Duration(job.ID[0]) * 10 * time.Millisecond
	if result.RetryAfter > delay {
		delay = result.RetryAfter
	}
	if delay > time.Hour {
		delay = time.Hour
	}
	return delay
}

func (s *AlertService) ProcessOne(ctx context.Context) (bool, error) {
	job, err := s.repo.ClaimJob(ctx)
	if err != nil || job == nil {
		return false, err
	}
	result := AlertSendResult{Code: "INVALID_JOB"}
	// Recheck live configuration after claim: tenant/scope and consent may have
	// changed. An already in-flight request cannot be recalled.
	channel, readErr := s.repo.GetChannel(ctx, job.Channel.TenantID, job.Channel.ID)
	if errors.Is(readErr, domain.ErrNotFound) {
		result.Code = "CHANNEL_DISABLED"
	} else if readErr != nil {
		result = AlertSendResult{Code: "CHANNEL_UNAVAILABLE", Retry: true}
	} else if channel != nil && channel.Enabled && channel.TenantID == job.Event.TenantID && channel.TenantID == job.Channel.TenantID && channel.ID == job.Channel.ID && channel.UpdatedAt.Equal(job.Channel.UpdatedAt) && (channel.CompanyID == nil || (job.Event.CompanyID != nil && *channel.CompanyID == *job.Event.CompanyID)) && alertChannelAccepts(*channel, job.Event.Kind) {
		config, decryptErr := s.decrypt(*channel)
		if decryptErr != nil {
			result.Code = "CONFIGURATION_UNAVAILABLE"
		} else {
			sendCtx, cancel := context.WithTimeout(ctx, 12*time.Second)
			result = s.dispatcher.Send(sendCtx, channel.Kind, config, job.Event)
			cancel()
		}
	} else {
		result.Code = "CHANNEL_CHANGED"
	}
	if err = s.repo.CompleteJob(ctx, job, result.Code, result.Retry, alertRetryDelay(job, result)); err != nil {
		return true, err
	}
	return true, nil
}
func alertChannelAccepts(channel domain.AlertChannel, kind string) bool {
	for _, event := range channel.Events {
		if event == kind {
			return true
		}
	}
	return false
}

func (s *AlertService) Run(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		for i := 0; i < 20 && ctx.Err() == nil; i++ {
			worked, err := s.ProcessOne(ctx)
			if err != nil {
				if ctx.Err() == nil {
					s.logger.Warn("Alert worker operation failed")
				}
				break
			}
			if !worked {
				break
			}
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

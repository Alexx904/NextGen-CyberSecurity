package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrAlertEncryptionUnavailable = errors.New("alert encryption key unavailable")

type AlertChannelKind string

const (
	AlertSlack    AlertChannelKind = "slack"
	AlertTeams    AlertChannelKind = "teams"
	AlertTelegram AlertChannelKind = "telegram"
	AlertWebhook  AlertChannelKind = "webhook"
	AlertSMTP     AlertChannelKind = "smtp"
)

type AlertChannel struct {
	ID              uuid.UUID        `json:"id"`
	TenantID        uuid.UUID        `json:"tenant_id"`
	CompanyID       *uuid.UUID       `json:"company_id,omitempty"`
	Name            string           `json:"name"`
	Kind            AlertChannelKind `json:"kind"`
	Enabled         bool             `json:"enabled"`
	Events          []string         `json:"events"`
	CreatedAt       time.Time        `json:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at"`
	EncryptedConfig string           `json:"-"`
}

// AlertChannelConfig is write-only. Never embed this in a public response.
type AlertChannelConfig struct {
	Endpoint     string   `json:"endpoint,omitempty"`
	Token        string   `json:"token,omitempty"`
	ChatID       string   `json:"chat_id,omitempty"`
	HMACSecret   string   `json:"hmac_secret,omitempty"`
	SMTPHost     string   `json:"smtp_host,omitempty"`
	SMTPPort     int      `json:"smtp_port,omitempty"`
	SMTPUsername string   `json:"smtp_username,omitempty"`
	SMTPPassword string   `json:"smtp_password,omitempty"`
	SMTPFrom     string   `json:"smtp_from,omitempty"`
	SMTPTo       []string `json:"smtp_to,omitempty"`
}

type AlertChannelRequest struct {
	ExpectedUpdatedAt *time.Time          `json:"expected_updated_at,omitempty"`
	Name              string              `json:"name"`
	CompanyID         *uuid.UUID          `json:"company_id,omitempty"`
	Kind              AlertChannelKind    `json:"kind"`
	Enabled           bool                `json:"enabled"`
	Events            []string            `json:"events"`
	Config            *AlertChannelConfig `json:"config,omitempty"`
}

// AlertEvent contains only the small audited notification payload, never banners.
type AlertEvent struct {
	ID         uuid.UUID  `json:"id"`
	TenantID   uuid.UUID  `json:"tenant_id"`
	CompanyID  *uuid.UUID `json:"company_id,omitempty"`
	Kind       string     `json:"kind"`
	OccurredAt time.Time  `json:"occurred_at"`
	EntityID   string     `json:"entity_id"`
	HostIP     string     `json:"host_ip,omitempty"`
	Port       int        `json:"port,omitempty"`
	CVEID      string     `json:"cve_id,omitempty"`
	Severity   string     `json:"severity,omitempty"`
}

type AlertJob struct {
	ID         uuid.UUID
	Channel    AlertChannel
	Event      AlertEvent
	Attempts   int
	ClaimToken uuid.UUID
}

type AlertRepository interface {
	ListChannels(context.Context, uuid.UUID) ([]AlertChannel, error)
	GetChannel(context.Context, uuid.UUID, uuid.UUID) (*AlertChannel, error)
	SaveChannel(context.Context, *AlertChannel, bool) error
	DeleteChannel(context.Context, uuid.UUID, uuid.UUID) error
	ClaimJob(context.Context) (*AlertJob, error)
	CompleteJob(context.Context, *AlertJob, string, bool, time.Duration) error
}

package domain

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

const ScanAuthorizationStatement = "scan-authorization-v1"

var ErrScanAuthorizationRequired = errors.New("autorizzazione esplicita richiesta per lo scope e la configurazione correnti")

type AuthorizedTarget struct {
	ID    uuid.UUID `json:"id"`
	Value string    `json:"value"`
}

// ScanAuthorizationRequest contains the configuration explicitly confirmed by the
// operator. Identity and acceptance time are supplied only by the server.
type ScanAuthorizationRequest struct {
	Accepted       bool               `json:"accepted"`
	Statement      string             `json:"statement"`
	Targets        []AuthorizedTarget `json:"targets"`
	NmapArgs       string             `json:"nmap_args"`
	TimeoutSeconds int                `json:"timeout_seconds"`
	CronExpression string             `json:"cron_expression"`
}

type ScanAuthorization struct {
	ScanAuthorizationRequest
	TenantID   uuid.UUID  `json:"tenant_id"`
	UserID     uuid.UUID  `json:"user_id"`
	AcceptedAt time.Time  `json:"accepted_at"`
	ScheduleID *uuid.UUID `json:"schedule_id,omitempty"`
}

type scanAuthorizationKey struct{}

func WithScanAuthorization(ctx context.Context, tenantID, userID uuid.UUID, request *ScanAuthorizationRequest) context.Context {
	if request == nil {
		return ctx
	}
	copy := *request
	copy.Targets = append([]AuthorizedTarget(nil), request.Targets...)
	return context.WithValue(ctx, scanAuthorizationKey{}, &ScanAuthorization{
		ScanAuthorizationRequest: copy, TenantID: tenantID, UserID: userID, AcceptedAt: time.Now().UTC(),
	})
}

func ScanAuthorizationFromContext(ctx context.Context) *ScanAuthorization {
	auth, _ := ctx.Value(scanAuthorizationKey{}).(*ScanAuthorization)
	return auth
}

func (a *ScanAuthorization) Validate(tenantID uuid.UUID, targets []*Target, profile *ScanProfile, cron string) error {
	if a == nil || !a.Accepted || a.Statement != ScanAuthorizationStatement || a.TenantID != tenantID || a.UserID == uuid.Nil || a.AcceptedAt.IsZero() || profile == nil || len(targets) == 0 || len(a.Targets) != len(targets) {
		return ErrScanAuthorizationRequired
	}
	if profile.TenantID != nil && *profile.TenantID != tenantID {
		return ErrScanAuthorizationRequired
	}
	if strings.Join(strings.Fields(a.NmapArgs), " ") != strings.Join(strings.Fields(profile.NmapArgs), " ") || a.TimeoutSeconds != profile.TimeoutSeconds || a.CronExpression != cron {
		return ErrScanAuthorizationRequired
	}
	approved := make(map[uuid.UUID]string, len(a.Targets))
	for _, target := range a.Targets {
		if _, exists := approved[target.ID]; exists {
			return ErrScanAuthorizationRequired
		}
		approved[target.ID] = target.Value
	}
	for _, target := range targets {
		if target == nil || target.TenantID != tenantID || target.ID == uuid.Nil {
			return ErrScanAuthorizationRequired
		}
		value, exists := approved[target.ID]
		if !exists || value != target.Value {
			return ErrScanAuthorizationRequired
		}
	}
	return nil
}

// ForTarget retains the original actor/time and narrows the immutable evidence
// to the target of one queued job.
func (a *ScanAuthorization) ForTarget(target *Target) *ScanAuthorization {
	if a == nil {
		return nil
	}
	copy := *a
	copy.Targets = []AuthorizedTarget{{ID: target.ID, Value: target.Value}}
	return &copy
}

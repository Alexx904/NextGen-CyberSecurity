package service

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"nextgen-cybersec/internal/domain"
	"nextgen-cybersec/pkg/report"
)

type ComplianceService struct{ repo domain.ComplianceRepository }

func NewComplianceService(repo domain.ComplianceRepository) *ComplianceService {
	return &ComplianceService{repo}
}
func (s *ComplianceService) Snapshot(ctx context.Context, tenant uuid.UUID, scope domain.ComplianceScope) (*domain.ComplianceSnapshot, error) {
	return s.repo.Snapshot(ctx, tenant, scope)
}
func (s *ComplianceService) CreateAsset(ctx context.Context, tenant uuid.UUID, asset domain.AssessmentAsset) (domain.AssessmentAsset, error) {
	return s.repo.CreateAsset(ctx, tenant, asset)
}
func (s *ComplianceService) Observe(ctx context.Context, tenant, actor, asset, finding uuid.UUID) (uuid.UUID, error) {
	return s.repo.ObserveFinding(ctx, tenant, actor, asset, finding)
}
func (s *ComplianceService) Command(ctx context.Context, tenant, id uuid.UUID, command domain.IssueCommand) error {
	return s.repo.AppendCommand(ctx, tenant, id, command)
}
func (s *ComplianceService) Evidence(ctx context.Context, tenant uuid.UUID, evidence domain.ControlEvidence) error {
	return s.repo.AppendEvidence(ctx, tenant, evidence)
}

func (s *ComplianceService) Audit(ctx context.Context, tenant uuid.UUID, scope domain.ComplianceScope, metadata report.AuditMetadata) ([]byte, error) {
	if strings.TrimSpace(metadata.LegalEntity) == "" || strings.TrimSpace(metadata.Period) == "" || len(metadata.LegalEntity) > 500 || len(metadata.Signatory) > 200 || len(metadata.Period) > 200 || len(metadata.LegalDetails) > 2000 || metadata.AuditDate.IsZero() || metadata.AuditDate.After(time.Now().Add(24*time.Hour)) {
		return nil, domain.ErrValidation
	}
	snapshot, err := s.repo.Snapshot(ctx, tenant, scope)
	if err != nil {
		return nil, err
	}
	return report.GenerateAuditPDF(*snapshot, metadata)
}

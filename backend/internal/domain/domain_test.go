package domain_test

import (
	"errors"
	"testing"

	"nextgen-cybersec/internal/domain"
)

func TestUserRolesAndPermissions(t *testing.T) {
	roles := []domain.UserRole{
		domain.RoleSuperAdmin,
		domain.RoleTenantAdmin,
		domain.RoleSecOps,
		domain.RoleAuditor,
	}

	for _, role := range roles {
		if !role.IsValid() {
			t.Errorf("Atteso ruolo valido per %s", role)
		}
	}

	invalidRole := domain.UserRole("AnonymousHacker")
	if invalidRole.IsValid() {
		t.Error("Atteso ruolo non valido per AnonymousHacker")
	}

	// Verifica matrice permessi
	// SecOps
	if !domain.RoleSecOps.CanTriggerScans() {
		t.Error("SecOps dovrebbe poter avviare scansioni")
	}
	if !domain.RoleSecOps.CanManageTargets() {
		t.Error("SecOps dovrebbe poter gestire i target")
	}
	if domain.RoleSecOps.CanManageUsers() {
		t.Error("SecOps non dovrebbe poter gestire utenti")
	}
	if domain.RoleSecOps.CanManageBranding() {
		t.Error("SecOps non dovrebbe poter modificare il branding")
	}
	if domain.RoleSecOps.IsReadOnly() {
		t.Error("SecOps non è un ruolo read-only")
	}

	// Auditor
	if domain.RoleAuditor.CanTriggerScans() {
		t.Error("Auditor non deve poter avviare scansioni")
	}
	if !domain.RoleAuditor.IsReadOnly() {
		t.Error("Auditor deve essere read-only")
	}

	// TenantAdmin
	if !domain.RoleTenantAdmin.CanManageUsers() {
		t.Error("TenantAdmin deve poter gestire utenti")
	}
	if !domain.RoleTenantAdmin.CanManageBranding() {
		t.Error("TenantAdmin deve poter gestire il branding")
	}
}

func TestTargetTypeValidation(t *testing.T) {
	validTypes := []domain.TargetType{
		domain.TargetTypeIP,
		domain.TargetTypeCIDR,
		domain.TargetTypeRange,
		domain.TargetTypeFQDN,
	}

	for _, vt := range validTypes {
		if !vt.IsValid() {
			t.Errorf("Tipo target atteso valido: %s", vt)
		}
	}

	if domain.TargetType("UNKNOWN").IsValid() {
		t.Error("Tipo target UNKNOWN non dovrebbe essere valido")
	}
}

func TestScanStateMachine(t *testing.T) {
	scan := &domain.Scan{
		Status: domain.ScanStatusQueued,
	}

	if !scan.CanCancel() {
		t.Error("Una scansione QUEUED deve poter essere annullata")
	}

	// QUEUED -> RUNNING
	err := scan.MarkRunning()
	if err != nil {
		t.Fatalf("Errore transizione QUEUED -> RUNNING: %v", err)
	}
	if scan.Status != domain.ScanStatusRunning || scan.StartedAt == nil {
		t.Error("Stato non aggiornato a RUNNING o StartedAt nil")
	}

	// Tentativo transizione illecita: RUNNING -> RUNNING
	err = scan.MarkRunning()
	if err == nil || !errors.Is(err, domain.ErrInvalidStateTransition) {
		t.Errorf("Atteso ErrInvalidStateTransition per RUNNING -> RUNNING, ottenuto: %v", err)
	}

	// RUNNING -> COMPLETED
	stats := domain.ScanSummaryStats{
		TotalHosts:     1,
		UpHosts:        1,
		TotalOpenPorts: 3,
		CriticalCount:  1,
	}
	err = scan.MarkCompleted(stats)
	if err != nil {
		t.Fatalf("Errore transizione RUNNING -> COMPLETED: %v", err)
	}
	if scan.Status != domain.ScanStatusCompleted || scan.CompletedAt == nil {
		t.Error("Stato non aggiornato a COMPLETED o CompletedAt nil")
	}
	if !scan.Status.IsTerminal() {
		t.Error("COMPLETED deve essere uno stato terminale")
	}

	// Tentativo di annullare scansione già COMPLETED
	if scan.CanCancel() {
		t.Error("Una scansione COMPLETED non deve poter essere annullata")
	}
	err = scan.MarkCancelled()
	if err == nil || !errors.Is(err, domain.ErrScanCannotBeCancelled) {
		t.Errorf("Atteso ErrScanCannotBeCancelled, ottenuto: %v", err)
	}
}

func TestScanCancellationAndFailure(t *testing.T) {
	// QUEUED -> CANCELLED
	queuedScan := &domain.Scan{Status: domain.ScanStatusQueued}
	if err := queuedScan.MarkCancelled(); err != nil {
		t.Fatalf("Errore annullamento scansione QUEUED: %v", err)
	}
	if queuedScan.Status != domain.ScanStatusCancelled {
		t.Errorf("Atteso CANCELLED, ottenuto %s", queuedScan.Status)
	}

	// RUNNING -> FAILED
	runningScan := &domain.Scan{Status: domain.ScanStatusRunning}
	errMsg := "connessione Nmap interrotta da timeout"
	if err := runningScan.MarkFailed(errMsg); err != nil {
		t.Fatalf("Errore fallimento scansione RUNNING: %v", err)
	}
	if runningScan.Status != domain.ScanStatusFailed || runningScan.ErrorMessage == nil || *runningScan.ErrorMessage != errMsg {
		t.Errorf("Atteso FAILED con messaggio di errore corretto, ottenuto: %v", runningScan.ErrorMessage)
	}
}

func TestSeverityFromCVSS(t *testing.T) {
	testCases := []struct {
		score    float64
		expected domain.SeverityLevel
	}{
		{10.0, domain.SeverityCritical},
		{9.8, domain.SeverityCritical},
		{9.0, domain.SeverityCritical},
		{8.9, domain.SeverityHigh},
		{7.0, domain.SeverityHigh},
		{6.9, domain.SeverityMedium},
		{4.0, domain.SeverityMedium},
		{3.9, domain.SeverityLow},
		{0.1, domain.SeverityLow},
		{0.0, domain.SeverityInfo},
		{-1.0, domain.SeverityInfo},
	}

	for _, tc := range testCases {
		res := domain.SeverityFromCVSS(tc.score)
		if res != tc.expected {
			t.Errorf("CVSS score %.1f: atteso %s, ottenuto %s", tc.score, tc.expected, res)
		}
		if !res.IsValid() {
			t.Errorf("Severità calcolata non valida: %s", res)
		}
	}
}

func TestDefaultBrandColors(t *testing.T) {
	colors := domain.DefaultBrandColors()
	if colors.Primary == "" || colors.Secondary == "" || colors.Accent == "" {
		t.Error("DefaultBrandColors non dovrebbe avere campi vuoti")
	}
}

func TestSeverityLevelValidation(t *testing.T) {
	validSeverities := []domain.SeverityLevel{
		domain.SeverityCritical,
		domain.SeverityHigh,
		domain.SeverityMedium,
		domain.SeverityLow,
		domain.SeverityInfo,
	}

	for _, s := range validSeverities {
		if !s.IsValid() {
			t.Errorf("Attesa severità valida per %s", s)
		}
	}

	invalid := domain.SeverityLevel("EXTREME")
	if invalid.IsValid() {
		t.Error("EXTREME non dovrebbe essere una severità valida")
	}
}



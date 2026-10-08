package service_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"nextgen-cybersec/internal/domain"
	"nextgen-cybersec/internal/scanner"
	"nextgen-cybersec/internal/service"
)

// mockCVERepository implementa domain.CVERepository in memoria per i test unitari
type mockCVERepository struct {
	entries map[string]*domain.CVEEntry
}

func newMockCVERepo() *mockCVERepository {
	return &mockCVERepository{
		entries: make(map[string]*domain.CVEEntry),
	}
}

func (m *mockCVERepository) GetByID(ctx context.Context, cveID string) (*domain.CVEEntry, error) {
	if entry, ok := m.entries[cveID]; ok {
		return entry, nil
	}
	return nil, domain.ErrNotFound
}

func (m *mockCVERepository) SearchByCPEURI(ctx context.Context, cpeURI string) ([]*domain.CVEEntry, error) {
	var result []*domain.CVEEntry
	for _, e := range m.entries {
		for _, crit := range e.CPEMatchCriteria {
			if crit.CPE23URI == cpeURI {
				result = append(result, e)
				break
			}
		}
	}
	return result, nil
}

func (m *mockCVERepository) SearchByCPE(ctx context.Context, cpePrefix string, version string) ([]*domain.CVEEntry, error) {
	var result []*domain.CVEEntry
	cleanPrefix := strings.TrimSuffix(cpePrefix, "%")
	for _, e := range m.entries {
		for _, crit := range e.CPEMatchCriteria {
			if strings.HasPrefix(crit.CPE23URI, cleanPrefix) {
				if version == "" || strings.Contains(crit.CPE23URI, ":"+version+":") || strings.Contains(crit.CPE23URI, ":"+version) {
					result = append(result, e)
					break
				}
			}
		}
	}
	return result, nil
}

func (m *mockCVERepository) Search(ctx context.Context, filter domain.CVEFilter) ([]*domain.CVEEntry, int, error) {
	var result []*domain.CVEEntry
	for _, e := range m.entries {
		if filter.Query != "" && !strings.Contains(strings.ToLower(e.Description), strings.ToLower(filter.Query)) {
			continue
		}
		result = append(result, e)
	}
	return result, len(result), nil
}

func (m *mockCVERepository) UpsertBatch(ctx context.Context, entries []*domain.CVEEntry) error {
	for _, e := range entries {
		m.entries[e.CVEID] = e
	}
	return nil
}

func (m *mockCVERepository) Count(ctx context.Context) (int64, error) {
	return int64(len(m.entries)), nil
}

func (m *mockCVERepository) GetStats(ctx context.Context) (*domain.CVEStats, error) {
	stats := &domain.CVEStats{
		TotalCount:     int64(len(m.entries)),
		SeverityCounts: make(map[domain.SeverityLevel]int64),
	}
	for _, e := range m.entries {
		stats.SeverityCounts[e.Severity]++
	}
	return stats, nil
}

func TestNormalizeCPETo23(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "cpe:/a:apache:http_server:2.4.49",
			expected: "cpe:2.3:a:apache:http_server:2.4.49:*:*:*:*:*:*:*",
		},
		{
			input:    "cpe:/o:linux:linux_kernel:5.4",
			expected: "cpe:2.3:o:linux:linux_kernel:5.4:*:*:*:*:*:*:*",
		},
		{
			input:    "cpe:2.3:a:apache:http_server:2.4.49:*:*:*:*:*:*:*",
			expected: "cpe:2.3:a:apache:http_server:2.4.49:*:*:*:*:*:*:*",
		},
		{
			input:    "cpe:2.3:a:apache:http_server:2.4.49",
			expected: "cpe:2.3:a:apache:http_server:2.4.49:*:*:*:*:*:*:*",
		},
		{
			input:    "",
			expected: "",
		},
	}

	for _, tc := range tests {
		actual := service.NormalizeCPETo23(tc.input)
		if actual != tc.expected {
			t.Errorf("NormalizeCPETo23(%q) = %q, atteso %q", tc.input, actual, tc.expected)
		}
	}
}

func TestGenerateRemediation(t *testing.T) {
	remCrit := service.GenerateRemediation("Apache httpd", "2.4.49", "CVE-2021-42013", domain.SeverityCritical)
	if !strings.Contains(remCrit, "URGENTE") || !strings.Contains(remCrit, "CVE-2021-42013") {
		t.Errorf("Remediation CRITICAL non valida: %s", remCrit)
	}

	remHigh := service.GenerateRemediation("Apache httpd", "2.4.49", "CVE-2021-41773", domain.SeverityHigh)
	if !strings.Contains(remHigh, "Aggiornare") || !strings.Contains(remHigh, "CVE-2021-41773") {
		t.Errorf("Remediation HIGH non valida: %s", remHigh)
	}

	remMed := service.GenerateRemediation("OpenSSL", "1.1.1", "CVE-2022-0778", domain.SeverityMedium)
	if !strings.Contains(remMed, "Pianificare") || !strings.Contains(remMed, "CVE-2022-0778") {
		t.Errorf("Remediation MEDIUM non valida: %s", remMed)
	}
}

func TestCVECorrelator_CorrelateService_And_Port(t *testing.T) {
	repo := newMockCVERepo()

	scoreCrit := 9.8
	scoreHigh := 7.5

	repo.entries["CVE-2021-41773"] = &domain.CVEEntry{
		CVEID:       "CVE-2021-41773",
		CVSSScore:   &scoreHigh,
		Severity:    domain.SeverityHigh,
		Description: "Apache HTTP Server 2.4.49 Path Traversal",
		CPEMatchCriteria: []domain.CPEMatchCriterion{
			{
				CPE23URI:   "cpe:2.3:a:apache:http_server:2.4.49:*:*:*:*:*:*:*",
				Vulnerable: true,
			},
		},
		UpdatedAt: time.Now(),
	}

	repo.entries["CVE-2021-42013"] = &domain.CVEEntry{
		CVEID:       "CVE-2021-42013",
		CVSSScore:   &scoreCrit,
		Severity:    domain.SeverityCritical,
		Description: "Apache HTTP Server 2.4.49/2.4.50 RCE",
		CPEMatchCriteria: []domain.CPEMatchCriterion{
			{
				CPE23URI:   "cpe:2.3:a:apache:http_server:2.4.49:*:*:*:*:*:*:*",
				Vulnerable: true,
			},
			{
				CPE23URI:   "cpe:2.3:a:apache:http_server:2.4.50:*:*:*:*:*:*:*",
				Vulnerable: true,
			},
		},
		UpdatedAt: time.Now(),
	}

	correlator := service.NewCVECorrelator(repo)
	ctx := context.Background()

	// 1. Correlazione Servizio con CPE 2.2
	vulnerableService := &scanner.ParsedService{
		Name:    "http",
		Product: "Apache httpd",
		Version: "2.4.49",
		CPEs:    []string{"cpe:/a:apache:http_server:2.4.49"},
	}

	cves, err := correlator.CorrelateService(ctx, vulnerableService)
	if err != nil {
		t.Fatalf("CorrelateService fallito: %v", err)
	}
	if len(cves) != 2 {
		t.Fatalf("Attese 2 CVE per Apache 2.4.49, trovate %d", len(cves))
	}
	// Verifica ordinamento: la prima deve essere CRITICAL (CVE-2021-42013 con 9.8)
	if cves[0].CVEID != "CVE-2021-42013" {
		t.Errorf("Attesa CVE-2021-42013 come prima minaccia ordinata per CVSS, ottenuta %s", cves[0].CVEID)
	}

	// 2. Correlazione Servizio Sicuro (Nessuna CVE)
	safeService := &scanner.ParsedService{
		Name:    "ssh",
		Product: "OpenSSH",
		Version: "9.8p1",
		CPEs:    []string{"cpe:/a:openssh:openssh:9.8p1"},
	}
	safeCVEs, err := correlator.CorrelateService(ctx, safeService)
	if err != nil {
		t.Fatalf("CorrelateService su servizio sicuro fallito: %v", err)
	}
	if len(safeCVEs) != 0 {
		t.Errorf("Attese 0 CVE per OpenSSH 9.8p1, trovate %d", len(safeCVEs))
	}

	// 3. Test CorrelatePort per porta con vulnerabilità
	scanID := uuid.New()
	tenantID := uuid.New()
	host := scanner.ParsedHost{
		IP:              "192.168.1.100",
		PrimaryHostname: "webserver.local",
	}
	port := scanner.ParsedPort{
		PortID:   80,
		Protocol: "tcp",
		State:    "open",
		Service:  vulnerableService,
	}

	findings, err := correlator.CorrelatePort(ctx, scanID, tenantID, host, port)
	if err != nil {
		t.Fatalf("CorrelatePort fallito: %v", err)
	}
	if len(findings) != 2 {
		t.Fatalf("Attesi 2 findings arricchiti, ottenuti %d", len(findings))
	}

	critFinding := findings[0]
	if critFinding.CVEID == nil || *critFinding.CVEID != "CVE-2021-42013" {
		t.Errorf("Atteso CVEID CVE-2021-42013, ottenuto %v", critFinding.CVEID)
	}
	if critFinding.Severity != domain.SeverityCritical {
		t.Errorf("Attesa severità CRITICAL, ottenuta %s", critFinding.Severity)
	}
	if critFinding.Remediation == nil || !strings.Contains(*critFinding.Remediation, "URGENTE") {
		t.Errorf("Remediation mancante o non corretta nel finding")
	}

	// 4. Test CorrelatePort per porta sicura -> deve produrre 1 finding INFO
	safePort := scanner.ParsedPort{
		PortID:   22,
		Protocol: "tcp",
		State:    "open",
		Service:  safeService,
	}
	safeFindings, err := correlator.CorrelatePort(ctx, scanID, tenantID, host, safePort)
	if err != nil {
		t.Fatalf("CorrelatePort su porta sicura fallito: %v", err)
	}
	if len(safeFindings) != 1 {
		t.Fatalf("Atteso 1 finding INFO per porta sicura, ottenuti %d", len(safeFindings))
	}
	if safeFindings[0].Severity != domain.SeverityInfo {
		t.Errorf("Attesa severità INFO per porta priva di CVE, ottenuta %s", safeFindings[0].Severity)
	}
	if safeFindings[0].CVEID != nil {
		t.Errorf("CVEID atteso nil per porta priva di vulnerabilità")
	}
}

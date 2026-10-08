package service_test

import (
	"context"
	"strings"
	"testing"

	"nextgen-cybersec/internal/domain"
	"nextgen-cybersec/internal/service"
)

func TestCVESyncService_SyncFromReader_Batching(t *testing.T) {
	repo := newMockCVERepo()
	syncService := service.NewCVESyncService(repo, 2) // Batch size 2 per testare chunking

	jsonFeed := `[
		{"cve_id": "CVE-2021-0001", "cvss_score": 9.0, "description": "Vuln 1"},
		{"cve_id": "CVE-2021-0002", "cvss_score": 7.0, "description": "Vuln 2"},
		{"cve_id": "CVE-2021-0003", "cvss_score": 5.0, "description": "Vuln 3"},
		{"cve_id": "CVE-2021-0004", "cvss_score": 3.0, "description": "Vuln 4"},
		{"cve_id": "CVE-2021-0005", "cvss_score": 9.5, "description": "Vuln 5"}
	]`

	reader := strings.NewReader(jsonFeed)
	stats, err := syncService.SyncFromReader(context.Background(), reader)
	if err != nil {
		t.Fatalf("SyncFromReader fallito: %v", err)
	}

	if stats.TotalProcessed != 5 {
		t.Errorf("Attesi 5 processed, ottenuti %d", stats.TotalProcessed)
	}
	if stats.TotalUpserted != 5 {
		t.Errorf("Attesi 5 upserted, ottenuti %d", stats.TotalUpserted)
	}
	// Con batchSize = 2 su 5 elementi: batch da 2, batch da 2, e flush finale da 1 -> 3 batch
	if stats.BatchCount != 3 {
		t.Errorf("Attesi 3 batch eseguiti, ottenuti %d", stats.BatchCount)
	}

	count, _ := repo.Count(context.Background())
	if count != 5 {
		t.Errorf("Attesi 5 elementi nel repository, trovati %d", count)
	}

	// Verifica calcolo automatico severità
	e1, _ := repo.GetByID(context.Background(), "CVE-2021-0001")
	if e1.Severity != domain.SeverityCritical {
		t.Errorf("Attesa severità CRITICAL per punteggio 9.0, ottenuta %s", e1.Severity)
	}

	e3, _ := repo.GetByID(context.Background(), "CVE-2021-0003")
	if e3.Severity != domain.SeverityMedium {
		t.Errorf("Attesa severità MEDIUM per punteggio 5.0, ottenuta %s", e3.Severity)
	}
}

func TestCVESyncService_SyncDefaultFeed(t *testing.T) {
	repo := newMockCVERepo()
	syncService := service.NewCVESyncService(repo, 100)

	stats, err := syncService.SyncDefaultFeed(context.Background())
	if err != nil {
		t.Fatalf("SyncDefaultFeed fallito: %v", err)
	}

	if stats.TotalUpserted < 5 {
		t.Errorf("Attese almeno 5 CVE dal feed predefinito, ottenute %d", stats.TotalUpserted)
	}

	// Verifica presenza di CVE note incorporate
	cveRegresshion, err := repo.GetByID(context.Background(), "CVE-2024-6387")
	if err != nil {
		t.Fatalf("CVE-2024-6387 non trovata nel feed predefinito: %v", err)
	}
	if !strings.Contains(cveRegresshion.Description, "sshd") {
		t.Errorf("Descrizione inattesa per CVE-2024-6387: %s", cveRegresshion.Description)
	}

	cveApache, err := repo.GetByID(context.Background(), "CVE-2021-41773")
	if err != nil {
		t.Fatalf("CVE-2021-41773 non trovata nel feed predefinito: %v", err)
	}
	if len(cveApache.CPEMatchCriteria) == 0 {
		t.Errorf("Attesi criteri CPE per CVE-2021-41773")
	}
}

func TestCVESyncService_EdgeCases(t *testing.T) {
	repo := newMockCVERepo()
	syncService := service.NewCVESyncService(repo, 10)

	// 1. Reader nullo
	_, err := syncService.SyncFromReader(context.Background(), nil)
	if err == nil {
		t.Error("Atteso errore per reader nullo")
	}

	// 2. JSON malformato
	malformed := strings.NewReader(`[{"cve_id": "CVE-1", "cvss_score": `)
	_, err = syncService.SyncFromReader(context.Background(), malformed)
	if err == nil {
		t.Error("Atteso errore per JSON malformato")
	}

	// 3. Skip di voci prive di cve_id
	feedWithEmptyID := strings.NewReader(`[
		{"cve_id": "", "description": "Voce invalida"},
		{"cve_id": "CVE-VALID-1", "description": "Voce valida"}
	]`)
	stats, err := syncService.SyncFromReader(context.Background(), feedWithEmptyID)
	if err != nil {
		t.Fatalf("Errore inatteso: %v", err)
	}
	if stats.TotalUpserted != 1 {
		t.Errorf("Atteso 1 elemento valido upserted, ottenuti %d", stats.TotalUpserted)
	}
}

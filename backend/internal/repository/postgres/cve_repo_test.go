package postgres_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"nextgen-cybersec/internal/domain"
	"nextgen-cybersec/internal/repository/postgres"
)

func TestCVERepo_Integration(t *testing.T) {
	cfg := getTestConfig()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := postgres.New(ctx, cfg)
	if err != nil {
		t.Skipf("Database non raggiungibile per i test di integrazione (%v): salto il test", err)
		return
	}
	defer db.Close()

	// Assicuriamoci che le migrazioni (inclusa 000002_cve_indexes) siano applicate
	if err := postgres.RunMigrations(cfg.DB.DSN()); err != nil {
		t.Fatalf("Esecuzione migrazioni fallita: %v", err)
	}

	repo := postgres.NewCVERepo(db.Pool)

	// Fixture di test: dataset CVE noto
	scoreHigh := 7.5
	scoreCrit := 9.8
	scoreMed := 5.5
	pubDate := time.Date(2021, time.October, 5, 0, 0, 0, 0, time.UTC)

	fixtures := []*domain.CVEEntry{
		{
			CVEID:       "CVE-2021-41773",
			CVSSScore:   &scoreHigh,
			Severity:    domain.SeverityHigh,
			Description: "Path traversal and file disclosure in Apache HTTP Server 2.4.49",
			CPEMatchCriteria: []domain.CPEMatchCriterion{
				{
					CPE23URI:   "cpe:2.3:a:apache:http_server:2.4.49:*:*:*:*:*:*:*",
					Vulnerable: true,
				},
			},
			PublishedDate: &pubDate,
			UpdatedAt:     time.Now().UTC(),
		},
		{
			CVEID:       "CVE-2021-42013",
			CVSSScore:   &scoreCrit,
			Severity:    domain.SeverityCritical,
			Description: "Remote code execution via path traversal in Apache HTTP Server 2.4.49 and 2.4.50",
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
			PublishedDate: &pubDate,
			UpdatedAt:     time.Now().UTC(),
		},
		{
			CVEID:       "CVE-2023-38606",
			CVSSScore:   &scoreMed,
			Severity:    domain.SeverityMedium,
			Description: "State manipulation vulnerability in Apple WebKit engine",
			CPEMatchCriteria: []domain.CPEMatchCriterion{
				{
					CPE23URI:   "cpe:2.3:a:apple:webkit:16.5:*:*:*:*:*:*:*",
					Vulnerable: true,
				},
			},
			PublishedDate: &pubDate,
			UpdatedAt:     time.Now().UTC(),
		},
	}

	// 1. Test UpsertBatch
	t.Run("UpsertBatch", func(t *testing.T) {
		err := repo.UpsertBatch(ctx, fixtures)
		if err != nil {
			t.Fatalf("UpsertBatch fallito: %v", err)
		}
	})

	// 2. Test GetByID
	t.Run("GetByID", func(t *testing.T) {
		entry, err := repo.GetByID(ctx, "CVE-2021-41773")
		if err != nil {
			t.Fatalf("GetByID fallito per CVE-2021-41773: %v", err)
		}
		if entry.CVEID != "CVE-2021-41773" {
			t.Errorf("Atteso CVEID 'CVE-2021-41773', ottenuto '%s'", entry.CVEID)
		}
		if entry.Severity != domain.SeverityHigh {
			t.Errorf("Attesa severità HIGH, ottenuta '%s'", entry.Severity)
		}
		if entry.CVSSScore == nil || *entry.CVSSScore != scoreHigh {
			t.Errorf("Atteso CVSS %.1f, ottenuto %v", scoreHigh, entry.CVSSScore)
		}
		if len(entry.CPEMatchCriteria) != 1 {
			t.Errorf("Atteso 1 criterio CPE, ottenuti %d", len(entry.CPEMatchCriteria))
		}

		// Caso Not Found
		_, err = repo.GetByID(ctx, "CVE-9999-00000")
		if !errors.Is(err, domain.ErrNotFound) {
			t.Errorf("Atteso ErrNotFound per CVE inesistente, ottenuto: %v", err)
		}
	})

	// 3. Test UpsertBatch Idempotence & Update
	t.Run("UpsertBatch_IdempotenceAndUpdatedValues", func(t *testing.T) {
		updatedDesc := "Updated description for Apache 2.4.49 path traversal"
		updatedEntry := &domain.CVEEntry{
			CVEID:       "CVE-2021-41773",
			CVSSScore:   &scoreHigh,
			Severity:    domain.SeverityHigh,
			Description: updatedDesc,
			CPEMatchCriteria: []domain.CPEMatchCriterion{
				{
					CPE23URI:   "cpe:2.3:a:apache:http_server:2.4.49:*:*:*:*:*:*:*",
					Vulnerable: true,
				},
			},
			PublishedDate: &pubDate,
			UpdatedAt:     time.Now().UTC(),
		}

		if err := repo.UpsertBatch(ctx, []*domain.CVEEntry{updatedEntry}); err != nil {
			t.Fatalf("Aggiornamento via UpsertBatch fallito: %v", err)
		}

		fetched, err := repo.GetByID(ctx, "CVE-2021-41773")
		if err != nil {
			t.Fatalf("GetByID fallito dopo update: %v", err)
		}
		if fetched.Description != updatedDesc {
			t.Errorf("Descrizione attesa '%s', ottenuta '%s'", updatedDesc, fetched.Description)
		}
	})

	// 4. Test SearchByCPEURI (Exact GIN match)
	t.Run("SearchByCPEURI", func(t *testing.T) {
		matches, err := repo.SearchByCPEURI(ctx, "cpe:2.3:a:apache:http_server:2.4.49:*:*:*:*:*:*:*")
		if err != nil {
			t.Fatalf("SearchByCPEURI fallito: %v", err)
		}
		if len(matches) < 2 {
			t.Fatalf("Attese almeno 2 CVE per Apache 2.4.49, trovate %d", len(matches))
		}
		// Primo elemento deve avere il CVSS più alto (CVE-2021-42013 con 9.8)
		if matches[0].CVEID != "CVE-2021-42013" {
			t.Errorf("Atteso CVE con CVSS più alto in prima posizione (CVE-2021-42013), ottenuto %s", matches[0].CVEID)
		}
	})

	// 5. Test SearchByCPE (Prefix and version matching)
	t.Run("SearchByCPE", func(t *testing.T) {
		matches, err := repo.SearchByCPE(ctx, "cpe:2.3:a:apache:http_server", "2.4.50")
		if err != nil {
			t.Fatalf("SearchByCPE fallito: %v", err)
		}
		if len(matches) == 0 {
			t.Fatal("Attesa almeno 1 CVE per Apache 2.4.50")
		}
		found := false
		for _, m := range matches {
			if m.CVEID == "CVE-2021-42013" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Atteso CVE-2021-42013 tra i risultati per Apache 2.4.50")
		}
	})

	// 6. Test Search (Filters & Pagination)
	t.Run("Search_Filters", func(t *testing.T) {
		// Filtro per severità CRITICAL
		critSev := domain.SeverityCritical
		results, total, err := repo.Search(ctx, domain.CVEFilter{
			Severity: &critSev,
		})
		if err != nil {
			t.Fatalf("Search per severità fallita: %v", err)
		}
		if total < 1 || len(results) == 0 {
			t.Errorf("Attesa almeno 1 CVE CRITICAL, total: %d", total)
		}

		// Filtro per testo query "WebKit"
		results, total, err = repo.Search(ctx, domain.CVEFilter{
			Query: "WebKit",
		})
		if err != nil {
			t.Fatalf("Search per query text fallita: %v", err)
		}
		if total < 1 || len(results) == 0 {
			t.Errorf("Attesa almeno 1 CVE per query WebKit, total: %d", total)
		}
		if results[0].CVEID != "CVE-2023-38606" {
			t.Errorf("Atteso CVE-2023-38606, ottenuto %s", results[0].CVEID)
		}

		// Paginazione
		results, _, err = repo.Search(ctx, domain.CVEFilter{
			Limit:  1,
			Offset: 0,
		})
		if err != nil {
			t.Fatalf("Search paginata fallita: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("Atteso esattamente 1 risultato con Limit=1, ottenuti %d", len(results))
		}
	})

	// 7. Test Count & GetStats
	t.Run("Count_And_GetStats", func(t *testing.T) {
		count, err := repo.Count(ctx)
		if err != nil {
			t.Fatalf("Count fallito: %v", err)
		}
		if count < 3 {
			t.Errorf("Atteso Count >= 3, ottenuto %d", count)
		}

		stats, err := repo.GetStats(ctx)
		if err != nil {
			t.Fatalf("GetStats fallito: %v", err)
		}
		if stats.TotalCount < 3 {
			t.Errorf("Atteso stats.TotalCount >= 3, ottenuto %d", stats.TotalCount)
		}
		if stats.LastUpdatedAt == nil {
			t.Error("LastUpdatedAt non dovrebbe essere nil")
		}
		if stats.SeverityCounts[domain.SeverityCritical] < 1 {
			t.Errorf("Atteso almeno 1 conteggio CRITICAL nelle statistiche")
		}
	})
}


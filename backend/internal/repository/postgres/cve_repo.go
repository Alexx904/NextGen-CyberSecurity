package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"nextgen-cybersec/internal/domain"
)

// CVERepo implementa domain.CVERepository per PostgreSQL con supporto
// a query GIN su JSONB e batch processing ad alte prestazioni (NFR-PERF-03).
type CVERepo struct {
	pool *pgxpool.Pool
}

// NewCVERepo crea una nuova istanza di CVERepo
func NewCVERepo(pool *pgxpool.Pool) *CVERepo {
	return &CVERepo{pool: pool}
}

// GetByID recupera una singola voce CVE in base al suo identificativo univoco (es. "CVE-2021-41773")
func (r *CVERepo) GetByID(ctx context.Context, cveID string) (*domain.CVEEntry, error) {
	query := `
		SELECT cve_id, cvss_score, severity, description, cpe_match_criteria, published_date, updated_at
		FROM cve_entries
		WHERE cve_id = $1
	`

	var entry domain.CVEEntry
	var severityStr string
	var criteriaRaw []byte

	err := r.pool.QueryRow(ctx, query, cveID).Scan(
		&entry.CVEID,
		&entry.CVSSScore,
		&severityStr,
		&entry.Description,
		&criteriaRaw,
		&entry.PublishedDate,
		&entry.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("recupero CVE per ID '%s' fallito: %w", cveID, err)
	}

	entry.Severity = domain.SeverityLevel(severityStr)
	if len(criteriaRaw) > 0 {
		if err := json.Unmarshal(criteriaRaw, &entry.CPEMatchCriteria); err != nil {
			return nil, fmt.Errorf("unmarshal cpe_match_criteria fallito: %w", err)
		}
	} else {
		entry.CPEMatchCriteria = []domain.CPEMatchCriterion{}
	}

	return &entry, nil
}

// SearchByCPEURI esegue una ricerca ad altissima efficienza sfruttando l'indice GIN (jsonb_path_ops)
// e l'operatore di contenimento JSONB (@>) per un CPE 2.3 URI esatto.
func (r *CVERepo) SearchByCPEURI(ctx context.Context, cpeURI string) ([]*domain.CVEEntry, error) {
	if cpeURI == "" {
		return nil, nil
	}

	// Payload JSON da verificare con operatore @>
	matchFilter := []map[string]interface{}{
		{"cpe23Uri": cpeURI},
	}
	filterJSON, err := json.Marshal(matchFilter)
	if err != nil {
		return nil, fmt.Errorf("serializzazione filtro CPE JSON fallita: %w", err)
	}

	query := `
		SELECT cve_id, cvss_score, severity, description, cpe_match_criteria, published_date, updated_at
		FROM cve_entries
		WHERE cpe_match_criteria @> $1::jsonb
		ORDER BY cvss_score DESC NULLS LAST, cve_id ASC
	`

	rows, err := r.pool.Query(ctx, query, string(filterJSON))
	if err != nil {
		return nil, fmt.Errorf("ricerca CVE per CPE URI fallita: %w", err)
	}
	defer rows.Close()

	return r.scanEntries(rows)
}

// SearchByCPE cerca le CVE che presentano criteri CPE corrispondenti al prefisso specificato
// (es. "cpe:2.3:a:apache:http_server") e alla versione target.
func (r *CVERepo) SearchByCPE(ctx context.Context, cpePrefix string, version string) ([]*domain.CVEEntry, error) {
	if cpePrefix == "" {
		return nil, nil
	}

	pattern := cpePrefix
	if version != "" && !strings.Contains(cpePrefix, version) {
		if !strings.HasSuffix(pattern, ":") {
			pattern += ":"
		}
		pattern += version
	}
	if !strings.HasSuffix(pattern, "%") {
		pattern += "%"
	}

	query := `
		SELECT DISTINCT cve_id, cvss_score, severity, description, cpe_match_criteria, published_date, updated_at
		FROM cve_entries,
		     jsonb_array_elements(cpe_match_criteria) AS elem
		WHERE elem->>'cpe23Uri' ILIKE $1
		ORDER BY cvss_score DESC NULLS LAST, cve_id ASC
	`

	rows, err := r.pool.Query(ctx, query, pattern)
	if err != nil {
		return nil, fmt.Errorf("ricerca CVE per prefisso CPE fallita: %w", err)
	}
	defer rows.Close()

	return r.scanEntries(rows)
}

// Search esegue una ricerca paginata e filtrata nel catalogo CVE
func (r *CVERepo) Search(ctx context.Context, filter domain.CVEFilter) ([]*domain.CVEEntry, int, error) {
	whereClauses := []string{"1=1"}
	args := []interface{}{}
	argIdx := 1

	if strings.TrimSpace(filter.Query) != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("(cve_id ILIKE $%d OR description ILIKE $%d OR cpe_match_criteria::text ILIKE $%d)", argIdx, argIdx, argIdx))
		args = append(args, "%"+strings.TrimSpace(filter.Query)+"%")
		argIdx++
	}

	if filter.Severity != nil && filter.Severity.IsValid() {
		whereClauses = append(whereClauses, fmt.Sprintf("severity = $%d", argIdx))
		args = append(args, string(*filter.Severity))
		argIdx++
	}

	if filter.MinCVSS != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("cvss_score >= $%d", argIdx))
		args = append(args, *filter.MinCVSS)
		argIdx++
	}

	whereSQL := strings.Join(whereClauses, " AND ")

	// Conteggio complessivo elementi corrispondenti
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM cve_entries WHERE %s", whereSQL)
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("conteggio CVE fallito: %w", err)
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	query := fmt.Sprintf(`
		SELECT cve_id, cvss_score, severity, description, cpe_match_criteria, published_date, updated_at
		FROM cve_entries
		WHERE %s
		ORDER BY cvss_score DESC NULLS LAST, cve_id ASC
		LIMIT $%d OFFSET $%d
	`, whereSQL, argIdx, argIdx+1)

	queryArgs := append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("query ricerca CVE fallita: %w", err)
	}
	defer rows.Close()

	entries, err := r.scanEntries(rows)
	if err != nil {
		return nil, 0, err
	}

	return entries, total, nil
}

// UpsertBatch inserisce o aggiorna atomicamente una collezione di voci CVE tramite pgx.Batch (NFR-PERF-03)
func (r *CVERepo) UpsertBatch(ctx context.Context, entries []*domain.CVEEntry) error {
	if len(entries) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	now := time.Now().UTC()

	upsertSQL := `
		INSERT INTO cve_entries (
			cve_id, cvss_score, severity, description, cpe_match_criteria, published_date, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (cve_id) DO UPDATE SET
			cvss_score = EXCLUDED.cvss_score,
			severity = EXCLUDED.severity,
			description = EXCLUDED.description,
			cpe_match_criteria = EXCLUDED.cpe_match_criteria,
			published_date = EXCLUDED.published_date,
			updated_at = EXCLUDED.updated_at
	`

	for _, entry := range entries {
		if entry.Severity == "" {
			if entry.CVSSScore != nil {
				entry.Severity = domain.SeverityFromCVSS(*entry.CVSSScore)
			} else {
				entry.Severity = domain.SeverityInfo
			}
		}

		criteria := entry.CPEMatchCriteria
		if criteria == nil {
			criteria = []domain.CPEMatchCriterion{}
		}

		criteriaJSON, err := json.Marshal(criteria)
		if err != nil {
			return fmt.Errorf("serializzazione criteri per CVE %s fallita: %w", entry.CVEID, err)
		}

		updatedAt := entry.UpdatedAt
		if updatedAt.IsZero() {
			updatedAt = now
		}

		batch.Queue(
			upsertSQL,
			entry.CVEID,
			entry.CVSSScore,
			string(entry.Severity),
			entry.Description,
			criteriaJSON,
			entry.PublishedDate,
			updatedAt,
		)
	}

	br := r.pool.SendBatch(ctx, batch)
	defer br.Close()

	for i := 0; i < len(entries); i++ {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("upsert batch CVE #%d (%s) fallito: %w", i, entries[i].CVEID, err)
		}
	}

	return nil
}

// Count restituisce il numero totale di CVE archiviate nel database locale
func (r *CVERepo) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM cve_entries").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("conteggio CVE fallito: %w", err)
	}
	return count, nil
}

// GetStats estrae le metriche di sintesi del catalogo CVE locale (totali, conteggi per severità e data di aggiornamento)
func (r *CVERepo) GetStats(ctx context.Context) (*domain.CVEStats, error) {
	stats := &domain.CVEStats{
		SeverityCounts: make(map[domain.SeverityLevel]int64),
	}

	// 1. Conteggio complessivo e timestamp dell'ultimo aggiornamento
	var maxUpdated *time.Time
	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*), MAX(updated_at)
		FROM cve_entries
	`).Scan(&stats.TotalCount, &maxUpdated)
	if err != nil {
		return nil, fmt.Errorf("recupero conteggio generale CVE fallito: %w", err)
	}
	stats.LastUpdatedAt = maxUpdated

	// 2. Raggruppamento per livello di severità
	rows, err := r.pool.Query(ctx, `
		SELECT severity, COUNT(*)
		FROM cve_entries
		GROUP BY severity
	`)
	if err != nil {
		return nil, fmt.Errorf("recupero statistiche severità CVE fallito: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var sevStr string
		var count int64
		if err := rows.Scan(&sevStr, &count); err != nil {
			return nil, fmt.Errorf("lettura riga statistica severità fallita: %w", err)
		}
		stats.SeverityCounts[domain.SeverityLevel(sevStr)] = count
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterazione righe statistiche severità fallita: %w", err)
	}

	return stats, nil
}

// scanEntries deserializza le righe risultanti in uno slice di *domain.CVEEntry
func (r *CVERepo) scanEntries(rows pgx.Rows) ([]*domain.CVEEntry, error) {
	var entries []*domain.CVEEntry

	for rows.Next() {
		var entry domain.CVEEntry
		var severityStr string
		var criteriaRaw []byte

		if err := rows.Scan(
			&entry.CVEID,
			&entry.CVSSScore,
			&severityStr,
			&entry.Description,
			&criteriaRaw,
			&entry.PublishedDate,
			&entry.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("lettura record CVE fallita: %w", err)
		}

		entry.Severity = domain.SeverityLevel(severityStr)
		if len(criteriaRaw) > 0 {
			if err := json.Unmarshal(criteriaRaw, &entry.CPEMatchCriteria); err != nil {
				return nil, fmt.Errorf("unmarshal cpe_match_criteria per %s fallito: %w", entry.CVEID, err)
			}
		} else {
			entry.CPEMatchCriteria = []domain.CPEMatchCriterion{}
		}

		entries = append(entries, &entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterazione record CVE fallita: %w", err)
	}

	return entries, nil
}


package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"nextgen-cybersec/internal/domain"
)

// FindingRepo implementa domain.FindingRepository per PostgreSQL
type FindingRepo struct {
	pool *pgxpool.Pool
}

// NewFindingRepo crea una nuova istanza di FindingRepo
func NewFindingRepo(pool *pgxpool.Pool) *FindingRepo {
	return &FindingRepo{pool: pool}
}

// GetByID recupera un singolo finding garantendo l'isolamento tenant
func (r *FindingRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*domain.Finding, error) {
	query := `
		SELECT id, scan_id, tenant_id, host_ip, hostname, port, protocol, port_state,
		       service_name, service_version, cve_id, cvss_score, severity,
		       confidence, detection_method, cpe,
		       raw_banner, remediation, created_at
		FROM findings
		WHERE tenant_id = $1 AND id = $2
	`

	var f domain.Finding
	var severityStr, confStr, detStr string

	err := r.pool.QueryRow(ctx, query, tenantID, id).Scan(
		&f.ID,
		&f.ScanID,
		&f.TenantID,
		&f.HostIP,
		&f.Hostname,
		&f.Port,
		&f.Protocol,
		&f.PortState,
		&f.ServiceName,
		&f.ServiceVersion,
		&f.CVEID,
		&f.CVSSScore,
		&severityStr,
		&confStr,
		&detStr,
		&f.CPE,
		&f.RawBanner,
		&f.Remediation,
		&f.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("recupero finding per ID fallito: %w", err)
	}

	f.Severity = domain.SeverityLevel(severityStr)
	f.Confidence = domain.FindingConfidence(confStr)
	f.DetectionMethod = domain.DetectionMethod(detStr)
	if f.PortState == "" {
		f.PortState = "open"
	}
	return &f, nil
}

// List recupera l'elenco paginato dei finding con filtri per scan, IP, porta e severità
func (r *FindingRepo) List(ctx context.Context, tenantID uuid.UUID, filter domain.FindingFilter) ([]*domain.Finding, int, error) {
	whereClauses := []string{"tenant_id = $1"}
	args := []interface{}{tenantID}
	argIdx := 2

	if filter.ScanID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("scan_id = $%d", argIdx))
		args = append(args, *filter.ScanID)
		argIdx++
	}

	if filter.HostIP != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("host_ip = $%d", argIdx))
		args = append(args, filter.HostIP)
		argIdx++
	}

	if filter.Severity != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("severity = $%d", argIdx))
		args = append(args, string(*filter.Severity))
		argIdx++
	}

	if filter.Port != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("port = $%d", argIdx))
		args = append(args, *filter.Port)
		argIdx++
	}

	if filter.CVEID != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("cve_id = $%d", argIdx))
		args = append(args, filter.CVEID)
		argIdx++
	}

	whereSQL := strings.Join(whereClauses, " AND ")

	// Conteggio complessivo
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM findings WHERE %s", whereSQL)
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("conteggio findings fallito: %w", err)
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
		SELECT id, scan_id, tenant_id, host_ip, hostname, port, protocol, port_state,
		       service_name, service_version, cve_id, cvss_score, severity,
		       confidence, detection_method, cpe,
		       raw_banner, remediation, created_at
		FROM findings
		WHERE %s
		ORDER BY CASE severity
			WHEN 'CRITICAL' THEN 1
			WHEN 'HIGH' THEN 2
			WHEN 'MEDIUM' THEN 3
			WHEN 'LOW' THEN 4
			ELSE 5 END ASC, port ASC, id ASC
		LIMIT $%d OFFSET $%d
	`, whereSQL, argIdx, argIdx+1)

	queryArgs := append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("query findings fallita: %w", err)
	}
	defer rows.Close()

	var findings []*domain.Finding
	for rows.Next() {
		var f domain.Finding
		var severityStr, confStr, detStr string

		if err := rows.Scan(
			&f.ID,
			&f.ScanID,
			&f.TenantID,
			&f.HostIP,
			&f.Hostname,
			&f.Port,
			&f.Protocol,
			&f.PortState,
			&f.ServiceName,
			&f.ServiceVersion,
			&f.CVEID,
			&f.CVSSScore,
			&severityStr,
			&confStr,
			&detStr,
			&f.CPE,
			&f.RawBanner,
			&f.Remediation,
			&f.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("lettura riga finding fallita: %w", err)
		}

		f.Severity = domain.SeverityLevel(severityStr)
		f.Confidence = domain.FindingConfidence(confStr)
		f.DetectionMethod = domain.DetectionMethod(detStr)
		if f.PortState == "" {
			f.PortState = "open"
		}
		findings = append(findings, &f)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterazione findings fallita: %w", err)
	}

	return findings, total, nil
}

// CreateBatch inserisce atomicamente una collezione di findings tramite pgx.Batch (NFR-PERF-03)
func (r *FindingRepo) CreateBatch(ctx context.Context, findings []*domain.Finding) error {
	if len(findings) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	now := time.Now()

	insertSQL := `
		INSERT INTO findings (id, scan_id, tenant_id, host_ip, hostname, port, protocol, port_state,
		                       service_name, service_version, cve_id, cvss_score, severity,
		                       confidence, detection_method, cpe,
		                       raw_banner, remediation, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)
	`

	for _, f := range findings {
		if f.ID == uuid.Nil {
			f.ID = uuid.New()
		}
		if f.CreatedAt.IsZero() {
			f.CreatedAt = now
		}
		if f.Severity == "" {
			f.Severity = domain.SeverityInfo
		}
		if f.Confidence == "" {
			f.Confidence = domain.ConfidenceConfirmed
		}
		if f.DetectionMethod == "" {
			f.DetectionMethod = domain.DetectionMethodBanner
		}
		if f.PortState == "" {
			f.PortState = "open"
		}

		batch.Queue(
			insertSQL,
			f.ID,
			f.ScanID,
			f.TenantID,
			f.HostIP,
			f.Hostname,
			f.Port,
			f.Protocol,
			f.PortState,
			f.ServiceName,
			f.ServiceVersion,
			f.CVEID,
			f.CVSSScore,
			string(f.Severity),
			string(f.Confidence),
			string(f.DetectionMethod),
			f.CPE,
			f.RawBanner,
			f.Remediation,
			f.CreatedAt,
		)
	}

	br := r.pool.SendBatch(ctx, batch)
	defer br.Close()

	for i := 0; i < len(findings); i++ {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("inserimento batch finding #%d fallito: %w", i, err)
		}
	}

	return nil
}

// GetSeverityCountsByScan raggruppa e conta i finding per livello di severità per una specifica scansione
func (r *FindingRepo) GetSeverityCountsByScan(ctx context.Context, tenantID, scanID uuid.UUID) (map[domain.SeverityLevel]int, error) {
	query := `
		SELECT severity, COUNT(*)
		FROM findings
		WHERE tenant_id = $1 AND scan_id = $2
		GROUP BY severity
	`

	rows, err := r.pool.Query(ctx, query, tenantID, scanID)
	if err != nil {
		return nil, fmt.Errorf("conteggio severità fallito: %w", err)
	}
	defer rows.Close()

	counts := make(map[domain.SeverityLevel]int)
	for rows.Next() {
		var sevStr string
		var count int
		if err := rows.Scan(&sevStr, &count); err != nil {
			return nil, fmt.Errorf("lettura conteggio severità fallita: %w", err)
		}
		counts[domain.SeverityLevel(sevStr)] = count
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterazione conteggio severità fallita: %w", err)
	}

	return counts, nil
}


package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"nextgen-cybersec/internal/domain"
)

// ReportRepo implementa domain.ReportRepository per PostgreSQL
type ReportRepo struct {
	pool *pgxpool.Pool
}

// NewReportRepo crea una nuova istanza di ReportRepo
func NewReportRepo(pool *pgxpool.Pool) *ReportRepo {
	return &ReportRepo{pool: pool}
}

// List recupera l'elenco paginato dei report aggregati per un dato tenant applicando i filtri specificati
func (r *ReportRepo) List(ctx context.Context, tenantID uuid.UUID, filter domain.ReportFilter) ([]*domain.ReportItem, int, error) {
	whereClauses := []string{"s.tenant_id = $1"}
	args := []interface{}{tenantID}
	argIdx := 2

	if filter.TargetID != nil && *filter.TargetID != uuid.Nil {
		whereClauses = append(whereClauses, fmt.Sprintf("s.target_id = $%d", argIdx))
		args = append(args, *filter.TargetID)
		argIdx++
	}

	if filter.GroupID != nil && *filter.GroupID != uuid.Nil {
		whereClauses = append(whereClauses, fmt.Sprintf("t.group_id = $%d", argIdx))
		args = append(args, *filter.GroupID)
		argIdx++
	}

	if filter.TargetType != nil && filter.TargetType.IsValid() {
		whereClauses = append(whereClauses, fmt.Sprintf("t.target_type = $%d", argIdx))
		args = append(args, string(*filter.TargetType))
		argIdx++
	}

	if filter.Status != nil && filter.Status.IsValid() {
		whereClauses = append(whereClauses, fmt.Sprintf("s.status = $%d", argIdx))
		args = append(args, string(*filter.Status))
		argIdx++
	}

	if filter.Severity != nil && *filter.Severity != "" {
		caseMaxSeverity := `(CASE 
			WHEN COALESCE((s.summary_stats->>'critical_count')::int, 0) > 0 THEN 'CRITICAL'
			WHEN COALESCE((s.summary_stats->>'high_count')::int, 0) > 0 THEN 'HIGH'
			WHEN COALESCE((s.summary_stats->>'medium_count')::int, 0) > 0 THEN 'MEDIUM'
			WHEN COALESCE((s.summary_stats->>'low_count')::int, 0) > 0 THEN 'LOW'
			ELSE 'INFO'
		END)`
		whereClauses = append(whereClauses, fmt.Sprintf("%s = $%d", caseMaxSeverity, argIdx))
		args = append(args, string(*filter.Severity))
		argIdx++
	}

	if filter.DateFrom != nil && !filter.DateFrom.IsZero() {
		whereClauses = append(whereClauses, fmt.Sprintf("COALESCE(s.completed_at, s.created_at) >= $%d", argIdx))
		args = append(args, *filter.DateFrom)
		argIdx++
	}

	if filter.DateTo != nil && !filter.DateTo.IsZero() {
		whereClauses = append(whereClauses, fmt.Sprintf("COALESCE(s.completed_at, s.created_at) <= $%d", argIdx))
		args = append(args, *filter.DateTo)
		argIdx++
	}

	if strings.TrimSpace(filter.Search) != "" {
		pattern := "%" + strings.TrimSpace(filter.Search) + "%"
		whereClauses = append(whereClauses, fmt.Sprintf("(t.value ILIKE $%d OR COALESCE(t.description, '') ILIKE $%d OR sp.name ILIKE $%d)", argIdx, argIdx, argIdx))
		args = append(args, pattern)
		argIdx++
	}

	whereSQL := strings.Join(whereClauses, " AND ")

	// 1. Conteggio totale record
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM scans s
		JOIN targets t ON s.target_id = t.id
		LEFT JOIN target_groups tg ON t.group_id = tg.id
		JOIN scan_profiles sp ON s.profile_id = sp.id
		WHERE %s
	`, whereSQL)

	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("conteggio report fallito: %w", err)
	}

	// 2. Query paginata
	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	} else if limit > 100 {
		limit = 100
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	selectQuery := fmt.Sprintf(`
		SELECT 
			s.id,
			s.tenant_id,
			s.target_id,
			t.value,
			t.target_type,
			t.group_id,
			tg.name,
			s.profile_id,
			sp.name,
			s.status,
			s.trigger_type,
			s.started_at,
			s.completed_at,
			s.summary_stats,
			s.created_at
		FROM scans s
		JOIN targets t ON s.target_id = t.id
		LEFT JOIN target_groups tg ON t.group_id = tg.id
		JOIN scan_profiles sp ON s.profile_id = sp.id
		WHERE %s
		ORDER BY COALESCE(s.completed_at, s.created_at) DESC
		LIMIT $%d OFFSET $%d
	`, whereSQL, argIdx, argIdx+1)

	queryArgs := append(args, limit, offset)

	rows, err := r.pool.Query(ctx, selectQuery, queryArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("interrogazione storico report fallita: %w", err)
	}
	defer rows.Close()

	var items []*domain.ReportItem
	for rows.Next() {
		var item domain.ReportItem
		var targetTypeStr, statusStr, triggerStr string
		var groupName *string
		var statsRaw []byte

		if err := rows.Scan(
			&item.ScanID,
			&item.TenantID,
			&item.TargetID,
			&item.TargetValue,
			&targetTypeStr,
			&item.GroupID,
			&groupName,
			&item.ProfileID,
			&item.ProfileName,
			&statusStr,
			&triggerStr,
			&item.StartedAt,
			&item.CompletedAt,
			&statsRaw,
			&item.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("lettura riga report fallita: %w", err)
		}

		item.TargetType = domain.TargetType(targetTypeStr)
		item.Status = domain.ScanStatus(statusStr)
		item.TriggerType = domain.ScanTriggerType(triggerStr)
		item.GroupName = groupName

		if len(statsRaw) > 0 {
			_ = json.Unmarshal(statsRaw, &item.SummaryStats)
		}

		item.MaxSeverity = domain.CalculateMaxSeverity(item.SummaryStats)
		score, grade := domain.CalculatePostureScore(item.SummaryStats)
		item.PostureScore = score
		item.PostureGrade = grade

		items = append(items, &item)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterazione righe report fallita: %w", err)
	}

	return items, total, nil
}

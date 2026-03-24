package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gpilot/internal/domain/alert"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AlertRepo struct {
	pool *pgxpool.Pool
}

func NewAlertRepo(pool *pgxpool.Pool) *AlertRepo {
	return &AlertRepo{pool: pool}
}

func (r *AlertRepo) Create(ctx context.Context, a *alert.Alert) error {
	labelsJSON, _ := json.Marshal(a.Labels)
	annotationsJSON, _ := json.Marshal(a.Annotations)

	_, err := r.pool.Exec(ctx, `
		INSERT INTO alerts (id, fingerprint, group_id, status, severity, name, namespace, pod, node,
			labels, annotations, started_at, resolved_at, last_active_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`, a.ID, a.Fingerprint, a.GroupID, a.Status, a.Severity, a.Name, a.Namespace,
		a.Pod, a.Node, labelsJSON, annotationsJSON, a.StartedAt, a.ResolvedAt,
		a.LastActiveAt, a.CreatedAt, a.UpdatedAt)
	return err
}

func (r *AlertRepo) Update(ctx context.Context, a *alert.Alert) error {
	a.UpdatedAt = time.Now()
	labelsJSON, _ := json.Marshal(a.Labels)
	annotationsJSON, _ := json.Marshal(a.Annotations)

	_, err := r.pool.Exec(ctx, `
		UPDATE alerts SET
			group_id = $2, status = $3, severity = $4, adjusted_severity = $5,
			labels = $6, annotations = $7, resolved_at = $8, last_active_at = $9,
			acknowledged_by = $10, acknowledged_at = $11, updated_at = $12
		WHERE id = $1
	`, a.ID, a.GroupID, a.Status, a.Severity, a.AdjustedSeverity,
		labelsJSON, annotationsJSON, a.ResolvedAt, a.LastActiveAt,
		a.AcknowledgedBy, a.AcknowledgedAt, a.UpdatedAt)
	return err
}

func (r *AlertRepo) GetByID(ctx context.Context, id uuid.UUID) (*alert.Alert, error) {
	a := &alert.Alert{}
	var labelsJSON, annotationsJSON []byte

	err := r.pool.QueryRow(ctx, `
		SELECT id, fingerprint, group_id, status, severity, adjusted_severity, name, namespace,
			pod, node, labels, annotations, started_at, resolved_at, last_active_at,
			acknowledged_by, acknowledged_at, created_at, updated_at
		FROM alerts WHERE id = $1
	`, id).Scan(
		&a.ID, &a.Fingerprint, &a.GroupID, &a.Status, &a.Severity, &a.AdjustedSeverity,
		&a.Name, &a.Namespace, &a.Pod, &a.Node, &labelsJSON, &annotationsJSON,
		&a.StartedAt, &a.ResolvedAt, &a.LastActiveAt, &a.AcknowledgedBy,
		&a.AcknowledgedAt, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	_ = json.Unmarshal(labelsJSON, &a.Labels)
	_ = json.Unmarshal(annotationsJSON, &a.Annotations)
	return a, nil
}

func (r *AlertRepo) GetByFingerprint(ctx context.Context, fingerprint string) (*alert.Alert, error) {
	a := &alert.Alert{}
	var labelsJSON, annotationsJSON []byte

	err := r.pool.QueryRow(ctx, `
		SELECT id, fingerprint, group_id, status, severity, adjusted_severity, name, namespace,
			pod, node, labels, annotations, started_at, resolved_at, last_active_at,
			acknowledged_by, acknowledged_at, created_at, updated_at
		FROM alerts WHERE fingerprint = $1
	`, fingerprint).Scan(
		&a.ID, &a.Fingerprint, &a.GroupID, &a.Status, &a.Severity, &a.AdjustedSeverity,
		&a.Name, &a.Namespace, &a.Pod, &a.Node, &labelsJSON, &annotationsJSON,
		&a.StartedAt, &a.ResolvedAt, &a.LastActiveAt, &a.AcknowledgedBy,
		&a.AcknowledgedAt, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	_ = json.Unmarshal(labelsJSON, &a.Labels)
	_ = json.Unmarshal(annotationsJSON, &a.Annotations)
	return a, nil
}

func (r *AlertRepo) List(ctx context.Context, q alert.AlertListQuery) (*alert.AlertListResult, error) {
	var conditions []string
	var args []interface{}
	argIdx := 1

	if q.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, *q.Status)
		argIdx++
	}
	if q.Severity != nil {
		conditions = append(conditions, fmt.Sprintf("severity = $%d", argIdx))
		args = append(args, *q.Severity)
		argIdx++
	}
	if q.Namespace != "" {
		conditions = append(conditions, fmt.Sprintf("namespace = $%d", argIdx))
		args = append(args, q.Namespace)
		argIdx++
	}
	if q.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(name ILIKE $%d OR namespace ILIKE $%d)", argIdx, argIdx))
		args = append(args, "%"+q.Search+"%")
		argIdx++
	}
	if q.From != nil {
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", argIdx))
		args = append(args, *q.From)
		argIdx++
	}
	if q.To != nil {
		conditions = append(conditions, fmt.Sprintf("created_at <= $%d", argIdx))
		args = append(args, *q.To)
		argIdx++
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count total
	var total int64
	countQuery := "SELECT COUNT(*) FROM alerts " + where
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, err
	}

	// Pagination
	if q.Page < 1 {
		q.Page = 1
	}
	if q.Size < 1 || q.Size > 100 {
		q.Size = 20
	}
	offset := (q.Page - 1) * q.Size

	query := fmt.Sprintf(`
		SELECT id, fingerprint, group_id, status, severity, adjusted_severity, name, namespace,
			pod, node, labels, annotations, started_at, resolved_at, last_active_at,
			acknowledged_by, acknowledged_at, created_at, updated_at
		FROM alerts %s
		ORDER BY created_at DESC
		LIMIT %d OFFSET %d
	`, where, q.Size, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alerts []alert.Alert
	for rows.Next() {
		a := alert.Alert{}
		var labelsJSON, annotationsJSON []byte
		if err := rows.Scan(
			&a.ID, &a.Fingerprint, &a.GroupID, &a.Status, &a.Severity, &a.AdjustedSeverity,
			&a.Name, &a.Namespace, &a.Pod, &a.Node, &labelsJSON, &annotationsJSON,
			&a.StartedAt, &a.ResolvedAt, &a.LastActiveAt, &a.AcknowledgedBy,
			&a.AcknowledgedAt, &a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(labelsJSON, &a.Labels)
		_ = json.Unmarshal(annotationsJSON, &a.Annotations)
		alerts = append(alerts, a)
	}

	return &alert.AlertListResult{
		Alerts: alerts,
		Total:  total,
		Page:   q.Page,
		Size:   q.Size,
	}, nil
}

func (r *AlertRepo) Acknowledge(ctx context.Context, id uuid.UUID, user string) error {
	now := time.Now()
	_, err := r.pool.Exec(ctx, `
		UPDATE alerts SET status = $2, acknowledged_by = $3, acknowledged_at = $4, updated_at = $5
		WHERE id = $1
	`, id, alert.StatusAcknowledged, user, now, now)
	return err
}

// AlertGroupRepo implements alert.GroupRepository
type AlertGroupRepo struct {
	pool *pgxpool.Pool
}

func NewAlertGroupRepo(pool *pgxpool.Pool) *AlertGroupRepo {
	return &AlertGroupRepo{pool: pool}
}

func (r *AlertGroupRepo) Create(ctx context.Context, g *alert.AlertGroup) error {
	g.ID = uuid.New()
	g.CreatedAt = time.Now()
	g.UpdatedAt = time.Now()

	_, err := r.pool.Exec(ctx, `
		INSERT INTO alert_groups (id, name, namespace, status, alert_count, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, g.ID, g.Name, g.Namespace, g.Status, g.AlertCount, g.CreatedAt, g.UpdatedAt)
	return err
}

func (r *AlertGroupRepo) Update(ctx context.Context, g *alert.AlertGroup) error {
	g.UpdatedAt = time.Now()
	_, err := r.pool.Exec(ctx, `
		UPDATE alert_groups SET name = $2, status = $3, alert_count = $4, root_cause_id = $5, updated_at = $6
		WHERE id = $1
	`, g.ID, g.Name, g.Status, g.AlertCount, g.RootCauseID, g.UpdatedAt)
	return err
}

func (r *AlertGroupRepo) GetByID(ctx context.Context, id uuid.UUID) (*alert.AlertGroup, error) {
	g := &alert.AlertGroup{}
	err := r.pool.QueryRow(ctx, `
		SELECT id, name, namespace, status, alert_count, root_cause_id, created_at, updated_at
		FROM alert_groups WHERE id = $1
	`, id).Scan(&g.ID, &g.Name, &g.Namespace, &g.Status, &g.AlertCount,
		&g.RootCauseID, &g.CreatedAt, &g.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return g, nil
}

func (r *AlertGroupRepo) FindActiveByNamespace(ctx context.Context, namespace string) (*alert.AlertGroup, error) {
	g := &alert.AlertGroup{}
	err := r.pool.QueryRow(ctx, `
		SELECT id, name, namespace, status, alert_count, root_cause_id, created_at, updated_at
		FROM alert_groups
		WHERE namespace = $1 AND status = 'firing'
			AND updated_at > NOW() - INTERVAL '10 minutes'
		ORDER BY updated_at DESC LIMIT 1
	`, namespace).Scan(&g.ID, &g.Name, &g.Namespace, &g.Status, &g.AlertCount,
		&g.RootCauseID, &g.CreatedAt, &g.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return g, nil
}

func (r *AlertGroupRepo) List(ctx context.Context, page, size int) ([]alert.AlertGroup, int64, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}

	var total int64
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM alert_groups").Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * size
	rows, err := r.pool.Query(ctx, `
		SELECT id, name, namespace, status, alert_count, root_cause_id, created_at, updated_at
		FROM alert_groups ORDER BY updated_at DESC LIMIT $1 OFFSET $2
	`, size, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var groups []alert.AlertGroup
	for rows.Next() {
		g := alert.AlertGroup{}
		if err := rows.Scan(&g.ID, &g.Name, &g.Namespace, &g.Status, &g.AlertCount,
			&g.RootCauseID, &g.CreatedAt, &g.UpdatedAt); err != nil {
			return nil, 0, err
		}
		groups = append(groups, g)
	}
	return groups, total, nil
}

func (r *AlertGroupRepo) IncrementAlertCount(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE alert_groups SET alert_count = alert_count + 1, updated_at = NOW() WHERE id = $1
	`, id)
	return err
}

package postgres

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"gpilot/internal/domain/event"
)

type EventRepo struct {
	pool *pgxpool.Pool
}

func NewEventRepo(pool *pgxpool.Pool) *EventRepo {
	return &EventRepo{pool: pool}
}

func (r *EventRepo) Upsert(ctx context.Context, e *event.K8sEvent) error {
	involvedJSON, _ := json.Marshal(e.InvolvedObject)

	_, err := r.pool.Exec(ctx, `
		INSERT INTO k8s_events (id, uid, type, reason, message, namespace, involved_object,
			first_seen, last_seen, count, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (uid) DO UPDATE SET
			last_seen = EXCLUDED.last_seen,
			count = EXCLUDED.count,
			message = EXCLUDED.message
	`, e.ID, e.UID, e.Type, e.Reason, e.Message, e.Namespace, involvedJSON,
		e.FirstSeen, e.LastSeen, e.Count, e.CreatedAt)
	return err
}

func (r *EventRepo) GetByID(ctx context.Context, id uuid.UUID) (*event.K8sEvent, error) {
	e := &event.K8sEvent{}
	var involvedJSON []byte

	err := r.pool.QueryRow(ctx, `
		SELECT id, uid, type, reason, message, namespace, involved_object,
			first_seen, last_seen, count, created_at
		FROM k8s_events WHERE id = $1
	`, id).Scan(&e.ID, &e.UID, &e.Type, &e.Reason, &e.Message, &e.Namespace,
		&involvedJSON, &e.FirstSeen, &e.LastSeen, &e.Count, &e.CreatedAt)
	if err != nil {
		return nil, err
	}

	_ = json.Unmarshal(involvedJSON, &e.InvolvedObject)
	return e, nil
}

func (r *EventRepo) List(ctx context.Context, q event.EventQuery) ([]event.K8sEvent, int64, error) {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.Size < 1 || q.Size > 100 {
		q.Size = 20
	}

	// Simple query for MVP
	var total int64
	_ = r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM k8s_events").Scan(&total)

	offset := (q.Page - 1) * q.Size
	rows, err := r.pool.Query(ctx, `
		SELECT id, uid, type, reason, message, namespace, involved_object,
			first_seen, last_seen, count, created_at
		FROM k8s_events ORDER BY last_seen DESC LIMIT $1 OFFSET $2
	`, q.Size, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var events []event.K8sEvent
	for rows.Next() {
		e := event.K8sEvent{}
		var involvedJSON []byte
		if err := rows.Scan(&e.ID, &e.UID, &e.Type, &e.Reason, &e.Message, &e.Namespace,
			&involvedJSON, &e.FirstSeen, &e.LastSeen, &e.Count, &e.CreatedAt); err != nil {
			return nil, 0, err
		}
		_ = json.Unmarshal(involvedJSON, &e.InvolvedObject)
		events = append(events, e)
	}

	return events, total, nil
}

func (r *EventRepo) ListByNamespace(ctx context.Context, namespace string, limit int) ([]event.K8sEvent, error) {
	if limit <= 0 {
		limit = 50
	}

	cutoff := time.Now().Add(-24 * time.Hour)
	rows, err := r.pool.Query(ctx, `
		SELECT id, uid, type, reason, message, namespace, involved_object,
			first_seen, last_seen, count, created_at
		FROM k8s_events WHERE namespace = $1 AND last_seen > $2
		ORDER BY last_seen DESC LIMIT $3
	`, namespace, cutoff, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []event.K8sEvent
	for rows.Next() {
		e := event.K8sEvent{}
		var involvedJSON []byte
		if err := rows.Scan(&e.ID, &e.UID, &e.Type, &e.Reason, &e.Message, &e.Namespace,
			&involvedJSON, &e.FirstSeen, &e.LastSeen, &e.Count, &e.CreatedAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(involvedJSON, &e.InvolvedObject)
		events = append(events, e)
	}

	return events, nil
}

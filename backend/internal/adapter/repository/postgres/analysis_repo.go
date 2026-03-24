package postgres

import (
	"context"
	"encoding/json"
	"time"

	"gpilot/internal/domain/analysis"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AnalysisRepo struct {
	pool *pgxpool.Pool
}

func NewAnalysisRepo(pool *pgxpool.Pool) *AnalysisRepo {
	return &AnalysisRepo{pool: pool}
}

func (r *AnalysisRepo) Create(ctx context.Context, a *analysis.Analysis) error {
	a.ID = uuid.New()
	a.CreatedAt = time.Now()
	suggestionsJSON, _ := json.Marshal(a.Suggestions)
	contextJSON, _ := json.Marshal(a.ContextSnapshot)

	_, err := r.pool.Exec(ctx, `
		INSERT INTO analyses (id, alert_id, group_id, type, summary, root_cause, suggestions,
			severity_suggestion, context_snapshot, llm_model, llm_tokens_used, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`, a.ID, a.AlertID, a.GroupID, a.Type, a.Summary, a.RootCause, suggestionsJSON,
		a.SeveritySuggestion, contextJSON, a.LLMModel, a.LLMTokensUsed, a.CreatedAt)
	return err
}

func (r *AnalysisRepo) GetByID(ctx context.Context, id uuid.UUID) (*analysis.Analysis, error) {
	a := &analysis.Analysis{}
	var suggestionsJSON, contextJSON []byte

	err := r.pool.QueryRow(ctx, `
		SELECT id, alert_id, group_id, type, summary, root_cause, suggestions,
			severity_suggestion, context_snapshot, llm_model, llm_tokens_used, created_at
		FROM analyses WHERE id = $1
	`, id).Scan(&a.ID, &a.AlertID, &a.GroupID, &a.Type, &a.Summary, &a.RootCause,
		&suggestionsJSON, &a.SeveritySuggestion, &contextJSON, &a.LLMModel,
		&a.LLMTokensUsed, &a.CreatedAt)
	if err != nil {
		return nil, err
	}

	_ = json.Unmarshal(suggestionsJSON, &a.Suggestions)
	_ = json.Unmarshal(contextJSON, &a.ContextSnapshot)
	return a, nil
}

func (r *AnalysisRepo) ListByAlertID(ctx context.Context, alertID uuid.UUID) ([]analysis.Analysis, error) {
	return r.listBy(ctx, "alert_id = $1", alertID)
}

func (r *AnalysisRepo) ListByGroupID(ctx context.Context, groupID uuid.UUID) ([]analysis.Analysis, error) {
	return r.listBy(ctx, "group_id = $1", groupID)
}

func (r *AnalysisRepo) ListRecent(ctx context.Context, limit int) ([]analysis.Analysis, error) {
	if limit <= 0 {
		limit = 10
	}
	rows, err := r.pool.Query(ctx, `
		SELECT id, alert_id, group_id, type, summary, root_cause, suggestions,
			severity_suggestion, context_snapshot, llm_model, llm_tokens_used, created_at
		FROM analyses ORDER BY created_at DESC LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanRows(rows)
}

func (r *AnalysisRepo) listBy(ctx context.Context, condition string, arg interface{}) ([]analysis.Analysis, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, alert_id, group_id, type, summary, root_cause, suggestions,
			severity_suggestion, context_snapshot, llm_model, llm_tokens_used, created_at
		FROM analyses WHERE `+condition+` ORDER BY created_at DESC
	`, arg)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanRows(rows)
}

func (r *AnalysisRepo) scanRows(rows interface {
	Next() bool
	Scan(...interface{}) error
}) ([]analysis.Analysis, error) {
	var results []analysis.Analysis
	for rows.Next() {
		a := analysis.Analysis{}
		var suggestionsJSON, contextJSON []byte
		if err := rows.Scan(&a.ID, &a.AlertID, &a.GroupID, &a.Type, &a.Summary, &a.RootCause,
			&suggestionsJSON, &a.SeveritySuggestion, &contextJSON, &a.LLMModel,
			&a.LLMTokensUsed, &a.CreatedAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(suggestionsJSON, &a.Suggestions)
		_ = json.Unmarshal(contextJSON, &a.ContextSnapshot)
		results = append(results, a)
	}
	return results, nil
}

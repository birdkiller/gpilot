package analysis

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type AnalysisType string

const (
	TypeRootCause      AnalysisType = "root_cause"
	TypeLogAnalysis    AnalysisType = "log_analysis"
	TypeEventDiagnosis AnalysisType = "event_diagnosis"
)

type Suggestion struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Commands    []string `json:"commands,omitempty"`
}

type Analysis struct {
	ID                 uuid.UUID    `json:"id" db:"id"`
	AlertID            *uuid.UUID   `json:"alert_id,omitempty" db:"alert_id"`
	GroupID            *uuid.UUID   `json:"group_id,omitempty" db:"group_id"`
	Type               AnalysisType `json:"type" db:"type"`
	Summary            string       `json:"summary" db:"summary"`
	RootCause          string       `json:"root_cause" db:"root_cause"`
	Suggestions        []Suggestion `json:"suggestions" db:"suggestions"`
	SeveritySuggestion string       `json:"severity_suggestion,omitempty" db:"severity_suggestion"`
	ContextSnapshot    interface{}  `json:"context_snapshot,omitempty" db:"context_snapshot"`
	LLMModel           string       `json:"llm_model" db:"llm_model"`
	LLMTokensUsed      int          `json:"llm_tokens_used" db:"llm_tokens_used"`
	CreatedAt          time.Time    `json:"created_at" db:"created_at"`
}

// Repository defines the persistence interface for analyses
type Repository interface {
	Create(ctx context.Context, a *Analysis) error
	GetByID(ctx context.Context, id uuid.UUID) (*Analysis, error)
	ListByAlertID(ctx context.Context, alertID uuid.UUID) ([]Analysis, error)
	ListByGroupID(ctx context.Context, groupID uuid.UUID) ([]Analysis, error)
	ListRecent(ctx context.Context, limit int) ([]Analysis, error)
}

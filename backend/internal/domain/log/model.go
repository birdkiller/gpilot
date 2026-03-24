package log

import (
	"context"
	"time"
)

type LogEntry struct {
	Timestamp time.Time         `json:"timestamp"`
	Line      string            `json:"line"`
	Labels    map[string]string `json:"labels,omitempty"`
	Stream    string            `json:"stream,omitempty"`
}

type LogQuery struct {
	Source    string    `json:"source"` // "loki" or "elasticsearch"
	Query     string    `json:"query"`  // LogQL or ES query
	From      time.Time `json:"from"`
	To        time.Time `json:"to"`
	Limit     int       `json:"limit"`
	Direction string    `json:"direction"` // "forward" or "backward"
}

type LogQueryResult struct {
	Entries   []LogEntry `json:"entries"`
	Total     int        `json:"total"`
	QueryExpr string     `json:"query_expr"` // The actual query executed
}

type NaturalQueryRequest struct {
	Question string `json:"question"`
}

type NaturalQueryResult struct {
	TranslatedQuery string         `json:"translated_query"`
	Explanation     string         `json:"explanation"`
	Result          LogQueryResult `json:"result"`
}

type LogAnalyzeRequest struct {
	Logs    []string `json:"logs"`
	Context string   `json:"context,omitempty"`
}

type AnomalyResult struct {
	Namespace   string    `json:"namespace"`
	Description string    `json:"description"`
	Severity    string    `json:"severity"`
	DetectedAt  time.Time `json:"detected_at"`
	LogCount    int       `json:"log_count"`
	Baseline    float64   `json:"baseline"`
	Current     float64   `json:"current"`
}

// LogService defines the interface for log operations
type LogService interface {
	Query(ctx context.Context, q LogQuery) (*LogQueryResult, error)
	NaturalQuery(ctx context.Context, question string) (*NaturalQueryResult, error)
	GetAnomalies(ctx context.Context) ([]AnomalyResult, error)
}

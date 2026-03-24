package analysis

import (
	"context"
)

// AnalysisService defines the interface for running AI analyses
type AnalysisService interface {
	AnalyzeAlert(ctx context.Context, alertID string, streamFn func(chunk string)) (*Analysis, error)
	AnalyzeAlertGroup(ctx context.Context, groupID string, streamFn func(chunk string)) (*Analysis, error)
	AnalyzeLogs(ctx context.Context, logs []string, contextInfo string, streamFn func(chunk string)) (*Analysis, error)
}

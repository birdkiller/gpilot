package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"gpilot/internal/adapter/llm"
	"gpilot/internal/domain/alert"
	"gpilot/internal/domain/analysis"
	"gpilot/internal/domain/event"
	"gpilot/internal/infra/logger"
	ws "gpilot/internal/infra/websocket"
)

type AnalysisApp struct {
	analysisRepo analysis.Repository
	alertRepo    alert.Repository
	groupRepo    alert.GroupRepository
	eventRepo    event.Repository
	llmClient    *llm.Client
	hub          *ws.Hub
}

func NewAnalysisApp(
	analysisRepo analysis.Repository,
	alertRepo alert.Repository,
	groupRepo alert.GroupRepository,
	eventRepo event.Repository,
	llmClient *llm.Client,
	hub *ws.Hub,
) *AnalysisApp {
	return &AnalysisApp{
		analysisRepo: analysisRepo,
		alertRepo:    alertRepo,
		groupRepo:    groupRepo,
		eventRepo:    eventRepo,
		llmClient:    llmClient,
		hub:          hub,
	}
}

func (a *AnalysisApp) AnalyzeAlert(ctx context.Context, alertID string, streamFn func(chunk string)) (*analysis.Analysis, error) {
	uid, err := parseUUID(alertID)
	if err != nil {
		return nil, err
	}

	al, err := a.alertRepo.GetByID(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("alert not found: %w", err)
	}

	// Broadcast analysis started
	a.hub.BroadcastMessage(ws.Message{
		Type:    "analysis:started",
		Payload: map[string]string{"alert_id": alertID},
	})

	// Collect context
	var relatedAlerts []string
	var k8sEvents []string

	if al.GroupID != nil {
		// Get related alerts from same group
		result, err := a.alertRepo.List(ctx, alert.AlertListQuery{
			Namespace: al.Namespace,
			Page:      1,
			Size:      10,
		})
		if err == nil {
			for _, ra := range result.Alerts {
				if ra.ID != al.ID {
					relatedAlerts = append(relatedAlerts, fmt.Sprintf("[%s] %s (%s)", ra.Severity, ra.Name, ra.Status))
				}
			}
		}
	}

	// Get K8s events
	if a.eventRepo != nil && al.Namespace != "" {
		events, err := a.eventRepo.ListByNamespace(ctx, al.Namespace, 10)
		if err == nil {
			for _, e := range events {
				k8sEvents = append(k8sEvents, fmt.Sprintf("[%s] %s: %s", e.Type, e.Reason, e.Message))
			}
		}
	}

	description := al.Annotations["description"]
	if description == "" {
		description = al.Annotations["summary"]
	}

	userPrompt := llm.BuildRootCausePrompt(
		al.Name, string(al.Severity), description, al.Labels,
		relatedAlerts, nil, k8sEvents,
	)

	// Wrap streamFn to also broadcast via WebSocket
	wsStreamFn := func(chunk string) {
		if streamFn != nil {
			streamFn(chunk)
		}
		a.hub.BroadcastMessage(ws.Message{
			Type: "analysis:chunk",
			Payload: map[string]string{
				"alert_id": alertID,
				"chunk":    chunk,
			},
		})
	}

	fullResponse, tokens, err := a.llmClient.ChatStream(ctx, llm.SystemPromptRootCause, userPrompt, wsStreamFn)
	if err != nil {
		return nil, fmt.Errorf("llm analysis failed: %w", err)
	}

	// Parse response to extract structured data
	result := &analysis.Analysis{
		AlertID:   &uid,
		GroupID:   al.GroupID,
		Type:      analysis.TypeRootCause,
		Summary:   extractSection(fullResponse, "根因分析"),
		RootCause: extractSection(fullResponse, "根因分析"),
		Suggestions: []analysis.Suggestion{
			{
				Title:       "修复建议",
				Description: extractSection(fullResponse, "修复建议"),
			},
		},
		SeveritySuggestion: extractSection(fullResponse, "建议严重级别"),
		ContextSnapshot: map[string]interface{}{
			"related_alerts": relatedAlerts,
			"k8s_events":     k8sEvents,
		},
		LLMModel:      a.llmClient.ModelName(),
		LLMTokensUsed: tokens,
	}

	if err := a.analysisRepo.Create(ctx, result); err != nil {
		logger.L.Errorw("failed to save analysis", "error", err)
	}

	// Broadcast analysis done
	a.hub.BroadcastMessage(ws.Message{
		Type:    "analysis:done",
		Payload: result,
	})

	return result, nil
}

func (a *AnalysisApp) AnalyzeLogs(ctx context.Context, logs []string, contextInfo string, streamFn func(chunk string)) (*analysis.Analysis, error) {
	userPrompt := llm.BuildLogAnalysisPrompt(logs, contextInfo)

	wsStreamFn := func(chunk string) {
		if streamFn != nil {
			streamFn(chunk)
		}
		a.hub.BroadcastMessage(ws.Message{
			Type:    "analysis:chunk",
			Payload: map[string]string{"chunk": chunk},
		})
	}

	fullResponse, tokens, err := a.llmClient.ChatStream(ctx, llm.SystemPromptLogAnalysis, userPrompt, wsStreamFn)
	if err != nil {
		return nil, fmt.Errorf("log analysis failed: %w", err)
	}

	result := &analysis.Analysis{
		Type:    analysis.TypeLogAnalysis,
		Summary: extractSection(fullResponse, "错误摘要"),
		RootCause: extractSection(fullResponse, "根因推断"),
		Suggestions: []analysis.Suggestion{
			{
				Title:       "修复建议",
				Description: extractSection(fullResponse, "修复建议"),
			},
		},
		LLMModel:      a.llmClient.ModelName(),
		LLMTokensUsed: tokens,
	}

	if err := a.analysisRepo.Create(ctx, result); err != nil {
		logger.L.Errorw("failed to save log analysis", "error", err)
	}

	return result, nil
}

func (a *AnalysisApp) GetAnalysis(ctx context.Context, id string) (*analysis.Analysis, error) {
	uid, err := parseUUID(id)
	if err != nil {
		return nil, err
	}
	return a.analysisRepo.GetByID(ctx, uid)
}

func (a *AnalysisApp) ListByAlert(ctx context.Context, alertID string) ([]analysis.Analysis, error) {
	uid, err := parseUUID(alertID)
	if err != nil {
		return nil, err
	}
	return a.analysisRepo.ListByAlertID(ctx, uid)
}

func (a *AnalysisApp) ListRecent(ctx context.Context, limit int) ([]analysis.Analysis, error) {
	return a.analysisRepo.ListRecent(ctx, limit)
}

func extractSection(text, header string) string {
	lines := strings.Split(text, "\n")
	var result []string
	capturing := false

	for _, line := range lines {
		if strings.Contains(line, header) {
			capturing = true
			continue
		}
		if capturing && strings.HasPrefix(line, "## ") {
			break
		}
		if capturing {
			result = append(result, line)
		}
	}

	return strings.TrimSpace(strings.Join(result, "\n"))
}

func parseUUID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}

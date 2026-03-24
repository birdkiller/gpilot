package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gpilot/internal/adapter/datasource/loki"
	"gpilot/internal/adapter/llm"
	logdomain "gpilot/internal/domain/log"
	"gpilot/internal/infra/logger"
)

type LogApp struct {
	lokiClient *loki.Client
	llmClient  *llm.Client
}

func NewLogApp(lokiClient *loki.Client, llmClient *llm.Client) *LogApp {
	return &LogApp{
		lokiClient: lokiClient,
		llmClient:  llmClient,
	}
}

func (a *LogApp) Query(ctx context.Context, q logdomain.LogQuery) (*logdomain.LogQueryResult, error) {
	if q.Limit <= 0 {
		q.Limit = 100
	}
	if q.From.IsZero() {
		q.From = time.Now().Add(-1 * time.Hour)
	}
	if q.To.IsZero() {
		q.To = time.Now()
	}

	return a.lokiClient.Query(ctx, q)
}

func (a *LogApp) NaturalQuery(ctx context.Context, question string) (*logdomain.NaturalQueryResult, error) {
	userPrompt := llm.BuildNLQueryPrompt(question)
	response, _, err := a.llmClient.Chat(ctx, llm.SystemPromptNLToLogQL, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("nl query translation: %w", err)
	}

	// Extract LogQL from response
	logQL := extractLogQL(response)
	explanation := extractSection(response, "查询说明")

	if logQL == "" {
		return &logdomain.NaturalQueryResult{
			TranslatedQuery: "",
			Explanation:     response,
		}, nil
	}

	// Execute the translated query
	q := logdomain.LogQuery{
		Query:     logQL,
		From:      time.Now().Add(-1 * time.Hour),
		To:        time.Now(),
		Limit:     100,
		Direction: "backward",
	}

	result, err := a.lokiClient.Query(ctx, q)
	if err != nil {
		logger.L.Warnw("translated logql query failed", "query", logQL, "error", err)
		return &logdomain.NaturalQueryResult{
			TranslatedQuery: logQL,
			Explanation:     explanation,
		}, nil
	}

	return &logdomain.NaturalQueryResult{
		TranslatedQuery: logQL,
		Explanation:     explanation,
		Result:          *result,
	}, nil
}

func extractLogQL(text string) string {
	lines := strings.Split(text, "\n")
	inCodeBlock := false
	var logQL []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			if inCodeBlock {
				break
			}
			inCodeBlock = true
			continue
		}
		if inCodeBlock {
			logQL = append(logQL, line)
		}
	}

	return strings.TrimSpace(strings.Join(logQL, "\n"))
}

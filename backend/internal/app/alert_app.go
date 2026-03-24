package app

import (
	"context"

	"gpilot/internal/adapter/llm"
	"gpilot/internal/domain/alert"
	"gpilot/internal/domain/analysis"
	"gpilot/internal/infra/logger"
	ws "gpilot/internal/infra/websocket"
)

type AlertApp struct {
	pipeline     *alert.Pipeline
	repo         alert.Repository
	groupRepo    alert.GroupRepository
	analysisRepo analysis.Repository
	llmClient    *llm.Client
	hub          *ws.Hub
}

func NewAlertApp(
	pipeline *alert.Pipeline,
	repo alert.Repository,
	groupRepo alert.GroupRepository,
	analysisRepo analysis.Repository,
	llmClient *llm.Client,
	hub *ws.Hub,
) *AlertApp {
	return &AlertApp{
		pipeline:     pipeline,
		repo:         repo,
		groupRepo:    groupRepo,
		analysisRepo: analysisRepo,
		llmClient:    llmClient,
		hub:          hub,
	}
}

func (a *AlertApp) IngestAlerts(ctx context.Context, alerts []*alert.Alert) error {
	processed, err := a.pipeline.Run(ctx, alerts)
	if err != nil {
		return err
	}

	// Notify via WebSocket
	for _, al := range processed {
		msgType := "alert:new"
		if al.Status == alert.StatusResolved {
			msgType = "alert:resolved"
		}
		a.hub.BroadcastMessage(ws.Message{
			Type:    msgType,
			Payload: al,
		})
	}

	logger.L.Infow("alerts ingested", "count", len(processed))
	return nil
}

func (a *AlertApp) ListAlerts(ctx context.Context, q alert.AlertListQuery) (*alert.AlertListResult, error) {
	return a.repo.List(ctx, q)
}

func (a *AlertApp) GetAlert(ctx context.Context, id string) (*alert.Alert, error) {
	uid, err := parseUUID(id)
	if err != nil {
		return nil, err
	}
	return a.repo.GetByID(ctx, uid)
}

func (a *AlertApp) AcknowledgeAlert(ctx context.Context, id string, user string) error {
	uid, err := parseUUID(id)
	if err != nil {
		return err
	}
	if err := a.repo.Acknowledge(ctx, uid, user); err != nil {
		return err
	}

	al, _ := a.repo.GetByID(ctx, uid)
	if al != nil {
		a.hub.BroadcastMessage(ws.Message{
			Type:    "alert:updated",
			Payload: al,
		})
	}
	return nil
}

func (a *AlertApp) ListGroups(ctx context.Context, page, size int) ([]alert.AlertGroup, int64, error) {
	return a.groupRepo.List(ctx, page, size)
}

func (a *AlertApp) GetGroup(ctx context.Context, id string) (*alert.AlertGroup, error) {
	uid, err := parseUUID(id)
	if err != nil {
		return nil, err
	}
	return a.groupRepo.GetByID(ctx, uid)
}

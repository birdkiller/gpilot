package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gpilot/internal/app"
	"gpilot/internal/domain/alert"
	"gpilot/internal/infra/logger"
)

type WebhookHandler struct {
	alertApp *app.AlertApp
}

func NewWebhookHandler(alertApp *app.AlertApp) *WebhookHandler {
	return &WebhookHandler{alertApp: alertApp}
}

func (h *WebhookHandler) HandleAlertmanager(c *gin.Context) {
	var payload alert.AlertmanagerPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload: " + err.Error()})
		return
	}

	logger.L.Infow("received alertmanager webhook",
		"alert_count", len(payload.Alerts),
		"group_key", payload.GroupKey,
		"status", payload.Status,
	)

	var alerts []*alert.Alert
	for _, a := range payload.Alerts {
		domainAlert := alert.FromAlertmanager(a)
		alerts = append(alerts, &domainAlert)
	}

	if err := h.alertApp.IngestAlerts(c.Request.Context(), alerts); err != nil {
		logger.L.Errorw("failed to ingest alerts", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process alerts"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok", "processed": len(alerts)})
}

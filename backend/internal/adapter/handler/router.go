package handler

import (
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gpilot/internal/app"
	ws "gpilot/internal/infra/websocket"
)

func NewRouter(
	alertApp *app.AlertApp,
	analysisApp *app.AnalysisApp,
	logApp *app.LogApp,
	dashboardApp *app.DashboardApp,
	hub *ws.Hub,
) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(LoggerMiddleware())
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "gpilot"})
	})

	// WebSocket
	wsHandler := NewWSHandler(hub)
	r.GET("/ws/alerts", wsHandler.Handle)

	// API v1
	v1 := r.Group("/api/v1")
	{
		// Webhook
		webhookH := NewWebhookHandler(alertApp)
		v1.POST("/webhook/alertmanager", webhookH.HandleAlertmanager)

		// Alerts
		alertH := NewAlertHandler(alertApp)
		v1.GET("/alerts", alertH.List)
		v1.GET("/alerts/:id", alertH.Get)
		v1.PUT("/alerts/:id/acknowledge", alertH.Acknowledge)

		// Alert Groups
		v1.GET("/alert-groups", alertH.ListGroups)
		v1.GET("/alert-groups/:id", alertH.GetGroup)

		// Analysis
		analysisH := NewAnalysisHandler(analysisApp)
		v1.POST("/alerts/:id/analyze", analysisH.AnalyzeAlert)
		v1.GET("/alerts/:id/analyses", analysisH.ListByAlert)
		v1.GET("/analyses/:id", analysisH.Get)

		// Logs
		logH := NewLogHandler(logApp, analysisApp)
		v1.POST("/logs/query", logH.Query)
		v1.POST("/logs/natural-query", logH.NaturalQuery)
		v1.POST("/logs/analyze", logH.Analyze)

		// Dashboard
		dashH := NewDashboardHandler(dashboardApp)
		v1.GET("/dashboard/overview", dashH.Overview)
	}

	return r
}

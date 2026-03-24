package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gpilot/internal/app"
	logdomain "gpilot/internal/domain/log"
)

type LogHandler struct {
	logApp      *app.LogApp
	analysisApp *app.AnalysisApp
}

func NewLogHandler(logApp *app.LogApp, analysisApp *app.AnalysisApp) *LogHandler {
	return &LogHandler{logApp: logApp, analysisApp: analysisApp}
}

func (h *LogHandler) Query(c *gin.Context) {
	var req struct {
		Source    string `json:"source"`
		Query     string `json:"query"`
		From      string `json:"from"`
		To        string `json:"to"`
		Limit     int    `json:"limit"`
		Direction string `json:"direction"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	q := logdomain.LogQuery{
		Source:    req.Source,
		Query:     req.Query,
		Limit:     req.Limit,
		Direction: req.Direction,
	}

	if req.From != "" {
		if t, err := time.Parse(time.RFC3339, req.From); err == nil {
			q.From = t
		}
	}
	if req.To != "" {
		if t, err := time.Parse(time.RFC3339, req.To); err == nil {
			q.To = t
		}
	}

	result, err := h.logApp.Query(c.Request.Context(), q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *LogHandler) NaturalQuery(c *gin.Context) {
	var req logdomain.NaturalQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	result, err := h.logApp.NaturalQuery(c.Request.Context(), req.Question)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *LogHandler) Analyze(c *gin.Context) {
	var req logdomain.LogAnalyzeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	result, err := h.analysisApp.AnalyzeLogs(c.Request.Context(), req.Logs, req.Context, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gpilot/internal/app"
)

type AnalysisHandler struct {
	analysisApp *app.AnalysisApp
}

func NewAnalysisHandler(analysisApp *app.AnalysisApp) *AnalysisHandler {
	return &AnalysisHandler{analysisApp: analysisApp}
}

func (h *AnalysisHandler) AnalyzeAlert(c *gin.Context) {
	alertID := c.Param("id")

	// Run analysis (streaming is handled via WebSocket)
	result, err := h.analysisApp.AnalyzeAlert(c.Request.Context(), alertID, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *AnalysisHandler) Get(c *gin.Context) {
	result, err := h.analysisApp.GetAnalysis(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "analysis not found"})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *AnalysisHandler) ListByAlert(c *gin.Context) {
	results, err := h.analysisApp.ListByAlert(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"analyses": results})
}

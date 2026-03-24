package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gpilot/internal/app"
)

type DashboardHandler struct {
	dashApp *app.DashboardApp
}

func NewDashboardHandler(dashApp *app.DashboardApp) *DashboardHandler {
	return &DashboardHandler{dashApp: dashApp}
}

func (h *DashboardHandler) Overview(c *gin.Context) {
	overview, err := h.dashApp.GetOverview(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, overview)
}

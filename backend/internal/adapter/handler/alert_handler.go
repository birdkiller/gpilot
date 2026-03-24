package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gpilot/internal/app"
	"gpilot/internal/domain/alert"
)

type AlertHandler struct {
	alertApp *app.AlertApp
}

func NewAlertHandler(alertApp *app.AlertApp) *AlertHandler {
	return &AlertHandler{alertApp: alertApp}
}

func (h *AlertHandler) List(c *gin.Context) {
	q := alert.AlertListQuery{
		Namespace: c.Query("namespace"),
		Search:    c.Query("search"),
	}

	if s := c.Query("status"); s != "" {
		status := alert.Status(s)
		q.Status = &status
	}
	if s := c.Query("severity"); s != "" {
		sev := alert.Severity(s)
		q.Severity = &sev
	}
	if s := c.Query("from"); s != "" {
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			q.From = &t
		}
	}
	if s := c.Query("to"); s != "" {
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			q.To = &t
		}
	}

	q.Page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	q.Size, _ = strconv.Atoi(c.DefaultQuery("size", "20"))

	result, err := h.alertApp.ListAlerts(c.Request.Context(), q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *AlertHandler) Get(c *gin.Context) {
	al, err := h.alertApp.GetAlert(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "alert not found"})
		return
	}
	c.JSON(http.StatusOK, al)
}

func (h *AlertHandler) Acknowledge(c *gin.Context) {
	var req struct {
		User string `json:"user"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		req.User = "anonymous"
	}

	if err := h.alertApp.AcknowledgeAlert(c.Request.Context(), c.Param("id"), req.User); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "acknowledged"})
}

func (h *AlertHandler) ListGroups(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	groups, total, err := h.alertApp.ListGroups(c.Request.Context(), page, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"groups": groups,
		"total":  total,
		"page":   page,
		"size":   size,
	})
}

func (h *AlertHandler) GetGroup(c *gin.Context) {
	group, err := h.alertApp.GetGroup(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "group not found"})
		return
	}
	c.JSON(http.StatusOK, group)
}

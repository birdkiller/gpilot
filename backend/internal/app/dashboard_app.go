package app

import (
	"context"
	"time"

	"gpilot/internal/domain/alert"
	"gpilot/internal/domain/analysis"
)

type DashboardApp struct {
	alertRepo    alert.Repository
	groupRepo    alert.GroupRepository
	analysisRepo analysis.Repository
}

func NewDashboardApp(alertRepo alert.Repository, groupRepo alert.GroupRepository, analysisRepo analysis.Repository) *DashboardApp {
	return &DashboardApp{
		alertRepo:    alertRepo,
		groupRepo:    groupRepo,
		analysisRepo: analysisRepo,
	}
}

type DashboardOverview struct {
	AlertStats     AlertStats        `json:"alert_stats"`
	RecentAnalyses []analysis.Analysis `json:"recent_analyses"`
	TopNamespaces  []NamespaceCount  `json:"top_namespaces"`
}

type AlertStats struct {
	Firing       int64 `json:"firing"`
	Resolved     int64 `json:"resolved"`
	Acknowledged int64 `json:"acknowledged"`
	Suppressed   int64 `json:"suppressed"`
	TotalGroups  int64 `json:"total_groups"`
}

type NamespaceCount struct {
	Namespace string `json:"namespace"`
	Count     int64  `json:"count"`
}

func (a *DashboardApp) GetOverview(ctx context.Context) (*DashboardOverview, error) {
	// Count alerts by status
	firing := countByStatus(ctx, a.alertRepo, alert.StatusFiring)
	resolved := countByStatus(ctx, a.alertRepo, alert.StatusResolved)
	acknowledged := countByStatus(ctx, a.alertRepo, alert.StatusAcknowledged)
	suppressed := countByStatus(ctx, a.alertRepo, alert.StatusSuppressed)

	_, totalGroups, _ := a.groupRepo.List(ctx, 1, 1)

	recentAnalyses, _ := a.analysisRepo.ListRecent(ctx, 5)

	return &DashboardOverview{
		AlertStats: AlertStats{
			Firing:       firing,
			Resolved:     resolved,
			Acknowledged: acknowledged,
			Suppressed:   suppressed,
			TotalGroups:  totalGroups,
		},
		RecentAnalyses: recentAnalyses,
	}, nil
}

func countByStatus(ctx context.Context, repo alert.Repository, status alert.Status) int64 {
	now := time.Now()
	dayAgo := now.Add(-24 * time.Hour)
	result, err := repo.List(ctx, alert.AlertListQuery{
		Status: &status,
		From:   &dayAgo,
		Page:   1,
		Size:   1,
	})
	if err != nil {
		return 0
	}
	return result.Total
}

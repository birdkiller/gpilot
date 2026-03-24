package alert

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusFiring       Status = "firing"
	StatusResolved     Status = "resolved"
	StatusSuppressed   Status = "suppressed"
	StatusAcknowledged Status = "acknowledged"
)

type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityWarning  Severity = "warning"
	SeverityInfo     Severity = "info"
)

type Alert struct {
	ID               uuid.UUID              `json:"id" db:"id"`
	Fingerprint      string                 `json:"fingerprint" db:"fingerprint"`
	GroupID          *uuid.UUID             `json:"group_id,omitempty" db:"group_id"`
	Status           Status                 `json:"status" db:"status"`
	Severity         Severity               `json:"severity" db:"severity"`
	AdjustedSeverity *Severity              `json:"adjusted_severity,omitempty" db:"adjusted_severity"`
	Name             string                 `json:"name" db:"name"`
	Namespace        string                 `json:"namespace" db:"namespace"`
	Pod              string                 `json:"pod,omitempty" db:"pod"`
	Node             string                 `json:"node,omitempty" db:"node"`
	Labels           map[string]string      `json:"labels" db:"labels"`
	Annotations      map[string]string      `json:"annotations" db:"annotations"`
	StartedAt        time.Time              `json:"started_at" db:"started_at"`
	ResolvedAt       *time.Time             `json:"resolved_at,omitempty" db:"resolved_at"`
	LastActiveAt     time.Time              `json:"last_active_at" db:"last_active_at"`
	AcknowledgedBy   string                 `json:"acknowledged_by,omitempty" db:"acknowledged_by"`
	AcknowledgedAt   *time.Time             `json:"acknowledged_at,omitempty" db:"acknowledged_at"`
	Context          map[string]interface{} `json:"context,omitempty" db:"-"`
	CreatedAt        time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at" db:"updated_at"`
}

type AlertGroup struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	Name        string     `json:"name" db:"name"`
	Namespace   string     `json:"namespace" db:"namespace"`
	Status      Status     `json:"status" db:"status"`
	AlertCount  int        `json:"alert_count" db:"alert_count"`
	RootCauseID *uuid.UUID `json:"root_cause_id,omitempty" db:"root_cause_id"`
	Alerts      []Alert    `json:"alerts,omitempty" db:"-"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

// AlertmanagerPayload represents the webhook payload from Alertmanager
type AlertmanagerPayload struct {
	Version     string              `json:"version"`
	GroupKey    string              `json:"groupKey"`
	Status      string              `json:"status"`
	Receiver    string              `json:"receiver"`
	Alerts      []AlertmanagerAlert `json:"alerts"`
	GroupLabels map[string]string   `json:"groupLabels"`
}

type AlertmanagerAlert struct {
	Status       string            `json:"status"`
	Labels       map[string]string `json:"labels"`
	Annotations  map[string]string `json:"annotations"`
	StartsAt     time.Time         `json:"startsAt"`
	EndsAt       time.Time         `json:"endsAt"`
	GeneratorURL string            `json:"generatorURL"`
	Fingerprint  string            `json:"fingerprint"`
}

// ComputeFingerprint generates a unique fingerprint from labels
func ComputeFingerprint(labels map[string]string) string {
	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	for _, k := range keys {
		sb.WriteString(k)
		sb.WriteString("=")
		sb.WriteString(labels[k])
		sb.WriteString(",")
	}

	hash := sha256.Sum256([]byte(sb.String()))
	return fmt.Sprintf("%x", hash[:16])
}

// FromAlertmanager converts an Alertmanager alert to domain Alert
func FromAlertmanager(a AlertmanagerAlert) Alert {
	now := time.Now()
	status := StatusFiring
	if a.Status == "resolved" {
		status = StatusResolved
	}

	severity := SeverityWarning
	if s, ok := a.Labels["severity"]; ok {
		switch Severity(s) {
		case SeverityCritical:
			severity = SeverityCritical
		case SeverityInfo:
			severity = SeverityInfo
		default:
			severity = SeverityWarning
		}
	}

	fingerprint := a.Fingerprint
	if fingerprint == "" {
		fingerprint = ComputeFingerprint(a.Labels)
	}

	alert := Alert{
		ID:           uuid.New(),
		Fingerprint:  fingerprint,
		Status:       status,
		Severity:     severity,
		Name:         a.Labels["alertname"],
		Namespace:    a.Labels["namespace"],
		Pod:          a.Labels["pod"],
		Node:         a.Labels["node"],
		Labels:       a.Labels,
		Annotations:  a.Annotations,
		StartedAt:    a.StartsAt,
		LastActiveAt: now,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if status == StatusResolved && !a.EndsAt.IsZero() {
		alert.ResolvedAt = &a.EndsAt
	}

	return alert
}

// AlertListQuery represents query parameters for listing alerts
type AlertListQuery struct {
	Status    *Status    `json:"status,omitempty"`
	Severity  *Severity  `json:"severity,omitempty"`
	Namespace string     `json:"namespace,omitempty"`
	Search    string     `json:"search,omitempty"`
	From      *time.Time `json:"from,omitempty"`
	To        *time.Time `json:"to,omitempty"`
	Page      int        `json:"page"`
	Size      int        `json:"size"`
}

type AlertListResult struct {
	Alerts []Alert `json:"alerts"`
	Total  int64   `json:"total"`
	Page   int     `json:"page"`
	Size   int     `json:"size"`
}

// MarshalLabels converts labels map to JSON bytes for DB storage
func MarshalLabels(labels map[string]string) ([]byte, error) {
	return json.Marshal(labels)
}

func UnmarshalLabels(data []byte) (map[string]string, error) {
	var labels map[string]string
	err := json.Unmarshal(data, &labels)
	return labels, err
}

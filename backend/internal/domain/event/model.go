package event

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type K8sEvent struct {
	ID             uuid.UUID `json:"id" db:"id"`
	UID            string    `json:"uid" db:"uid"`
	Type           string    `json:"type" db:"type"` // Normal / Warning
	Reason         string    `json:"reason" db:"reason"`
	Message        string    `json:"message" db:"message"`
	Namespace      string    `json:"namespace" db:"namespace"`
	InvolvedObject ObjectRef `json:"involved_object" db:"involved_object"`
	FirstSeen      time.Time `json:"first_seen" db:"first_seen"`
	LastSeen       time.Time `json:"last_seen" db:"last_seen"`
	Count          int       `json:"count" db:"count"`
	Diagnosis      string    `json:"diagnosis,omitempty" db:"-"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

type ObjectRef struct {
	Kind string `json:"kind"`
	Name string `json:"name"`
	UID  string `json:"uid,omitempty"`
}

type EventQuery struct {
	Namespace string     `json:"namespace,omitempty"`
	Type      string     `json:"type,omitempty"`
	From      *time.Time `json:"from,omitempty"`
	To        *time.Time `json:"to,omitempty"`
	Page      int        `json:"page"`
	Size      int        `json:"size"`
}

// Repository defines the persistence interface for K8s events
type Repository interface {
	Upsert(ctx context.Context, event *K8sEvent) error
	GetByID(ctx context.Context, id uuid.UUID) (*K8sEvent, error)
	List(ctx context.Context, query EventQuery) ([]K8sEvent, int64, error)
	ListByNamespace(ctx context.Context, namespace string, limit int) ([]K8sEvent, error)
}

// EventService defines the interface for K8s event operations
type EventService interface {
	List(ctx context.Context, query EventQuery) ([]K8sEvent, int64, error)
	DiagnoseEvent(ctx context.Context, eventID string, streamFn func(chunk string)) (string, error)
}

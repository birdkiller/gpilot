package alert

import (
	"context"

	"github.com/google/uuid"
)

// Repository defines the persistence interface for alerts
type Repository interface {
	Create(ctx context.Context, alert *Alert) error
	Update(ctx context.Context, alert *Alert) error
	GetByID(ctx context.Context, id uuid.UUID) (*Alert, error)
	GetByFingerprint(ctx context.Context, fingerprint string) (*Alert, error)
	List(ctx context.Context, query AlertListQuery) (*AlertListResult, error)
	Acknowledge(ctx context.Context, id uuid.UUID, user string) error
}

// GroupRepository defines the persistence interface for alert groups
type GroupRepository interface {
	Create(ctx context.Context, group *AlertGroup) error
	Update(ctx context.Context, group *AlertGroup) error
	GetByID(ctx context.Context, id uuid.UUID) (*AlertGroup, error)
	FindActiveByNamespace(ctx context.Context, namespace string) (*AlertGroup, error)
	List(ctx context.Context, page, size int) ([]AlertGroup, int64, error)
	IncrementAlertCount(ctx context.Context, id uuid.UUID) error
}

// Cache defines the caching interface for alert deduplication
type Cache interface {
	GetFingerprint(ctx context.Context, fingerprint string) (string, error)
	SetFingerprint(ctx context.Context, fingerprint string, alertID string) error
	IncrFlapping(ctx context.Context, fingerprint string) (int64, error)
}

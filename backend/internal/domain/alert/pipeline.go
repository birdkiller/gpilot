package alert

import (
	"context"

	"gpilot/internal/infra/logger"
)

// Processor is a single step in the alert processing pipeline
type Processor interface {
	Name() string
	Process(ctx context.Context, alerts []*Alert) ([]*Alert, error)
}

// Pipeline chains multiple processors together
type Pipeline struct {
	processors []Processor
}

func NewPipeline(processors ...Processor) *Pipeline {
	return &Pipeline{processors: processors}
}

func (p *Pipeline) Run(ctx context.Context, alerts []*Alert) ([]*Alert, error) {
	var err error
	current := alerts

	for _, proc := range p.processors {
		logger.L.Infow("pipeline step", "processor", proc.Name(), "alert_count", len(current))
		current, err = proc.Process(ctx, current)
		if err != nil {
			logger.L.Errorw("pipeline step failed", "processor", proc.Name(), "error", err)
			return nil, err
		}
		if len(current) == 0 {
			logger.L.Infow("pipeline: all alerts filtered out", "processor", proc.Name())
			return current, nil
		}
	}

	return current, nil
}

// DeduplicateProcessor removes duplicate alerts based on fingerprint
type DeduplicateProcessor struct {
	cache Cache
	repo  Repository
}

func NewDeduplicateProcessor(cache Cache, repo Repository) *DeduplicateProcessor {
	return &DeduplicateProcessor{cache: cache, repo: repo}
}

func (p *DeduplicateProcessor) Name() string { return "deduplicate" }

func (p *DeduplicateProcessor) Process(ctx context.Context, alerts []*Alert) ([]*Alert, error) {
	var result []*Alert

	for _, a := range alerts {
		existingID, err := p.cache.GetFingerprint(ctx, a.Fingerprint)
		if err != nil {
			// Cache miss or error, check DB
			existing, dbErr := p.repo.GetByFingerprint(ctx, a.Fingerprint)
			if dbErr == nil && existing != nil {
				// Update existing alert
				existing.LastActiveAt = a.LastActiveAt
				existing.Status = a.Status
				if a.ResolvedAt != nil {
					existing.ResolvedAt = a.ResolvedAt
				}
				if err := p.repo.Update(ctx, existing); err != nil {
					logger.L.Errorw("failed to update existing alert", "error", err)
				}
				_ = p.cache.SetFingerprint(ctx, a.Fingerprint, existing.ID.String())
				// Still pass it through for notification purposes
				result = append(result, existing)
				continue
			}
		} else if existingID != "" {
			// Found in cache — this is a duplicate, update last active time
			existing, dbErr := p.repo.GetByFingerprint(ctx, a.Fingerprint)
			if dbErr == nil && existing != nil {
				existing.LastActiveAt = a.LastActiveAt
				existing.Status = a.Status
				if a.ResolvedAt != nil {
					existing.ResolvedAt = a.ResolvedAt
				}
				_ = p.repo.Update(ctx, existing)
				result = append(result, existing)
				continue
			}
		}

		// New alert
		result = append(result, a)
	}

	return result, nil
}

// CorrelateProcessor groups related alerts
type CorrelateProcessor struct {
	groupRepo GroupRepository
}

func NewCorrelateProcessor(groupRepo GroupRepository) *CorrelateProcessor {
	return &CorrelateProcessor{groupRepo: groupRepo}
}

func (p *CorrelateProcessor) Name() string { return "correlate" }

func (p *CorrelateProcessor) Process(ctx context.Context, alerts []*Alert) ([]*Alert, error) {
	for _, a := range alerts {
		if a.GroupID != nil {
			continue // Already in a group
		}

		ns := a.Namespace
		if ns == "" {
			ns = "default"
		}

		// Try to find an active group in the same namespace
		group, err := p.groupRepo.FindActiveByNamespace(ctx, ns)
		if err != nil || group == nil {
			// Create new group
			group = &AlertGroup{
				Name:       a.Name,
				Namespace:  ns,
				Status:     StatusFiring,
				AlertCount: 0,
			}
			if err := p.groupRepo.Create(ctx, group); err != nil {
				logger.L.Errorw("failed to create alert group", "error", err)
				continue
			}
		}

		a.GroupID = &group.ID
		if err := p.groupRepo.IncrementAlertCount(ctx, group.ID); err != nil {
			logger.L.Errorw("failed to increment group count", "error", err)
		}
	}

	return alerts, nil
}

// PersistProcessor saves alerts to the database
type PersistProcessor struct {
	repo  Repository
	cache Cache
}

func NewPersistProcessor(repo Repository, cache Cache) *PersistProcessor {
	return &PersistProcessor{repo: repo, cache: cache}
}

func (p *PersistProcessor) Name() string { return "persist" }

func (p *PersistProcessor) Process(ctx context.Context, alerts []*Alert) ([]*Alert, error) {
	for _, a := range alerts {
		// Check if this is an existing alert (already has been persisted)
		existing, _ := p.repo.GetByFingerprint(ctx, a.Fingerprint)
		if existing != nil {
			// Already updated in dedup step
			continue
		}

		if err := p.repo.Create(ctx, a); err != nil {
			logger.L.Errorw("failed to persist alert", "error", err, "fingerprint", a.Fingerprint)
			continue
		}

		_ = p.cache.SetFingerprint(ctx, a.Fingerprint, a.ID.String())
	}

	return alerts, nil
}

// NotifyProcessor sends alerts via WebSocket
type NotifyProcessor struct {
	notifyFn func(alerts []*Alert)
}

func NewNotifyProcessor(notifyFn func(alerts []*Alert)) *NotifyProcessor {
	return &NotifyProcessor{notifyFn: notifyFn}
}

func (p *NotifyProcessor) Name() string { return "notify" }

func (p *NotifyProcessor) Process(ctx context.Context, alerts []*Alert) ([]*Alert, error) {
	if p.notifyFn != nil {
		p.notifyFn(alerts)
	}
	return alerts, nil
}

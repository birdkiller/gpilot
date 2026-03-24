package kubernetes

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gpilot/internal/domain/event"
	"gpilot/internal/infra/logger"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	k8sclient "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Client struct {
	clientset *k8sclient.Clientset
	eventRepo event.Repository
}

func NewClient(kubeconfig string, eventRepo event.Repository) (*Client, error) {
	var config *rest.Config
	var err error

	if kubeconfig != "" {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	} else {
		config, err = rest.InClusterConfig()
	}

	if err != nil {
		// Return a client that operates without K8s connectivity
		logger.L.Warnw("k8s client not available, running in standalone mode", "error", err)
		return &Client{eventRepo: eventRepo}, nil
	}

	clientset, err := k8sclient.NewForConfig(config)
	if err != nil {
		logger.L.Warnw("k8s clientset creation failed", "error", err)
		return &Client{eventRepo: eventRepo}, nil
	}

	return &Client{clientset: clientset, eventRepo: eventRepo}, nil
}

func (c *Client) IsConnected() bool {
	return c.clientset != nil
}

// WatchEvents starts watching K8s events and storing them
func (c *Client) WatchEvents(ctx context.Context) {
	if c.clientset == nil {
		logger.L.Warn("k8s client not connected, event watching disabled")
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		watcher, err := c.clientset.CoreV1().Events("").Watch(ctx, metav1.ListOptions{})
		if err != nil {
			logger.L.Errorw("failed to watch k8s events", "error", err)
			time.Sleep(10 * time.Second)
			continue
		}

		for ev := range watcher.ResultChan() {
			if ev.Type == watch.Error {
				continue
			}
			k8sEvent, ok := ev.Object.(*corev1.Event)
			if !ok {
				continue
			}

			domainEvent := &event.K8sEvent{
				ID:        uuid.New(),
				UID:       string(k8sEvent.UID),
				Type:      k8sEvent.Type,
				Reason:    k8sEvent.Reason,
				Message:   k8sEvent.Message,
				Namespace: k8sEvent.Namespace,
				InvolvedObject: event.ObjectRef{
					Kind: k8sEvent.InvolvedObject.Kind,
					Name: k8sEvent.InvolvedObject.Name,
					UID:  string(k8sEvent.InvolvedObject.UID),
				},
				Count:     int(k8sEvent.Count),
				CreatedAt: time.Now(),
			}

			if !k8sEvent.FirstTimestamp.IsZero() {
				domainEvent.FirstSeen = k8sEvent.FirstTimestamp.Time
			} else {
				domainEvent.FirstSeen = time.Now()
			}
			if !k8sEvent.LastTimestamp.IsZero() {
				domainEvent.LastSeen = k8sEvent.LastTimestamp.Time
			} else {
				domainEvent.LastSeen = time.Now()
			}

			if err := c.eventRepo.Upsert(ctx, domainEvent); err != nil {
				logger.L.Errorw("failed to store k8s event", "error", err)
			}
		}

		logger.L.Warn("k8s event watcher closed, reconnecting...")
		time.Sleep(5 * time.Second)
	}
}

// GetPodEvents returns recent events for a specific pod
func (c *Client) GetPodEvents(ctx context.Context, namespace, podName string) ([]event.K8sEvent, error) {
	if c.eventRepo == nil {
		return nil, nil
	}
	return c.eventRepo.ListByNamespace(ctx, namespace, 50)
}

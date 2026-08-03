package helpers

import (
	"context"
	"fmt"

	pkgworkflows "github.com/nuonco/nuon/pkg/workflows"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

func (h *Helpers) processQueueNamespace(ctx context.Context, orgID string) (string, error) {
	isolated, err := h.featuresClient.OrgCronNamespaceIsolationEnabled(ctx, orgID)
	if err != nil {
		return "", fmt.Errorf("unable to evaluate cron namespace isolation: %w", err)
	}
	if isolated {
		return pkgworkflows.RunnerHealthcheckCronsNamespace, nil
	}
	return "runners", nil
}

func (h *Helpers) EnsureProcessQueueNamespaces(ctx context.Context, runner *app.Runner) error {
	namespace, err := h.processQueueNamespace(ctx, runner.OrgID)
	if err != nil {
		return err
	}

	var queues []app.Queue
	if res := h.db.WithContext(ctx).
		Where(app.Queue{OwnerID: runner.ID, OwnerType: "runners"}).
		Where("name LIKE ?", "runner-process-%").
		Find(&queues); res.Error != nil {
		return fmt.Errorf("unable to list process queues: %w", res.Error)
	}

	for _, q := range queues {
		if _, err := h.queueClient.Create(ctx, &queueclient.CreateQueueRequest{
			OwnerID:     q.OwnerID,
			OwnerType:   q.OwnerType,
			Namespace:   namespace,
			Name:        q.Name,
			MaxInFlight: q.MaxInFlight,
			MaxDepth:    q.MaxDepth,
		}); err != nil {
			return fmt.Errorf("unable to migrate process queue %s: %w", q.Name, err)
		}

		if err := h.emitterClient.MigrateQueueEmitters(ctx, q.ID, namespace); err != nil {
			return fmt.Errorf("unable to migrate emitters for process queue %s: %w", q.Name, err)
		}
	}

	return nil
}

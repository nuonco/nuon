package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	orgshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/helpers"
	queuemigration "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/signals/queue_migration"
)

type EnqueueOrgQueueMigrationsRequest struct {
	Tag string `json:"tag" temporaljson:"tag"`
}

type EnqueueOrgQueueMigrationsResponse struct {
	OrgsEnqueued int `json:"orgs_enqueued" temporaljson:"orgs_enqueued"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 5m
func (a *Activities) EnqueueOrgQueueMigrations(ctx context.Context, req EnqueueOrgQueueMigrationsRequest) (*EnqueueOrgQueueMigrationsResponse, error) {
	var orgs []app.Org
	if res := a.db.WithContext(ctx).
		Where(app.Org{OrgType: app.OrgTypeDefault}).
		Select("id").
		Find(&orgs); res.Error != nil {
		return nil, fmt.Errorf("unable to get default orgs: %w", res.Error)
	}

	// Keyed on the release tag so a retried promotion reuses the same signal
	// instead of stacking one migration per attempt.
	idempotencyKey := ""
	if req.Tag != "" {
		idempotencyKey = fmt.Sprintf("promotion-%s-org-queue-migration", req.Tag)
	}

	for _, org := range orgs {
		if err := a.orgsHelpers.EnqueueOrgSignal(ctx, orgshelpers.EnqueueOrgSignalParams{
			OrgID:          org.ID,
			Signal:         &queuemigration.Signal{OrgID: org.ID},
			IdempotencyKey: idempotencyKey,
		}); err != nil {
			return nil, fmt.Errorf("unable to enqueue queue migration for org %s: %w", org.ID, err)
		}
	}

	return &EnqueueOrgQueueMigrationsResponse{OrgsEnqueued: len(orgs)}, nil
}

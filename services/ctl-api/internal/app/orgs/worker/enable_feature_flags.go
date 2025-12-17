package worker

import (
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/worker/activities"
)

// @temporal-gen workflow
// @execution-timeout 30m
// @task-timeout 15m
func (w *Workflows) EnableFeatureFlags(ctx workflow.Context, sreq signals.RequestSignal) error {
	if err := activities.AwaitEnableFeaturesByOrgID(ctx, sreq.ID); err != nil {
		return errors.Wrap(err, "unable to enable features")
	}

	return nil
}

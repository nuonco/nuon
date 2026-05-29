package worker

import (
	"fmt"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	vcshealthcheck "github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/signals/v2/healthcheck"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
)

// Workflows holds the VCS worker's custom (non-shared) workflows.
type Workflows struct {
	cfg *internal.Config
}

func NewWorkflows(cfg *internal.Config) *Workflows {
	return &Workflows{cfg: cfg}
}

func (w *Workflows) All() []any {
	return []any{
		w.VCSHealthSweep,
	}
}

type VCSHealthSweepRequest struct{}

// VCSHealthSweep is the native-scheduling replacement for the per-connection VCS
// healthcheck cron emitters. A single Temporal Schedule fires this workflow; it
// lists VCS connections and enqueues the existing healthcheck signal onto each
// connection's queue. The signal handler (GitHub API probe) is unchanged - only
// the trigger moves from per-connection crons to one central sweep.
//
// @temporal-gen-v2 workflow
func (w *Workflows) VCSHealthSweep(ctx workflow.Context, _ VCSHealthSweepRequest) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}

	// Work-layer gate: legacy per-connection emitters are authoritative when
	// native scheduling is disabled.
	if !w.cfg.NativeSchedulingEnabled {
		l.Info("native scheduling disabled, vcs health sweep no-op")
		return nil
	}

	ids, err := activities.AwaitListActiveVCSConnectionIDs(ctx, &activities.ListActiveVCSConnectionIDsRequest{})
	if err != nil {
		return errors.Wrap(err, "unable to list vcs connections")
	}

	l.Info("running vcs health sweep", zap.Int("connection-count", len(ids)))

	for _, id := range ids {
		if _, err := sharedactivities.AwaitEnqueueSignalToOwner(ctx, &sharedactivities.EnqueueSignalToOwnerRequest{
			OwnerID:   id,
			OwnerType: "vcs_connections",
			QueueName: fmt.Sprintf("vcs-connection-%s", id),
			Signal: &vcshealthcheck.Signal{
				VCSConnectionID: id,
			},
		}); err != nil {
			l.Warn("unable to enqueue vcs health check",
				zap.String("vcs_connection_id", id),
				zap.Error(err))
		}
	}

	return nil
}

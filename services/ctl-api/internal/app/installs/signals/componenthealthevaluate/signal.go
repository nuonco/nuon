package componenthealthevaluate

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/componenthealthnotify"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
)

const SignalType signal.SignalType = "component-health-evaluate"

// installSignalsQueueName mirrors the constant in installs/helpers,
// duplicated as a literal to avoid an import cycle (helpers imports signals via fx wiring).
const installSignalsQueueName = "install-signals"

// Signal evaluates an install's component health verdicts from the runner's
// observations; the heavy lifting runs in one activity to keep handler history small.
type Signal struct {
	InstallID string `json:"install_id"`
}

var (
	_ signal.Signal                   = (*Signal)(nil)
	_ signal.SignalWithMaxInFlightAge = (*Signal)(nil)
)

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) MaxInFlightAge() time.Duration {
	return 2 * time.Minute
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.InstallID == "" {
		return fmt.Errorf("install_id is required")
	}
	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	resp, err := activities.AwaitEvaluateComponentHealth(ctx, &activities.EvaluateComponentHealthRequest{
		InstallID: s.InstallID,
	})
	if err != nil {
		return fmt.Errorf("unable to evaluate component health: %w", err)
	}

	if resp.Transitions > 0 {
		workflow.GetLogger(ctx).Info("component health verdicts changed",
			"install_id", s.InstallID,
			"transitions", resp.Transitions,
			"notifications", len(resp.Notifications),
		)
	}

	return s.notify(ctx, resp)
}

// notify enqueues one carrier signal per crossing onto the install-signals queue,
// so notifications don't queue behind the next evaluation. A failed enqueue is
// logged, not propagated — the verdict is already committed.
func (s *Signal) notify(ctx workflow.Context, resp *activities.EvaluateComponentHealthResponse) error {
	l := workflow.GetLogger(ctx)

	for _, n := range resp.Notifications {
		body := componenthealthnotify.ComponentSignal{
			InstallID:             s.InstallID,
			InstallName:           resp.InstallName,
			InstallComponentID:    n.InstallComponentID,
			ComponentID:           n.ComponentID,
			ComponentName:         n.ComponentName,
			Health:                n.Health,
			PreviousHealth:        n.PreviousHealth,
			Message:               n.Message,
			RootResourceKind:      n.RootResourceKind,
			RootResourceNamespace: n.RootResourceNamespace,
			RootResourceName:      n.RootResourceName,
		}

		var sig signal.Signal
		if n.Recovered {
			sig = &componenthealthnotify.ComponentRecoveredSignal{ComponentSignal: body}
		} else {
			sig = &componenthealthnotify.ComponentUnhealthySignal{ComponentSignal: body}
		}

		if err := s.enqueue(ctx, sig); err != nil {
			l.Warn("unable to enqueue component health notification",
				"install_id", s.InstallID,
				"install_component_id", n.InstallComponentID,
				"error", err,
			)
		}
	}

	if in := resp.InstallNotification; in != nil {
		sig := &componenthealthnotify.InstallDegradedSignal{
			InstallID:               s.InstallID,
			InstallName:             resp.InstallName,
			Health:                  in.Health,
			PreviousHealth:          in.PreviousHealth,
			Message:                 in.Message,
			UnhealthyComponentCount: in.UnhealthyComponentCount,
			DegradedComponentCount:  in.DegradedComponentCount,
		}
		if err := s.enqueue(ctx, sig); err != nil {
			l.Warn("unable to enqueue install degraded notification",
				"install_id", s.InstallID,
				"error", err,
			)
		}
	}

	return nil
}

func (s *Signal) enqueue(ctx workflow.Context, sig signal.Signal) error {
	_, err := sharedactivities.AwaitEnqueueSignalToOwner(ctx, &sharedactivities.EnqueueSignalToOwnerRequest{
		OwnerID:   s.InstallID,
		OwnerType: "installs",
		QueueName: installSignalsQueueName,
		Signal:    sig,
	})
	return err
}

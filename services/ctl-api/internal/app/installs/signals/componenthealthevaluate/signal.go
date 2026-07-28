package componenthealthevaluate

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "component-health-evaluate"

// Signal evaluates the health verdicts of an install's components from the
// runner's resource observations. Emitted by a per-install cron emitter; the
// heavy lifting (ClickHouse reads, debounce, Postgres writes) happens in a
// single activity so the handler history stays small.
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
		)
	}

	return nil
}

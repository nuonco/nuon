package sandbox

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
)

const (
	pollRunnerTimeout time.Duration = time.Minute * 5
	pollRunnerPeriod  time.Duration = time.Second * 10
)

func (w *Workflows) pollRunner(ctx workflow.Context, runnerID string) error {
	for {
		runner, err := activities.AwaitGetRunnerByID(ctx, runnerID)
		if err != nil {
			return fmt.Errorf("unable to get runner from database: %w", err)
		}

		if runner.Status == app.RunnerStatusActive {
			return nil
		}

		if runner.Status == app.RunnerStatusError {
			return fmt.Errorf("runner is in error state")
		}

		workflow.Sleep(ctx, pollRunnerPeriod)
	}
}

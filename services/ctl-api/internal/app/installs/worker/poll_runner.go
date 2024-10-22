package worker

import (
	"errors"
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
)

const (
	pollRunnerTimeout time.Duration = time.Minute * 5
	pollRunnerPeriod  time.Duration = time.Second * 10
)

func (w *Workflows) pollRunnerNotFound(ctx workflow.Context, runnerID string) error {
	for {
		_, err := activities.AwaitGetRunnerByID(ctx, runnerID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}

			return fmt.Errorf("unable to get runner from database: %w", err)
		}

		workflow.Sleep(ctx, pollRunnerPeriod)
	}
}

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

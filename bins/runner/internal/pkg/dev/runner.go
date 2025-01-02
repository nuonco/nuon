package dev

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/powertoolsdev/mono/pkg/retry"
)

func (d *devver) initRunner(ctx context.Context) error {
	if os.Getenv("RUNNER_ID") != "" {
		fmt.Println("runner id set from environment")
		d.runnerID = os.Getenv("RUNNER_ID")
		return nil
	}

	fn := func(ctx context.Context) error {
		runners, err := d.apiClient.ListRunners(ctx, d.watchRunnerType)
		if err != nil {
			return err
		}

		if len(runners) < 1 {
			return fmt.Errorf("no runners found")
		}

		d.runnerID = runners[0].ID
		return nil
	}

	// we will look for up to an hour for a runner to be created
	if err := retry.Retry(ctx, fn,
		retry.WithMaxAttempts(-1),
		retry.WithTimeout(time.Hour),
		retry.WithSleep(time.Second*5),
		retry.WithCBHook(func(attempt int) error {
			fmt.Println("waiting 5 seconds and trying again", d.watchRunnerType, "context")
			return nil
		}),
	); err != nil {
		return err
	}

	if d.runnerID == "" {
		return fmt.Errorf("logic is bad")
	}

	os.Setenv("RUNNER_ID", d.runnerID)
	fmt.Println("successfully set runner ID ", d.runnerID)

	return nil
}

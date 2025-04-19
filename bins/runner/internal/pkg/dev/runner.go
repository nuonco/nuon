package dev

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/pkg/errors"

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

		// once a runner is created, we must wait until the service account is created (as part of the provisioning
		// process), before running locally.
		_, err = d.apiClient.GetRunnerServiceAccount(ctx, runners[0].ID)
		if err != nil {
			fmt.Println("runner is created, but service account is not ready yet")
			return errors.Wrap(err, "unable to get service account")
		}

		// need to wait for the runner to have a local aws iam role to run with locally.
		// this role must be assumable by the support role

		d.runnerID = runners[0].ID
		return nil
	}

	// we will look for up to an hour for a runner to be created
	if err := retry.Retry(ctx, fn,
		retry.WithMaxAttempts(-1),
		retry.WithTimeout(time.Hour),
		retry.WithSleep(time.Second*5),
		retry.WithCBHook(func(_ context.Context, attempt int) error {
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

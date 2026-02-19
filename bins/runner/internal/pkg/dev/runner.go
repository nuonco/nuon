package dev

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/api"
	"github.com/nuonco/nuon/pkg/retry"
)

func (d *devver) initRunner(ctx context.Context) error {
	// Warn if legacy env vars are set — they are ignored in favor of self-registration.
	if os.Getenv("ORG_RUNNER_ID") != "" {
		fmt.Println("warning: ORG_RUNNER_ID is set but will be ignored; local runner will self-register instead.")
	}
	if os.Getenv("INSTALL_RUNNER_ID") != "" {
		fmt.Println("warning: INSTALL_RUNNER_ID is set but will be ignored; local runner will self-register instead.")
	}

	fn := func(ctx context.Context) error {
		runners, err := d.apiClient.ListRunners(ctx, d.watchRunnerType)
		if err != nil {
			return err
		}

		if len(runners) < 1 {
			return fmt.Errorf("no runners found")
		}

		// Check if a local runner already exists and cloud runners are already tainted
		var localRunner *api.Runner
		allCloudTainted := true
		for i := range runners {
			if runners[i].Platform == "local" {
				localRunner = &runners[i]
			} else if !runners[i].Tainted {
				allCloudTainted = false
			}
		}

		if localRunner != nil && allCloudTainted {
			fmt.Printf("local runner %s already registered and cloud runners tainted, skipping registration\n", localRunner.ID)
			d.runnerID = localRunner.ID
			d.runnerGroupID = localRunner.RunnerGroupID
			return nil
		}

		// Get the runner group ID from the first runner to self-register
		runnerGroupID := runners[0].RunnerGroupID
		if runnerGroupID == "" {
			return fmt.Errorf("runner group ID not set on runner %s", runners[0].ID)
		}

		// Self-register: find-or-create a local runner in this group
		resp, err := d.apiClient.CreateRunnerInGroup(ctx, runnerGroupID, "local")
		if err != nil {
			fmt.Println("unable to self-register local runner, falling back to cloud runner identity")
			return errors.Wrap(err, "unable to create local runner in group")
		}

		d.runnerID = resp.Runner.ID
		d.runnerGroupID = runnerGroupID
		if resp.Token != "" {
			d.runnerAPIToken = resp.Token
		}

		// Taint cloud runners in the group so the local runner wins leader election
		for _, runner := range runners {
			if runner.ID != d.runnerID && runner.Platform != "local" && !runner.Tainted {
				fmt.Printf("tainting cloud runner %s (%s)\n", runner.ID, runner.Name)
				if err := d.apiClient.TaintRunner(ctx, runner.ID); err != nil {
					fmt.Printf("warning: unable to taint runner %s: %v\n", runner.ID, err)
				}
			}
		}

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

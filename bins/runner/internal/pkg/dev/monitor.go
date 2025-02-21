package dev

import (
	"context"
	"fmt"
	"time"

	smithytime "github.com/aws/smithy-go/time"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/retry"
)

func (d *devver) monitorRunners() error {
	fmt.Println("monitoring runners to restart ", d.watchRunnerType)
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for {
		fn := func(ctx context.Context) error {
			runners, err := d.apiClient.ListRunners(ctx, d.watchRunnerType)
			if err != nil {
				return err
			}
			if len(runners) < 1 {
				return fmt.Errorf("no runners found")
			}

			if d.runnerID != runners[0].ID {
				fmt.Println("new runner was found, so restarting to act as " + runners[0].ID)
				return retry.AsNonRetryable(fmt.Errorf("new runner has been created, restarting"))
			}

			// make sure runner has a service account
			_, err = d.apiClient.GetRunnerServiceAccount(ctx, runners[0].ID)
			if err != nil {
				fmt.Println("runner service account does not exist yet, polling until it does")
				return errors.Wrap(err, "no service account created yet")
			}

			return nil
		}

		if err := retry.Retry(ctx, fn,
			retry.WithMaxAttempts(-1),
			retry.WithTimeout(time.Hour),
			retry.WithSleep(time.Second*5),
			retry.WithCBHook(func(_ context.Context, attempt int) error {
				fmt.Println("waiting 5 seconds and trying again")
				return nil
			}),
		); err != nil {
			return err
		}

		smithytime.SleepWithContext(ctx, time.Second*5)
	}
}

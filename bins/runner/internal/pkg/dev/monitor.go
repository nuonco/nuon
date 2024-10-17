package dev

import (
	"context"
	"fmt"
	"time"

	smithytime "github.com/aws/smithy-go/time"

	"github.com/powertoolsdev/mono/pkg/retry"
)

func (d *devver) monitorRunners() error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for {
		fn := func(ctx context.Context) error {
			runners, err := d.apiClient.ListRunners(ctx, d.runnerIDInput)
			if err != nil {
				return err
			}
			if len(runners) < 1 {
				return fmt.Errorf("no runners found")
			}

			if d.runnerID != runners[0].ID {
				fmt.Println("new runner was found, so restarting to act as " + runners[0].ID)
				return retry.AsNonRetryable(fmt.Errorf("new org runner has been created, restarting"))
			}
			return nil
		}

		if err := retry.Retry(ctx, fn,
			retry.WithMaxAttempts(-1),
			retry.WithTimeout(time.Hour),
			retry.WithSleep(time.Second*5),
			retry.WithCBHook(func(attempt int) error {
				fmt.Println("waiting 5 seconds and trying again")
				return nil
			}),
		); err != nil {
			return err
		}

		smithytime.SleepWithContext(ctx, time.Second*5)
	}
}

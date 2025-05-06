package dev

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/powertoolsdev/mono/pkg/retry"
)

func (d *devver) initToken(ctx context.Context) error {
	if d.runnerAPIToken != "" {
		fmt.Println("using runner api token from environment")
		return nil
	}

	fmt.Println("no runner api token found in environment, looking up a token from the api")
	fn := func(ctx context.Context) error {
		token, err := d.apiClient.GetRunnerServiceAccountToken(ctx, d.runnerID, time.Hour, false)
		if err != nil {
			return fmt.Errorf("unable to get service account token: %w", err)
		}

		d.runnerAPIToken = token

		return nil
	}
	if err := retry.Retry(
		ctx,
		fn,
		retry.WithMaxAttempts(5),
		retry.WithSleep(time.Second*5),
	); err != nil {
		return err
	}

	os.Setenv("RUNNER_API_TOKEN", d.runnerAPIToken)
	fmt.Println("successfully set runner api token", d.runnerAPIToken)
	return nil
}

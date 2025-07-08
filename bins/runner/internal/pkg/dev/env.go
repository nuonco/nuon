package dev

import (
	"context"
	"fmt"
	"os"

	"github.com/pkg/errors"
)

func (d *devver) initEnv(ctx context.Context) error {
	runner, err := d.apiClient.GetRunner(ctx, d.runnerID)
	if err != nil {
		return errors.Wrap(err, "unable to get runner")
	}

	runnerGroup, err := d.apiClient.GetRunnerGroup(ctx, runner.RunnerGroupID)
	if err != nil {
		return errors.Wrap(err, "unable to get runner group")
	}

	if runnerGroup.Type == "org" {
		return nil
	}

	fmt.Println("since this is an install runner instance, we set the registry port to 5002 to prevent conflicts if running an org runner as well")
	os.Setenv("REGISTRY_PORT", "5002")
	return nil
}

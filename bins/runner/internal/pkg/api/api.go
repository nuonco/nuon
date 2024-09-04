package api

import (
	"fmt"

	nuonrunner "github.com/nuonco/nuon-runner-go"

	"github.com/powertoolsdev/mono/bins/runner/internal"
)

func New(cfg *internal.Config) (nuonrunner.Client, error) {
	api, err := nuonrunner.New(
		nuonrunner.WithURL(cfg.RunnerAPIURL),
		nuonrunner.WithRunnerID(cfg.RunnerID),
		nuonrunner.WithAuthToken(cfg.RunnerAPIToken),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize runner: %w", err)
	}

	return api, nil
}

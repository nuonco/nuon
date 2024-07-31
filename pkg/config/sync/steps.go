package sync

import (
	"context"
	"fmt"
)

func (s *sync) syncSteps() ([]syncStep, error) {
	steps := []syncStep{
		{
			Resource: "app",
			Method: func(ctx context.Context) error {
				return s.syncApp(ctx, "app")
			},
		},
		{
			Resource: "app-inputs",
			Method: func(ctx context.Context) error {
				return s.syncAppInput(ctx, "app-inputs")
			},
		},
		{
			Resource: "app-sandbox",
			Method: func(ctx context.Context) error {
				return s.syncAppSandbox(ctx, "app-sandbox")
			},
		},
		{
			Resource: "app-runner",
			Method: func(ctx context.Context) error {
				return s.syncAppRunner(ctx, "runner")
			},
		},
		{
			Resource: "app-installer",
			Method: func(ctx context.Context) error {
				return s.syncAppInstaller(ctx, "installer")
			},
		},
	}

	deps := make([]string, 0)
	for _, comp := range s.cfg.Components {
		// thanks russ cox
		obj := comp
		obj.Dependencies = deps

		resourceName := fmt.Sprintf("component-%s", obj.Name)
		steps = append(steps, syncStep{
			Resource: resourceName,
			Method: func(ctx context.Context) error {
				compID, err := s.syncComponent(ctx, resourceName, obj)
				if err != nil {
					return err
				}

				deps = append(deps, compID)
				return nil
			},
		})
	}

	return steps, nil
}

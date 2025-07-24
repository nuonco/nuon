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
			Resource: "app-branch",
			Method: func(ctx context.Context) error {
				// TODO: Implement syncAppBranch method
				return nil
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
			Resource: "installer",
			Method: func(ctx context.Context) error {
				return s.syncAppInstaller(ctx, "installer")
			},
		},
		{
			Resource: "app-permissions",
			Method: func(ctx context.Context) error {
				return s.syncAppPermissions(ctx, "permissions")
			},
		},
		{
			Resource: "app-policies",
			Method: func(ctx context.Context) error {
				return s.syncAppPolicies(ctx, "policies")
			},
		},
		{
			Resource: "app-secrets",
			Method: func(ctx context.Context) error {
				return s.syncAppSecrets(ctx, "secrets")
			},
		},
		{
			Resource: "app-break-glass",
			Method: func(ctx context.Context) error {
				return s.syncAppBreakGlass(ctx, "break-glass")
			},
		},
		{
			Resource: "app-cloudformation-stack",
			Method: func(ctx context.Context) error {
				return s.syncAppCloudFormationStack(ctx, "cloudformation-stack")
			},
		},
	}

	// ensure all components
	for _, comp := range s.cfg.Components {
		resourceName := fmt.Sprintf("component-%s", comp.Name)
		steps = append(steps, syncStep{
			Resource: resourceName,
			Method: func(ctx context.Context) error {
				err := s.ensureComponent(ctx, resourceName, comp)
				if err != nil {
					return err
				}

				return nil
			},
		})
	}

	// sync all components and their configs
	for _, comp := range s.cfg.Components {
		resourceName := fmt.Sprintf("component-%s", comp.Name)
		steps = append(steps, syncStep{
			Resource: resourceName,
			Method: func(ctx context.Context) error {
				_, err := s.syncComponent(ctx, resourceName, comp)
				if err != nil {
					return err
				}

				return nil
			},
		})
	}

	for _, action := range s.cfg.Actions {
		obj := action

		resourceName := fmt.Sprintf("action-%s", obj.Name)
		steps = append(steps, syncStep{
			Resource: resourceName,
			Method: func(ctx context.Context) error {
				_, _, err := s.syncAction(ctx, resourceName, obj)
				if err != nil {
					return err
				}

				return nil
			},
		})
	}

	return steps, nil
}

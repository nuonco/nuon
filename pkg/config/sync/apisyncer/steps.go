package apisyncer

import (
	"context"
	"fmt"
)

// syncPhases returns the ordered list of sync phases. Steps within a phase
// are independent of each other and run concurrently; phases run
// sequentially relative to each other.
//
// Phase order:
//  1. top-level app config (inputs, sandbox, permissions, etc.) — independent
//  2. ensure all components — must finish before any component sync, since
//     sync paths read component IDs created during ensure
//  3. sync all components (including their configs / build scheduling)
//  4. actions
//  5. runbooks (depend on actions existing in the API so step refs by name
//     resolve cleanly)
func (s *syncer) syncPhases() ([][]syncStep, error) {
	topLevel := []syncStep{
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
			Resource: "app-permissions",
			Method: func(ctx context.Context) error {
				return s.syncAppPermissions(ctx, "permissions")
			},
		},
		{
			Resource: "app-operations-roles",
			Method: func(ctx context.Context) error {
				return s.syncAppOperationRules(ctx, "operations-roles")
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

	ensureComponents := make([]syncStep, 0, len(s.cfg.Components))
	syncComponents := make([]syncStep, 0, len(s.cfg.Components))
	for _, comp := range s.cfg.Components {
		comp := comp
		resourceName := fmt.Sprintf("component-%s", comp.Name)
		ensureComponents = append(ensureComponents, syncStep{
			Resource: resourceName,
			Method: func(ctx context.Context) error {
				return s.ensureComponent(ctx, resourceName, comp)
			},
		})
		syncComponents = append(syncComponents, syncStep{
			Resource: resourceName,
			Method: func(ctx context.Context) error {
				_, err := s.syncComponent(ctx, resourceName, comp)
				return err
			},
		})
	}

	actions := make([]syncStep, 0, len(s.cfg.Actions))
	for _, action := range s.cfg.Actions {
		obj := action
		resourceName := fmt.Sprintf("action-%s", obj.Name)
		actions = append(actions, syncStep{
			Resource: resourceName,
			Method: func(ctx context.Context) error {
				_, _, err := s.syncAction(ctx, resourceName, obj)
				return err
			},
		})
	}

	runbooks := make([]syncStep, 0, len(s.cfg.Runbooks))
	for _, runbook := range s.cfg.Runbooks {
		obj := runbook
		resourceName := fmt.Sprintf("runbook-%s", obj.Name)
		runbooks = append(runbooks, syncStep{
			Resource: resourceName,
			Method: func(ctx context.Context) error {
				_, _, err := s.syncRunbook(ctx, resourceName, obj)
				return err
			},
		})
	}

	return [][]syncStep{
		topLevel,
		ensureComponents,
		syncComponents,
		actions,
		runbooks,
	}, nil
}

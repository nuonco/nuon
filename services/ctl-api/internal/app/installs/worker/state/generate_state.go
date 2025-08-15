package state

import (
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/types/state"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
)

type GenerateStateRequest struct {
	InstallID string

	TriggeredByID   string
	TriggeredByType string

	NoWrite bool
}

type statePartial struct {
	name string
	fn   func(workflow.Context, string, *state.State) error
}

// GenerateState generates the state for an install
// @temporal-gen workflow
// @execution-timeout 60m
// @task-timeout 30m
// @id-template {{.CallerID}}-generate-state
func (w *Workflows) GenerateState(ctx workflow.Context, req *GenerateStateRequest) (*state.State, error) {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get logger")
	}
	l.Info("generating workflow state")

	is := state.New()

	install, err := activities.AwaitGetByInstallID(ctx, req.InstallID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install")
	}
	is.ID = install.ID
	is.Name = install.Name

	partials := []statePartial{
		{
			name: "org",
			fn: func(ctx workflow.Context, installID string, state *state.State) error {
				partial, err := w.getOrgStatePartial(ctx, req.InstallID)
				if err != nil {
					return err
				}
				is.Org = partial
				return nil
			},
		},
		{
			name: "app",
			fn: func(ctx workflow.Context, installID string, state *state.State) error {
				partial, err := w.getAppStatePartial(ctx, req.InstallID)
				if err != nil {
					return err
				}
				is.App = partial
				return nil
			},
		},
		{
			name: "domain",
			fn: func(ctx workflow.Context, installID string, state *state.State) error {
				partial, err := w.getDomainPartial(ctx, req.InstallID)
				if err != nil {
					return err
				}
				is.Domain = partial
				return nil
			},
		},
		{
			name: "runner",
			fn: func(ctx workflow.Context, installID string, state *state.State) error {
				partial, err := w.getRunnerStatePartial(ctx, req.InstallID)
				if err != nil {
					return err
				}
				is.Runner = partial
				return nil
			},
		},
		{
			name: "domain",
			fn: func(ctx workflow.Context, installID string, state *state.State) error {
				partial, err := w.getDomainPartial(ctx, req.InstallID)
				if err != nil {
					return err
				}
				is.Domain = partial
				return nil
			},
		},
		{
			name: "cloud",
			fn: func(ctx workflow.Context, installID string, state *state.State) error {
				partial, err := w.toCloudAccount(ctx, req.InstallID)
				if err != nil {
					return err
				}
				is.Cloud = partial
				return nil
			},
		},
		{
			name: "actions",
			fn: func(ctx workflow.Context, installID string, state *state.State) error {
				partial, err := w.getActionsStatePartial(ctx, req.InstallID)
				if err != nil {
					return err
				}
				is.Actions = partial
				return nil
			},
		},
		{
			name: "inputs",
			fn: func(ctx workflow.Context, installID string, state *state.State) error {
				partial, err := w.getInputsStatePartial(ctx, req.InstallID)
				if err != nil {
					return err
				}
				is.Inputs = partial
				return nil
			},
		},
		{
			name: "components",
			fn: func(ctx workflow.Context, installID string, st *state.State) error {
				partial, err := w.getStateComponentsPartial(ctx, req.InstallID)
				if err != nil {
					return err
				}

				is.Components = make(map[string]any, 0)
				for name, c := range partial.Components {
					cMap, err := state.AsMap(c)
					if err != nil {
						return errors.Wrap(err, "unable to create map")
					}

					is.Components[name] = cMap
				}
				return nil
			},
		},
		{
			name: "sandbox",
			fn: func(ctx workflow.Context, installID string, state *state.State) error {
				partial, err := w.getSandboxStatePartial(ctx, req.InstallID)
				if err != nil {
					return err
				}
				is.Sandbox = partial
				return nil
			},
		},
		{
			name: "stack",
			fn: func(ctx workflow.Context, installID string, state *state.State) error {
				partial, err := w.getStackStatePartial(ctx, req.InstallID)
				if err != nil {
					return err
				}
				is.InstallStack = partial
				return nil
			},
		},
		{
			name: "secrets",
			fn: func(ctx workflow.Context, installID string, state *state.State) error {
				partial, err := w.getSecretsStatePartial(ctx, req.InstallID)
				if err != nil {
					return err
				}
				is.Secrets = partial
				return nil
			},
		},
		{
			name: "legacy",
			fn: func(ctx workflow.Context, installID string, state *state.State) error {
				w.mapLegacyFields(state)
				return nil
			},
		},
	}

	for _, partial := range partials {
		l.Info("fetching partial state " + partial.name)
		if err := partial.fn(ctx, req.InstallID, is); err != nil {
			return nil, errors.Wrap(err, "unable to get partial state "+partial.name)
		}
	}

	// write the state into db
	if req.NoWrite {
		return is, nil
	}

	if _, err := activities.AwaitSaveState(ctx, &activities.SaveStateRequest{
		State:           is,
		InstallID:       req.InstallID,
		TriggeredByID:   req.TriggeredByID,
		TriggeredByType: req.TriggeredByType,
	}); err != nil {
		return nil, errors.Wrap(err, "unable to write state")
	}

	if err := activities.AwaitArchiveState(ctx, &activities.ArchiveStateRequest{
		InstallID: req.InstallID,
	}); err != nil {
		return nil, errors.Wrap(err, "unable to purge stale state")
	}

	return is, nil
}

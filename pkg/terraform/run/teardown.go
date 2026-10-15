package run

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-hclog"
	tfjson "github.com/hashicorp/terraform-json"

	"github.com/nuonco/nuon/pkg/terraform/teardown"
)

// WithTeardownRetain makes destroy plans leave the matching resources in
// place, and makes applying the plan remove them from state.
func WithTeardownRetain(retain []string) runOption {
	return func(r *run) error {
		r.teardownRetain = retain
		return nil
	}
}

type teardownResolution struct {
	state    *tfjson.State
	retained []string
	destroy  []string
}

func (r *run) resolveTeardown(ctx context.Context, log hclog.Logger) error {
	state, err := r.Workspace.Show(ctx, log)
	if err != nil {
		return fmt.Errorf("unable to read state: %w", err)
	}

	retained, destroy, unmatched, err := teardown.Partition(teardown.ManagedAddresses(state), r.teardownRetain)
	if err != nil {
		return fmt.Errorf("unable to resolve retained resources: %w", err)
	}
	for _, addr := range unmatched {
		log.Warn("teardown.retain entry matches no resources in state", "address", addr)
	}
	log.Info("resolved teardown resources", "retained", retained, "destroy", destroy)

	r.teardown = &teardownResolution{
		state:    state,
		retained: retained,
		destroy:  destroy,
	}
	return nil
}

func (r *run) planTeardown(ctx context.Context, log hclog.Logger) ([]byte, error) {
	if len(r.teardown.destroy) == 0 {
		log.Info("every resource is retained, planning a refresh only")
		return r.Workspace.PlanRefreshOnly(ctx, log)
	}

	return r.Workspace.PlanDestroyTargets(ctx, log, r.teardown.destroy)
}

func (r *run) showTeardownPlan(ctx context.Context, log hclog.Logger) (*tfjson.Plan, error) {
	plan, err := r.Workspace.ShowPlan(ctx, log)
	if err != nil {
		return nil, err
	}

	if err := teardown.CheckPlan(plan, r.teardownRetain); err != nil {
		return nil, err
	}

	teardown.AddForgetChanges(plan, r.teardown.state, r.teardown.retained)
	return plan, nil
}

// removeRetainedFromState runs after the destroy plan is applied, so a failed
// destroy leaves retained resources tracked and the teardown can be retried.
func (r *run) removeRetainedFromState(ctx context.Context, log hclog.Logger) error {
	state, err := r.Workspace.Show(ctx, log)
	if err != nil {
		return fmt.Errorf("unable to read state: %w", err)
	}

	retained, remaining, _, err := teardown.Partition(teardown.ManagedAddresses(state), r.teardownRetain)
	if err != nil {
		return fmt.Errorf("unable to resolve retained resources: %w", err)
	}
	if len(remaining) > 0 {
		log.Warn("resources that are not retained are still in state after teardown", "addresses", remaining)
	}
	if len(retained) == 0 {
		return nil
	}

	log.Info("removing retained resources from state", "addresses", retained)
	return r.Workspace.StateRm(ctx, log, retained)
}

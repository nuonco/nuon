package workspace

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/terraform-exec/tfexec"

	"github.com/nuonco/nuon/pkg/terraform/workspace/output"
)

func (w *workspace) PlanDestroy(ctx context.Context, log hclog.Logger) ([]byte, error) {
	client, err := w.getClient(ctx, log)
	if err != nil {
		return nil, err
	}

	return w.planDestroy(ctx, client, log, tfexec.Destroy(true))
}

// PlanDestroyTargets plans the destruction of only the given resource
// addresses. Terraform also destroys anything that depends on a target.
func (w *workspace) PlanDestroyTargets(ctx context.Context, log hclog.Logger, targets []string) ([]byte, error) {
	if len(targets) == 0 {
		return nil, fmt.Errorf("at least one target is required, an untargeted destroy plan destroys everything")
	}

	client, err := w.getClient(ctx, log)
	if err != nil {
		return nil, err
	}

	opts := []tfexec.PlanOption{tfexec.Destroy(true)}
	for _, target := range targets {
		opts = append(opts, tfexec.Target(target))
	}
	return w.planDestroy(ctx, client, log, opts...)
}

// PlanRefreshOnly writes a saved plan that only syncs state with the real
// infrastructure, for teardowns where nothing is destroyed.
func (w *workspace) PlanRefreshOnly(ctx context.Context, log hclog.Logger) ([]byte, error) {
	client, err := w.getClient(ctx, log)
	if err != nil {
		return nil, err
	}

	return w.planDestroy(ctx, client, log, tfexec.RefreshOnly(true))
}

func (w *workspace) planDestroy(ctx context.Context, client Terraform, log hclog.Logger, modeOpts ...tfexec.PlanOption) ([]byte, error) {
	out, err := output.New(w.v, output.WithLogger(log))
	if err != nil {
		return nil, fmt.Errorf("unable to get output: %w", err)
	}

	writer, err := out.Writer()
	if err != nil {
		return nil, fmt.Errorf("unable to get writer: %w", err)
	}

	tfplanFilePath := filepath.Join(w.Root(), "tfplan")
	opts := []tfexec.PlanOption{
		tfexec.Refresh(true),
		tfexec.Out(tfplanFilePath), // NOTE: this should probably be configured w/ a WithPlanOut
	}
	opts = append(opts, modeOpts...)
	for _, fp := range w.varsPaths {
		opts = append(opts, tfexec.VarFile(fp))
	}

	if _, err := client.PlanJSON(ctx,
		writer,
		opts...,
	); err != nil {
		return nil, fmt.Errorf("unable to plan: %w", err)
	}

	return out.Bytes()
}

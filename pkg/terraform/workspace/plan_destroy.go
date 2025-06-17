package workspace

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/terraform-exec/tfexec"

	"github.com/powertoolsdev/mono/pkg/terraform/workspace/output"
)

func (w *workspace) PlanDestroy(ctx context.Context, log hclog.Logger) ([]byte, error) {
	client, err := w.getClient(ctx, log)
	if err != nil {
		return nil, err
	}

	return w.planDestroy(ctx, client, log)
}

func (w *workspace) planDestroy(ctx context.Context, client Terraform, log hclog.Logger) ([]byte, error) {
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
		tfexec.Destroy(true),
		tfexec.Out(tfplanFilePath), // NOTE: this should probably be configured w/ a WithPlanOut
	}
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

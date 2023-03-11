package executor

import (
	"context"

	"github.com/hashicorp/terraform-exec/tfexec"
)

type planner interface {
	Plan(context.Context, ...tfexec.PlanOption) (bool, error)
}

var _ planner = (*tfexec.Terraform)(nil)

// Plan runs terraform plan
func (e *tfExecutor) Plan(ctx context.Context) error {
	// TODO(jdt): it may be useful to return whether the plan had a diff at some point...
	if _, err := e.planner.Plan(ctx, tfexec.Refresh(true), tfexec.VarFile(e.VarFile)); err != nil {
		return err
	}

	return nil
}

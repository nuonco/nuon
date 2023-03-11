package executor

import (
	"context"

	"github.com/hashicorp/terraform-exec/tfexec"
)

type applier interface {
	Apply(context.Context, ...tfexec.ApplyOption) error
}

var _ applier = (*tfexec.Terraform)(nil)

// Apply runs terraform apply for the current module
func (e *tfExecutor) Apply(ctx context.Context) error {
	if err := e.applier.Apply(ctx, tfexec.Refresh(true), tfexec.VarFile(e.VarFile)); err != nil {
		return err
	}

	return nil
}

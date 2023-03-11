package executor

import (
	"context"

	"github.com/hashicorp/terraform-exec/tfexec"
)

type destroyer interface {
	Destroy(context.Context, ...tfexec.DestroyOption) error
}

var _ destroyer = (*tfexec.Terraform)(nil)

// Destroy runs terraform apply -destroy for the current module
func (e *tfExecutor) Destroy(ctx context.Context) error {
	if err := e.destroyer.Destroy(ctx, tfexec.Refresh(true), tfexec.VarFile(e.VarFile)); err != nil {
		return err
	}

	return nil
}

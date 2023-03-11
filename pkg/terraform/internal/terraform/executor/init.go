package executor

import (
	"context"

	"github.com/hashicorp/terraform-exec/tfexec"
)

type initer interface {
	Init(context.Context, ...tfexec.InitOption) error
}

var _ initer = (*tfexec.Terraform)(nil)

// Init runs terraform init for the current module
func (e *tfExecutor) Init(ctx context.Context) error {
	if err := e.initer.Init(ctx, tfexec.BackendConfig(e.BackendConfigFile)); err != nil {
		return err
	}

	return nil
}

package builder

import (
	"context"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/waypoint-plugin-sdk/component"
	"github.com/hashicorp/waypoint-plugin-sdk/terminal"
	terraformv1 "github.com/powertoolsdev/mono/pkg/types/plugins/terraform/v1"
)

func (b *Builder) BuildFunc() interface{} {
	return b.build
}

func (b *Builder) BuildODRFunc() interface{} {
	return b.buildODR
}

// build creates and uploads an OCI artifact of the terraform module to the provided ECR repository
func (b *Builder) build(ctx context.Context,
	ui terminal.UI,
	src *component.Source,
	log hclog.Logger) (*terraformv1.BuildOutput, error) {
	u := ui.Status()
	defer u.Close()
	u.Step(terminal.StatusOK, "noop")
	return &terraformv1.BuildOutput{}, nil
}

func (b *Builder) buildODR(
	ctx context.Context,
	ui terminal.UI,
	src *component.Source,
	log hclog.Logger,
	ai *terraformv1.AccessInfo,
) (*terraformv1.BuildOutput, error) {
	return b.build(ctx, ui, src, log)
}

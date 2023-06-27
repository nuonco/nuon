package builder

import (
	"context"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/waypoint-plugin-sdk/component"
	"github.com/hashicorp/waypoint-plugin-sdk/terminal"
	helmv1 "github.com/powertoolsdev/mono/pkg/types/plugins/helm/v1"
)

func (b *Builder) BuildFunc() interface{} {
	return b.build
}

// build creates and uploads an OCI artifact of the terraform module to the provided ECR repository
func (b *Builder) build(ctx context.Context,
	ui terminal.UI,
	src *component.Source,
	log hclog.Logger) (*helmv1.BuildOutput, error) {
	ui.Output("executing noop build...")
	return &helmv1.BuildOutput{}, nil
}

package builder

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/waypoint-plugin-sdk/component"
	"github.com/hashicorp/waypoint-plugin-sdk/terminal"
	ociv1 "github.com/powertoolsdev/mono/pkg/types/plugins/oci/v1"
)

func (b *Builder) BuildFunc() interface{} {
	return b.build
}

func (b *Builder) BuildODRFunc() interface{} {
	return b.buildODR
}

func (b *Builder) buildODR(
	ctx context.Context,
	ui terminal.UI,
	src *component.Source,
	log hclog.Logger,
	accessInfo *ociv1.AccessInfo,
) (*ociv1.BuildOutput, error) {
	ui.Output("starting odr build")
	ui.Output("got access info credentials: %v", accessInfo)

	return &ociv1.BuildOutput{}, nil
}

// build creates and uploads an OCI artifact of the terraform module to the provided ECR repository
func (b *Builder) build(ctx context.Context,
	ui terminal.UI,
	src *component.Source,
	log hclog.Logger) (*ociv1.BuildOutput, error) {
	// NOTE: we only use an ODR build because we need the `accessInfo` to be injected
	return nil, fmt.Errorf("only ODR builds supported by this plugin")
}

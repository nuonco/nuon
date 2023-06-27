package registry

import (
	"context"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/waypoint-plugin-sdk/component"
	"github.com/hashicorp/waypoint-plugin-sdk/terminal"
	ociv1 "github.com/powertoolsdev/mono/pkg/types/plugins/oci/v1"
)

func (r *Registry) PushFunc() interface{} {
	return r.Push
}

func (r *Registry) Push(
	ctx context.Context,
	log hclog.Logger,
	bld *ociv1.BuildOutput,
	ui terminal.UI,
	src *component.Source,
) (*ociv1.Artifact, error) {
	ui.Output("pushing noop build...")
	return &ociv1.Artifact{}, nil
}

package registry

import (
	"context"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/waypoint-plugin-sdk/component"
	"github.com/hashicorp/waypoint-plugin-sdk/terminal"
	noopv1 "github.com/powertoolsdev/mono/pkg/types/plugins/noop/v1"
)

func (r *Registry) PushFunc() interface{} {
	return r.Push
}

func (r *Registry) Push(
	ctx context.Context,
	log hclog.Logger,
	bld *noopv1.BuildOutput,
	ui terminal.UI,
	src *component.Source,
) (*noopv1.Artifact, error) {
	ui.Output("pushing noop build...")
	return &noopv1.Artifact{}, nil
}

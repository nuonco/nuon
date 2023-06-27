package registry

import (
	"context"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/waypoint-plugin-sdk/component"
	"github.com/hashicorp/waypoint-plugin-sdk/terminal"
	ociv1 "github.com/powertoolsdev/mono/pkg/types/plugins/oci/v1"
)

// AccessInfoFunc
func (r *Registry) AccessInfoFunc() interface{} {
	return r.AccessInfo
}

func (r *Registry) AccessInfo(ctx context.Context,
	log hclog.Logger,
	ui terminal.UI,
	src *component.Source,
) (*ociv1.AccessInfo, error) {
	ui.Output("fetching noop access info...")
	return &ociv1.AccessInfo{}, nil
}

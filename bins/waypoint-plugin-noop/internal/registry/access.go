package registry

import (
	"context"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/waypoint-plugin-sdk/component"
	"github.com/hashicorp/waypoint-plugin-sdk/terminal"
	noopv1 "github.com/powertoolsdev/mono/pkg/types/plugins/noop/v1"
)

// AccessInfoFunc
func (r *Registry) AccessInfoFunc() interface{} {
	return r.AccessInfo
}

func (r *Registry) AccessInfo(ctx context.Context,
	log hclog.Logger,
	ui terminal.UI,
	src *component.Source,
) (*noopv1.AccessInfo, error) {
	ui.Output("fetching noop access info...")
	return &noopv1.AccessInfo{}, nil
}

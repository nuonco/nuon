package registry

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/waypoint-plugin-sdk/component"
	"github.com/hashicorp/waypoint-plugin-sdk/terminal"
	"github.com/powertoolsdev/mono/pkg/plugins/configs"
	ociv1 "github.com/powertoolsdev/mono/pkg/types/plugins/oci/v1"
)

const (
	defaultRoleSessionName string = "waypoint-plugin-oci"
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
	var (
		accessInfo *ociv1.AccessInfo
		err        error
	)

	switch r.config.RegistryType {
	case configs.OCIRegistryTypeACR:
		accessInfo, err = r.getACR(ctx)
	case configs.OCIRegistryTypeECR:
		accessInfo, err = r.getECR(ctx)
	default:
		return nil, fmt.Errorf("unsupported registry type: %s", r.config.RegistryType)
	}

	if err != nil {
		return nil, fmt.Errorf("unable to get access info: %w", err)
	}

	return accessInfo, nil
}

package registry

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/pkg/azure/acr"
	ociv1 "github.com/powertoolsdev/mono/pkg/types/plugins/oci/v1"
)

func (r *Registry) getACR(ctx context.Context) (*ociv1.AccessInfo, error) {
	token, err := acr.GetRepositoryToken(ctx, r.config.ACRAuth, r.config.LoginServer)
	if err != nil {
		return nil, fmt.Errorf("unable to get acr token: %w", err)
	}

	return &ociv1.AccessInfo{
		Image: r.config.Repository,
		Tag:   r.config.Tag,
		Auth: &ociv1.Auth{
			Username:      acr.DefaultACRUsername,
			Password:      token,
			ServerAddress: r.config.LoginServer,
		},
	}, nil
}

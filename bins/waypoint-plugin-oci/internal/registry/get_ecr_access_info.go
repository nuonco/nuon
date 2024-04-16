package registry

import (
	"context"
	"fmt"
	"strings"

	ecrauthorization "github.com/powertoolsdev/mono/pkg/aws/ecr-authorization"
	ociv1 "github.com/powertoolsdev/mono/pkg/types/plugins/oci/v1"
)

func (r *Registry) getECR(ctx context.Context) (*ociv1.AccessInfo, error) {
	authProvider, err := ecrauthorization.New(r.v,
		ecrauthorization.WithCredentials(r.config.ECRAuth),
		ecrauthorization.WithUseDefault(true),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to get auth provider: %w", err)
	}

	accessInfo, err := r.getECRAccessInfo(ctx, authProvider)
	if err != nil {
		return nil, fmt.Errorf("unable to get access info: %w", err)
	}

	return accessInfo, nil
}

func (r *Registry) getECRAccessInfo(ctx context.Context, client ecrauthorization.Client) (*ociv1.AccessInfo, error) {
	authorization, err := client.GetAuthorization(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get authorization: %w", err)
	}

	serverAddr := strings.TrimPrefix(authorization.ServerAddress, "https://")
	return &ociv1.AccessInfo{
		Image: r.config.Repository,
		Tag:   r.config.Tag,
		Auth: &ociv1.Auth{
			Username:      authorization.Username,
			Password:      authorization.RegistryToken,
			ServerAddress: serverAddr,
		},
	}, nil
}

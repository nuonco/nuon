package ecr

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"

	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/registry"
	ecrauthorization "github.com/powertoolsdev/mono/pkg/aws/ecr-authorization"
	"github.com/powertoolsdev/mono/pkg/plugins/configs"
)

func FetchAccessInfo(ctx context.Context, cfg *configs.OCIRegistryRepository) (*registry.AccessInfo, error) {
	authProvider, err := ecrauthorization.New(validator.New(),
		ecrauthorization.WithCredentials(cfg.ECRAuth),
		ecrauthorization.WithRepository(cfg.Repository),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to get auth provider: %w", err)
	}

	authorization, err := authProvider.GetAuthorization(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get authorization: %w", err)
	}

	return &registry.AccessInfo{
		Image: cfg.Repository,
		Auth: &registry.AccessInfoAuth{
			Username:      authorization.Username,
			Password:      authorization.RegistryToken,
			ServerAddress: authorization.ServerAddress,
		},
	}, nil
}

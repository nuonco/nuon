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
		ecrauthorization.WithUseDefault(true),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to get auth provider: %w", err)
	}

	authorization, err := authProvider.GetAuthorization(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get authorization: %w", err)
	}

	// trim the server address from the repository
	repoName, err := ecrauthorization.TrimRepositoryName(cfg.Repository, authorization.ServerAddress)
	if err != nil {
		return nil, fmt.Errorf("unable to trim repository name: %w", err)
	}

	// return the registry
	return &registry.AccessInfo{
		Image: repoName,
		Auth: &registry.AccessInfoAuth{
			Username:      authorization.Username,
			Password:      authorization.RegistryToken,
			ServerAddress: authorization.ServerAddress,
		},
	}, nil
}

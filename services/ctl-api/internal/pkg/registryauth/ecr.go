package registryauth

import (
	"context"
	"fmt"
	"strings"

	awscredentials "github.com/nuonco/nuon/pkg/aws/credentials"
	ecrauthorization "github.com/nuonco/nuon/pkg/aws/ecr-authorization"
)

func (i *Issuer) ECR(ctx context.Context, creds *awscredentials.Config) (*Token, error) {
	authorizer, err := ecrauthorization.New(i.v, ecrauthorization.WithCredentials(creds))
	if err != nil {
		return nil, fmt.Errorf("unable to build ecr authorizer: %w", err)
	}

	auth, err := authorizer.GetAuthorization(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get ecr authorization: %w", err)
	}

	return &Token{
		Username:      auth.Username,
		Password:      auth.RegistryToken,
		ServerAddress: strings.TrimPrefix(auth.ServerAddress, "https://"),
	}, nil
}

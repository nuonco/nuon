package ecr

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/ecr"
	ecr_types "github.com/aws/aws-sdk-go-v2/service/ecr/types"
	"github.com/powertoolsdev/mono/pkg/aws/credentials"
)

//
//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=authorization_mock_test.go -source=authorization.go -package=ecr
type Authorization struct {
	RegistryToken string `validate:"required"`
	Username      string `validate:"required"`
	ServerAddress string `validate:"required"`
}

func (e *ecrAuthorizer) GetAuthorization(ctx context.Context) (*Authorization, error) {
	cfg, err := credentials.Fetch(ctx, e.Credentials)
	if err != nil {
		return nil, fmt.Errorf("unable to get credentials: %w", err)
	}

	ecrClient := ecr.NewFromConfig(cfg)
	authData, err := e.getAuthorizationData(ctx, ecrClient)
	if err != nil {
		return nil, fmt.Errorf("unable to get ecr authorization token: %w", err)
	}

	return e.parseAuthorizationData(authData)
}

// parseAuthorizationData: parses authorization data into the required return format
func (e *ecrAuthorizer) parseAuthorizationData(data *ecr_types.AuthorizationData) (*Authorization, error) {
	auth, err := base64.StdEncoding.DecodeString(*data.AuthorizationToken)
	if err != nil {
		return nil, fmt.Errorf("unable to decode auth string: %w", err)
	}

	authPieces := strings.SplitN(string(auth), ":", 2)
	return &Authorization{
		RegistryToken: authPieces[1],
		Username:      authPieces[0],
		ServerAddress: *data.ProxyEndpoint,
	}, nil
}

type awsECRClient interface {
	GetAuthorizationToken(context.Context, *ecr.GetAuthorizationTokenInput, ...func(*ecr.Options)) (*ecr.GetAuthorizationTokenOutput, error)
}

// getAuthorizationData: returns authentication data for connecting to an ECR repo
func (e *ecrAuthorizer) getAuthorizationData(ctx context.Context, client awsECRClient) (*ecr_types.AuthorizationData, error) {
	registryIDs := make([]string, 0)
	if e.RegistryID != "" {
		registryIDs = append(registryIDs, e.RegistryID)
	}
	params := &ecr.GetAuthorizationTokenInput{
		RegistryIds: registryIDs,
	}

	resp, err := client.GetAuthorizationToken(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("unable to get authorization token: %w", err)
	}

	if len(resp.AuthorizationData) < 1 {
		return nil, fmt.Errorf("invalid get authorization token response")
	}

	return &resp.AuthorizationData[0], nil
}

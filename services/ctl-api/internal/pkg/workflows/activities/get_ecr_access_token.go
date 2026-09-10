package activities

import (
	"context"
	"fmt"
	"strings"

	awscredentials "github.com/nuonco/nuon/pkg/aws/credentials"
	ecrauthorization "github.com/nuonco/nuon/pkg/aws/ecr-authorization"
)

type GetECRAccessTokenRequest struct {
	Credentials *awscredentials.Config
}

type ECRAccessToken struct {
	Username string
	Password string
	// Scheme-less, so it can prefix-match the repository URI.
	ServerAddress string
}

// GetECRAccessToken lives in the shared activity set for the same reason as its
// GAR and ACR counterparts: the installs namespace schedules it when resolving
// the sandbox artifact, and activities are registered per worker.
//
// @temporal-gen-v2 activity
// @max-retries 1
func (a *Activities) GetECRAccessToken(ctx context.Context, req *GetECRAccessTokenRequest) (*ECRAccessToken, error) {
	authorizer, err := ecrauthorization.New(a.v, ecrauthorization.WithCredentials(req.Credentials))
	if err != nil {
		return nil, fmt.Errorf("unable to build ecr authorizer: %w", err)
	}

	auth, err := authorizer.GetAuthorization(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get ecr authorization: %w", err)
	}

	return &ECRAccessToken{
		Username:      auth.Username,
		Password:      auth.RegistryToken,
		ServerAddress: strings.TrimPrefix(auth.ServerAddress, "https://"),
	}, nil
}

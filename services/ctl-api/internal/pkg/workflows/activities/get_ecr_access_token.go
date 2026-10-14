package activities

import (
	"context"

	awscredentials "github.com/nuonco/nuon/pkg/aws/credentials"
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

// GetECRAccessToken is in the shared set because activities register per worker
// and the installs namespace schedules it.
//
// @temporal-gen-v2 activity
// @max-retries 1
func (a *Activities) GetECRAccessToken(ctx context.Context, req *GetECRAccessTokenRequest) (*ECRAccessToken, error) {
	token, err := a.registryAuth.ECR(ctx, req.Credentials)
	if err != nil {
		return nil, err
	}

	return &ECRAccessToken{
		Username:      token.Username,
		Password:      token.Password,
		ServerAddress: token.ServerAddress,
	}, nil
}

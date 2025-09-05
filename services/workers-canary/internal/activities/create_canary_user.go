package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/pkg/api"
)

type CreateUserRequest struct {
	CanaryID string `validate:"required"`
}

type CreateUserResponse struct {
	APIToken        string
	GithubInstallID string
}

func (a *Activities) CreateUser(ctx context.Context, req *CreateUserRequest) (*CreateUserResponse, error) {
	internalAPIClient, err := api.New(a.v,
		api.WithURL(a.cfg.InternalAPIURL),
		api.WithAdminEmail("canary@serviceaccount.nuon.co"),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create internal api client: %w", err)
	}

	org, err := internalAPIClient.CreateCanaryUser(ctx, req.CanaryID)
	if err != nil {
		return nil, fmt.Errorf("unable to create integration user: %w", err)
	}

	return &CreateUserResponse{
		GithubInstallID: org.GithubInstallID,
		APIToken:        org.APIToken,
	}, nil
}

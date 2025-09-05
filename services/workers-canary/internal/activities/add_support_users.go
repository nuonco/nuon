package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/pkg/api"
)

type AddSupportUsersRequest struct {
	OrgID string `validate:"required"`
}

type AddSupportUsersResponse struct{}

func (a *Activities) AddSupportUsers(ctx context.Context, req *AddSupportUsersRequest) (*AddSupportUsersResponse, error) {
	internalAPIClient, err := api.New(a.v,
		api.WithURL(a.cfg.InternalAPIURL),
		api.WithAdminEmail("canary@serviceaccount.nuon.co"),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create internal api client: %w", err)
	}

	err = internalAPIClient.AddSupportUsers(ctx, req.OrgID)
	if err != nil {
		return nil, fmt.Errorf("unable to create support users for org: %w", err)
	}

	return &AddSupportUsersResponse{}, nil
}

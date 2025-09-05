package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/pkg/api"
)

type GetOrgRequest struct {
	CanaryID string `validate:"required"`
}

type GetOrgResponse struct {
	OrgID string
}

func (a *Activities) GetOrg(ctx context.Context, req *GetOrgRequest) (*GetOrgResponse, error) {
	internalAPIClient, err := api.New(a.v,
		api.WithURL(a.cfg.InternalAPIURL),
		api.WithAdminEmail("canary@serviceaccount.nuon.co"),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create internal api client: %w", err)
	}

	org, err := internalAPIClient.GetOrg(ctx, req.CanaryID)
	if err != nil {
		return nil, fmt.Errorf("unable to list orgs: %w", err)
	}

	return &GetOrgResponse{
		OrgID: org.Id,
	}, nil
}

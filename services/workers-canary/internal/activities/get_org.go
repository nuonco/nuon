package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/pkg/api"
)

type GetOrgRequest struct {
	CanaryID string
}

type GetOrgResponse struct {
	OrgID string
}

func (a *Activities) GetOrg(ctx context.Context, req *GetOrgRequest) (*GetOrgResponse, error) {
	internalAPIClient, err := api.New(a.v,
		api.WithURL(a.cfg.InternalAPIURL),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create internal api client: %w", err)
	}

	orgs, err := internalAPIClient.ListOrgs(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to create org: %w", err)
	}

	for _, org := range orgs {
		if org.Name == req.CanaryID {
			return &GetOrgResponse{
				OrgID: org.Id,
			}, nil
		}
	}

	return nil, fmt.Errorf("org %s not found", req.CanaryID)
}

package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/pkg/api"
)

type DeleteOrgRequest struct {
	CanaryID string
	OrgID    string
}

type DeleteOrgResponse struct {
	OrgID string
}

func (a *Activities) DeleteOrg(ctx context.Context, req *DeleteOrgRequest) (*DeleteOrgResponse, error) {
	internalAPIClient, err := api.New(a.v,
		api.WithURL(a.cfg.InternalAPIURL),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create internal api client: %w", err)
	}

	err = internalAPIClient.DeleteOrg(ctx, req.OrgID)
	if err != nil {
		return nil, fmt.Errorf("unable to create org: %w", err)
	}

	return &DeleteOrgResponse{
		OrgID: req.OrgID,
	}, nil
}

package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon-go"
)

type DeleteOrgRequest struct {
	CanaryID string
	OrgID	 string
}

type DeleteOrgResponse struct {
	OrgID string
}

func (a *Activities) DeleteOrg(ctx context.Context, req *DeleteOrgRequest) (*DeleteOrgResponse, error) {
	apiClient, err := nuon.New(a.v,
		nuon.WithURL(a.cfg.APIURL),
		nuon.WithAuthToken(a.cfg.APIToken),
		nuon.WithOrgID(req.OrgID),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create api client: %w", err)
	}

	_, err = apiClient.DeleteOrg(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to create org: %w", err)
	}

	return &DeleteOrgResponse{
		OrgID: req.OrgID,
	}, nil
}

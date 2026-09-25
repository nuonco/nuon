package nuon

import (
	"context"

	"github.com/nuonco/nuon/sdks/nuon-go/client/operations"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

type GetInstallDeploymentsQuery struct {
	Type         string
	Status       string
	Resource     string
	Search       string
	CreatedAtGte string
	CreatedAtLte string
	Limit        int
	Offset       int
}

func (c *client) GetInstallDeployments(ctx context.Context, installID string, query *GetInstallDeploymentsQuery) (*models.ServiceGetInstallDeploymentsResponse, error) {
	params := &operations.GetInstallDeploymentsParams{
		Context:   ctx,
		InstallID: installID,
	}

	limit := 20
	var offset int
	if query != nil {
		if query.Type != "" {
			params.Type = &query.Type
		}
		if query.Status != "" {
			params.Status = &query.Status
		}
		if query.Resource != "" {
			params.Resource = &query.Resource
		}
		if query.Search != "" {
			params.Search = &query.Search
		}
		if query.CreatedAtGte != "" {
			params.CreatedAtGte = &query.CreatedAtGte
		}
		if query.CreatedAtLte != "" {
			params.CreatedAtLte = &query.CreatedAtLte
		}
		if query.Limit > 0 {
			limit = query.Limit
		}
		offset = query.Offset
	}

	l := int64(limit)
	o := int64(offset)
	params.Limit = &l
	params.Offset = &o

	resp, err := c.genClient.Operations.GetInstallDeployments(params, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

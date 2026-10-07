package nuon

import (
	"context"

	"github.com/nuonco/nuon/sdks/nuon-go/client/operations"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

type GetInstallDeploymentSummariesQuery struct {
	Type         string
	Status       string
	Resource     string
	Search       string
	CreatedAtGte string
	CreatedAtLte string
	State        string
	Sort         string
	Cursor       string
	Limit        int
	Offset       int
}

func (c *client) GetInstallDeploymentSummaries(ctx context.Context, installID string, query *GetInstallDeploymentSummariesQuery) (*models.ServiceGetInstallDeploymentSummariesResponse, error) {
	params := &operations.GetInstallDeploymentSummariesParams{
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
		if query.State != "" {
			params.State = &query.State
		}
		if query.Sort != "" {
			params.Sort = &query.Sort
		}
		if query.Cursor != "" {
			params.Cursor = &query.Cursor
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

	resp, err := c.genClient.Operations.GetInstallDeploymentSummaries(params, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

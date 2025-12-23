package nuon

import (
	"context"

	"github.com/nuonco/nuon-go/client/operations"
	"github.com/nuonco/nuon-go/models"
)

func (c *client) GetInstallSandboxRuns(ctx context.Context, installID string, query *models.GetPaginatedQuery) ([]*models.AppInstallSandboxRun, bool, error) {
	params := &operations.GetInstallSandboxRunsParams{
		InstallID: installID,
		Context:   ctx,
	}

	query = handlePaginationQuery(query)

	if query != nil {
		offset := int64(query.Offset)
		limit := int64(query.Limit)
		params.Offset = &offset
		params.Limit = &limit
	}

	resp, err := c.genClient.Operations.GetInstallSandboxRuns(params, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, false, err
	}

	if query != nil {
		items, hasMore := handlePagination(resp.Payload, int64(query.Offset), int64(query.Limit))
		return items, hasMore, nil
	}

	return resp.Payload, false, nil
}

func (c *client) DeprovisionInstallSandbox(ctx context.Context, installID string) error {
	_, err := c.genClient.Operations.DeprovisionInstallSandbox(&operations.DeprovisionInstallSandboxParams{
		InstallID: installID,
		Context:   ctx,
		// TODO: make this configurable
		Req: &models.ServiceDeprovisionInstallSandboxRequest{
			PlanOnly: false,
		},
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return err
	}

	return nil
}

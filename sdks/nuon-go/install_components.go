package nuon

import (
	"context"

	"github.com/nuonco/nuon-go/client/operations"
	"github.com/nuonco/nuon-go/models"
)

type getinstallcomponentsquery struct {
	Offset int
	Limit  int
}

// install components
func (c *client) GetInstallComponents(ctx context.Context, installID string, query *models.GetPaginatedQuery) ([]*models.AppInstallComponent, bool, error) {
	params := &operations.GetInstallComponentsParams{
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

	resp, err := c.genClient.Operations.GetInstallComponents(params, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, false, err
	}

	if query != nil {
		items, hasMore := handlePagination(resp.Payload, int64(query.Offset), int64(query.Limit))
		return items, hasMore, nil
	}

	return resp.Payload, false, nil
}

func (c *client) GetInstallComponentDeploys(ctx context.Context, installID string, componentID string, query *models.GetPaginatedQuery) ([]*models.AppInstallDeploy, bool, error) {
	params := &operations.GetInstallComponentDeploysParams{
		InstallID:   installID,
		ComponentID: componentID,
		Context:     ctx,
	}

	query = handlePaginationQuery(query)

	if query != nil {
		offset := int64(query.Offset)
		limit := int64(query.Limit)
		params.Offset = &offset
		params.Limit = &limit
	}

	resp, err := c.genClient.Operations.GetInstallComponentDeploys(params, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, false, err
	}

	if query != nil {
		items, hasMore := handlePagination(resp.Payload, int64(query.Offset), int64(query.Limit))
		return items, hasMore, nil
	}

	return resp.Payload, false, nil
}

func (c *client) GetInstallComponentLatestDeploy(ctx context.Context, installID string, componentID string) (*models.AppInstallDeploy, error) {
	resp, err := c.genClient.Operations.GetInstallComponentLatestDeploy(&operations.GetInstallComponentLatestDeployParams{
		ComponentID: componentID,
		InstallID:   installID,
		Context:     ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) TeardownInstallComponent(ctx context.Context, installID, componentID string) error {
	resp, err := c.genClient.Operations.TeardownInstallComponent(&operations.TeardownInstallComponentParams{
		InstallID:   installID,
		ComponentID: componentID,
		Context:     ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return err
	}

	if resp.Payload != "ok" {
		return statusErr{resp.Payload}
	}

	return nil
}

func (c *client) TeardownInstallComponents(ctx context.Context, installID string) error {
	resp, err := c.genClient.Operations.TeardownInstallComponents(&operations.TeardownInstallComponentsParams{
		InstallID: installID,
		Context:   ctx,
		// TODO(jm): make this configurable
		Req: &models.ServiceTeardownInstallComponentsRequest{},
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return err
	}

	if resp.Payload != "ok" {
		return statusErr{resp.Payload}
	}

	return nil
}

func (c *client) DeployInstallComponents(ctx context.Context, installID string, planOnly bool) error {
	resp, err := c.genClient.Operations.DeployInstallComponents(&operations.DeployInstallComponentsParams{
		InstallID: installID,
		Context:   ctx,
		Req: &models.ServiceDeployInstallComponentsRequest{
			PlanOnly: planOnly,
		},
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return err
	}

	if resp.Payload != "ok" {
		return statusErr{resp.Payload}
	}

	return nil
}

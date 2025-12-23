package nuon

import (
	"context"

	"github.com/nuonco/nuon-go/client/operations"
	"github.com/nuonco/nuon-go/models"
)

// releases methods
func (c *client) GetRelease(ctx context.Context, releaseID string) (*models.AppComponentRelease, error) {
	resp, err := c.genClient.Operations.GetRelease(&operations.GetReleaseParams{
		ReleaseID: releaseID,
		Context:   ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) GetReleaseSteps(ctx context.Context, releaseID string, query *models.GetPaginatedQuery) ([]*models.AppComponentReleaseStep, bool, error) {
	params := &operations.GetReleaseStepsParams{
		ReleaseID: releaseID,
		Context:   ctx,
	}

	query = handlePaginationQuery(query)

	if query != nil {
		offset := int64(query.Offset)
		limit := int64(query.Limit)
		params.Offset = &offset
		params.Limit = &limit
	}

	resp, err := c.genClient.Operations.GetReleaseSteps(params, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, false, err
	}

	if query != nil {
		items, hasMore := handlePagination(resp.Payload, int64(query.Offset), int64(query.Limit))
		return items, hasMore, nil
	}

	return resp.Payload, false, nil
}

func (c *client) GetAppReleases(ctx context.Context, appID string, query *models.GetPaginatedQuery) ([]*models.AppComponentRelease, bool, error) {
	params := &operations.GetAppReleasesParams{
		AppID:   appID,
		Context: ctx,
	}

	query = handlePaginationQuery(query)

	if query != nil {
		offset := int64(query.Offset)
		limit := int64(query.Limit)
		params.Offset = &offset
		params.Limit = &limit
	}

	resp, err := c.genClient.Operations.GetAppReleases(params, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, false, err
	}

	if query != nil {
		items, hasMore := handlePagination(resp.Payload, int64(query.Offset), int64(query.Limit))
		return items, hasMore, nil
	}

	return resp.Payload, false, nil
}

func (c *client) GetComponentReleases(ctx context.Context, componentID string, query *models.GetPaginatedQuery) ([]*models.AppComponentRelease, bool, error) {
	params := &operations.GetComponentReleasesParams{
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

	resp, err := c.genClient.Operations.GetComponentReleases(params, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, false, err
	}

	if query != nil {
		items, hasMore := handlePagination(resp.Payload, int64(query.Offset), int64(query.Limit))
		return items, hasMore, nil
	}

	return resp.Payload, false, nil
}

func (c *client) CreateComponentRelease(ctx context.Context, componentID string, req *models.ServiceCreateComponentReleaseRequest) (*models.AppComponentRelease, error) {
	resp, err := c.genClient.Operations.CreateComponentRelease(&operations.CreateComponentReleaseParams{
		ComponentID: componentID,
		Req:         req,
		Context:     ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

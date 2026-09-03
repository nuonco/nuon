package nuon

import (
	"context"

	"github.com/nuonco/nuon/sdks/nuon-go/client/operations"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func (c *client) CreateAppRelease(ctx context.Context, appID, appConfigID string) (*models.AppAppRelease, error) {
	ok, created, err := c.genClient.Operations.CreateAppRelease(&operations.CreateAppReleaseParams{
		AppID: appID, Context: ctx,
		Request: &models.ServiceCreateReleaseRequest{AppConfigID: &appConfigID},
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}
	if ok != nil {
		return ok.Payload, nil
	}
	return created.Payload, nil
}

func (c *client) ListAppReleases(ctx context.Context, appID string, query *models.GetPaginatedQuery) ([]*models.AppAppRelease, bool, error) {
	params := &operations.ListAppReleasesParams{AppID: appID, Context: ctx}
	params.Offset, params.Limit = applyPaginationQuery(query)
	hr := newResponseHeaderReader(&operations.ListAppReleasesReader{})
	resp, err := c.genClient.Operations.ListAppReleases(params, c.getOrgIDAuthInfo(), hr.ClientOption())
	if err != nil {
		return nil, false, err
	}
	return resp.Payload, hasNextPage(hr), nil
}

func (c *client) GetAppRelease(ctx context.Context, appID, releaseID string) (*models.AppAppRelease, error) {
	resp, err := c.genClient.Operations.GetAppRelease(&operations.GetAppReleaseParams{
		AppID: appID, ReleaseID: releaseID, Context: ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}
	return resp.Payload, nil
}

package nuon

import (
	"context"

	"github.com/nuonco/nuon/sdks/nuon-go/client/operations"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

type GetAppBundlesQuery struct {
	Pagination  models.GetPaginatedQuery
	AppConfigID string
	Status      string
}

func (c *client) GetAppBundles(ctx context.Context, appID string, query *GetAppBundlesQuery) ([]*models.ServiceBundleResponse, bool, error) {
	params := &operations.GetAppBundlesParams{Context: ctx, AppID: appID}
	if query != nil {
		params.Offset, params.Limit = applyPaginationQuery(&query.Pagination)
		if query.AppConfigID != "" {
			params.AppConfigID = &query.AppConfigID
		}
		if query.Status != "" {
			params.Status = &query.Status
		}
	}
	hr := newResponseHeaderReader(&operations.GetAppBundlesReader{})
	resp, err := c.genClient.Operations.GetAppBundles(params, c.getOrgIDAuthInfo(), hr.ClientOption())
	if err != nil {
		return nil, false, err
	}
	return resp.Payload, hasNextPage(hr), nil
}

func (c *client) GetAppBundle(ctx context.Context, appID, bundleID string) (*models.ServiceBundleResponse, error) {
	resp, err := c.genClient.Operations.GetAppBundle(&operations.GetAppBundleParams{
		Context: ctx, AppID: appID, BundleID: bundleID,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}
	return resp.Payload, nil
}

func (c *client) CreateAppBundle(ctx context.Context, appID string, req *models.ServiceCreateBundleRequest) (*models.ServiceBundleResponse, error) {
	active, accepted, err := c.genClient.Operations.CreateAppBundle(&operations.CreateAppBundleParams{
		Context: ctx, AppID: appID, Request: req,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}
	if active != nil {
		return active.Payload, nil
	}
	return accepted.Payload, nil
}

func (c *client) CreateAppBundleDownloadGrant(ctx context.Context, appID, bundleID string) (*models.ServiceDownloadGrantResponse, error) {
	resp, err := c.genClient.Operations.CreateAppBundleDownloadGrant(&operations.CreateAppBundleDownloadGrantParams{
		Context: ctx, AppID: appID, BundleID: bundleID,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}
	return resp.Payload, nil
}

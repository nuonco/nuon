package nuon

import (
	"context"

	"github.com/nuonco/nuon/sdks/nuon-go/client/operations"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func (c *client) CreateCustomerManagedBundle(ctx context.Context, appID, appConfigID, targetPlatform string) (*models.ServiceBundleResponse, error) {
	return c.CreateCustomerManagedBundleWithRequest(ctx, appID, &models.ServiceCreateBundleRequest{AppConfigID: &appConfigID, TargetPlatform: targetPlatform})
}

func (c *client) CreateCustomerManagedBundleWithRequest(ctx context.Context, appID string, req *models.ServiceCreateBundleRequest) (*models.ServiceBundleResponse, error) {
	ok, accepted, err := c.genClient.Operations.CreateCustomerManagedBundle(&operations.CreateCustomerManagedBundleParams{
		AppID: appID, Context: ctx,
		Request: req,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}
	if ok != nil {
		return ok.Payload, nil
	}
	return accepted.Payload, nil
}

func (c *client) ListCustomerManagedBundles(ctx context.Context, appID string, query *models.GetPaginatedQuery) ([]*models.ServiceBundleResponse, bool, error) {
	params := &operations.ListCustomerManagedBundlesParams{
		AppID:   appID,
		Context: ctx,
	}
	params.Offset, params.Limit = applyPaginationQuery(query)

	hr := newResponseHeaderReader(&operations.ListCustomerManagedBundlesReader{})
	resp, err := c.genClient.Operations.ListCustomerManagedBundles(params, c.getOrgIDAuthInfo(), hr.ClientOption())
	if err != nil {
		return nil, false, err
	}

	return resp.Payload, hasNextPage(hr), nil
}

func (c *client) GetCustomerManagedBundle(ctx context.Context, appID, bundleID string) (*models.ServiceBundleResponse, error) {
	resp, err := c.genClient.Operations.GetCustomerManagedBundle(&operations.GetCustomerManagedBundleParams{
		AppID:    appID,
		BundleID: bundleID,
		Context:  ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) CreateCustomerManagedBundleBlobGrants(ctx context.Context, appID, bundleID string, digests []string) (*models.ServiceBlobGrantsResponse, error) {
	resp, err := c.genClient.Operations.CreateCustomerManagedBundleBlobGrants(&operations.CreateCustomerManagedBundleBlobGrantsParams{
		AppID:    appID,
		BundleID: bundleID,
		Context:  ctx,
		Req:      &models.ServiceBlobGrantsRequest{Digests: digests},
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) CreateCustomerManagedBundleDownloadGrant(ctx context.Context, appID, bundleID string) (*models.ServiceDownloadGrantResponse, error) {
	resp, err := c.genClient.Operations.CreateCustomerManagedBundleDownloadGrant(&operations.CreateCustomerManagedBundleDownloadGrantParams{
		AppID:    appID,
		BundleID: bundleID,
		Context:  ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

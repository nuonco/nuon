package nuon

import (
	"context"

	"github.com/nuonco/nuon/sdks/nuon-go/client/operations"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func (c *client) CreateAirgapBundle(ctx context.Context, appID, appConfigID, targetPlatform string) (*models.ServiceBundleResponse, error) {
	ok, accepted, err := c.genClient.Operations.CreateAirgapBundle(&operations.CreateAirgapBundleParams{
		AppID: appID, Context: ctx,
		Request: &models.ServiceCreateBundleRequest{AppConfigID: &appConfigID, TargetPlatform: targetPlatform},
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}
	if ok != nil {
		return ok.Payload, nil
	}
	return accepted.Payload, nil
}

func (c *client) ListAirgapBundles(ctx context.Context, appID string, query *models.GetPaginatedQuery) ([]*models.ServiceBundleResponse, bool, error) {
	params := &operations.ListAirgapBundlesParams{
		AppID:   appID,
		Context: ctx,
	}
	params.Offset, params.Limit = applyPaginationQuery(query)

	hr := newResponseHeaderReader(&operations.ListAirgapBundlesReader{})
	resp, err := c.genClient.Operations.ListAirgapBundles(params, c.getOrgIDAuthInfo(), hr.ClientOption())
	if err != nil {
		return nil, false, err
	}

	return resp.Payload, hasNextPage(hr), nil
}

func (c *client) GetAirgapBundle(ctx context.Context, appID, bundleID string) (*models.ServiceBundleResponse, error) {
	resp, err := c.genClient.Operations.GetAirgapBundle(&operations.GetAirgapBundleParams{
		AppID:    appID,
		BundleID: bundleID,
		Context:  ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) CreateAirgapInstall(ctx context.Context, appID, bundleID, name string) (*models.ServiceAirgapInstallResponse, error) {
	resp, err := c.genClient.Operations.CreateAirgapInstall(&operations.CreateAirgapInstallParams{
		AppID:    appID,
		BundleID: bundleID,
		Context:  ctx,
		Request:  &models.ServiceCreateAirgapInstallRequest{Name: &name},
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) ListAirgapInstalls(ctx context.Context, appID, bundleID string, query *models.GetPaginatedQuery) ([]*models.ServiceAirgapInstallResponse, bool, error) {
	params := &operations.ListAirgapInstallsParams{
		AppID:    appID,
		BundleID: bundleID,
		Context:  ctx,
	}
	params.Offset, params.Limit = applyPaginationQuery(query)

	hr := newResponseHeaderReader(&operations.ListAirgapInstallsReader{})
	resp, err := c.genClient.Operations.ListAirgapInstalls(params, c.getOrgIDAuthInfo(), hr.ClientOption())
	if err != nil {
		return nil, false, err
	}

	return resp.Payload, hasNextPage(hr), nil
}

func (c *client) CreateAirgapBundleDownloadGrant(ctx context.Context, appID, bundleID string) (*models.ServiceDownloadGrantResponse, error) {
	resp, err := c.genClient.Operations.CreateAirgapBundleDownloadGrant(&operations.CreateAirgapBundleDownloadGrantParams{
		AppID:    appID,
		BundleID: bundleID,
		Context:  ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

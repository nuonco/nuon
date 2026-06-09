package nuon

import (
	"bytes"
	"context"
	"fmt"

	"github.com/nuonco/nuon/sdks/nuon-go/client/operations"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

// installs
func (c *client) CreateInstall(ctx context.Context, appID string, req *models.ServiceCreateInstallRequest) (*models.AppInstall, error) {
	resp, err := c.genClient.Operations.CreateInstall(&operations.CreateInstallParams{
		AppID:   appID,
		Req:     req,
		Context: ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) GetAppInstalls(ctx context.Context, appID string, query *models.GetPaginatedQuery) ([]*models.AppInstall, bool, error) {
	params := &operations.GetAppInstallsParams{
		AppID:   appID,
		Context: ctx,
	}

	params.Offset, params.Limit = applyPaginationQuery(query)

	hr := newResponseHeaderReader(&operations.GetAppInstallsReader{})
	resp, err := c.genClient.Operations.GetAppInstalls(params, c.getOrgIDAuthInfo(), hr.ClientOption())
	if err != nil {
		return nil, false, err
	}

	return resp.Payload, hasNextPage(hr), nil
}

func (c *client) GetAllInstalls(ctx context.Context, query *models.GetPaginatedQuery) ([]*models.AppInstall, bool, error) {
	params := &operations.GetOrgInstallsParams{
		Context: ctx,
	}

	params.Offset, params.Limit = applyPaginationQuery(query)

	hr := newResponseHeaderReader(&operations.GetOrgInstallsReader{})
	resp, err := c.genClient.Operations.GetOrgInstalls(params, c.getOrgIDAuthInfo(), hr.ClientOption())
	if err != nil {
		return nil, false, err
	}

	return resp.Payload, hasNextPage(hr), nil
}

func (c *client) GetInstall(ctx context.Context, installID string) (*models.AppInstall, error) {
	resp, err := c.genClient.Operations.GetInstall(&operations.GetInstallParams{
		InstallID: installID,
		Context:   ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) UpdateInstall(ctx context.Context, installID string, req *models.ServiceUpdateInstallRequest) (*models.AppInstall, error) {
	resp, err := c.genClient.Operations.UpdateInstall(&operations.UpdateInstallParams{
		Req:       req,
		InstallID: installID,
		Context:   ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, fmt.Errorf("unable to update install: %w", err)
	}

	return resp.Payload, nil
}

func (c *client) DeleteInstall(ctx context.Context, installID string) (*models.AppWorkflowResponse, error) {
	resp, err := c.genClient.Operations.DeleteInstall(&operations.DeleteInstallParams{
		InstallID: installID,
		Context:   ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) ForgetInstall(ctx context.Context, installID string) (bool, error) {
	_, err := c.genClient.Operations.ForgetInstall(&operations.ForgetInstallParams{
		InstallID: installID,
		Context:   ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return false, err
	}

	return true, nil
}

func (c *client) ReprovisionInstall(ctx context.Context, installID string) (*models.AppWorkflowResponse, error) {
	resp, err := c.genClient.Operations.ReprovisionInstall(&operations.ReprovisionInstallParams{
		InstallID: installID,
		Context:   ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) DeprovisionInstall(ctx context.Context, installID string) (*models.AppWorkflowResponse, error) {
	resp, err := c.genClient.Operations.DeprovisionInstall(&operations.DeprovisionInstallParams{
		InstallID: installID,
		Context:   ctx,
		// TODO(jm): make this configurable
		Req: &models.ServiceDeprovisionInstallRequest{},
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) AddInstallLabels(ctx context.Context, installID string, labels map[string]string) (*models.AppInstall, error) {
	resp, err := c.genClient.Operations.AddInstallLabels(&operations.AddInstallLabelsParams{
		InstallID: installID,
		Req:       &models.ServiceAddInstallLabelsRequest{Labels: labels},
		Context:   ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}
	return resp.Payload, nil
}

func (c *client) RemoveInstallLabels(ctx context.Context, installID string, keys []string) (*models.AppInstall, error) {
	resp, err := c.genClient.Operations.RemoveInstallLabels(&operations.RemoveInstallLabelsParams{
		InstallID: installID,
		Req:       &models.ServiceRemoveInstallLabelsRequest{Keys: keys},
		Context:   ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}
	return resp.Payload, nil
}

func (c *client) GenerateCLIInstallConfig(ctx context.Context, installID string) ([]byte, error) {
	var buf bytes.Buffer
	if _, err := c.genClient.Operations.GenerateCLIInstallConfig(&operations.GenerateCLIInstallConfigParams{
		InstallID: installID,
		Context:   ctx,
	}, c.getOrgIDAuthInfo(), &buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

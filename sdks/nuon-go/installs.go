package nuon

import (
	"bytes"
	"context"
	"fmt"

	"github.com/nuonco/nuon-go/client/operations"
	"github.com/nuonco/nuon-go/models"
)

const (
	HeaderInstallWorkflowID = "X-Nuon-Install-Workflow-ID"
)

// installs
func (c *client) CreateInstall(ctx context.Context, appID string, req *models.ServiceCreateInstallRequest) (*models.AppInstall, string, error) {
	hr := newResponseHeaderReader(&operations.CreateInstallReader{})

	resp, err := c.genClient.Operations.CreateInstall(&operations.CreateInstallParams{
		AppID:   appID,
		Req:     req,
		Context: ctx,
	}, c.getOrgIDAuthInfo(), hr.ClientOption())
	if err != nil {
		return nil, "", err
	}

	installWOrkflowID := hr.GetHeader(HeaderInstallWorkflowID)

	return resp.Payload, installWOrkflowID, nil
}

func (c *client) GetAppInstalls(ctx context.Context, appID string, query *models.GetPaginatedQuery) ([]*models.AppInstall, bool, error) {
	params := &operations.GetAppInstallsParams{
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

	resp, err := c.genClient.Operations.GetAppInstalls(params, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, false, err
	}

	if query != nil {
		items, hasMore := handlePagination(resp.Payload, int64(query.Offset), int64(query.Limit))
		return items, hasMore, nil
	}

	return resp.Payload, false, nil
}

func (c *client) GetAllInstalls(ctx context.Context, query *models.GetPaginatedQuery) ([]*models.AppInstall, bool, error) {
	params := &operations.GetOrgInstallsParams{
		Context: ctx,
	}

	query = handlePaginationQuery(query)

	if query != nil {
		offset := int64(query.Offset)
		limit := int64(query.Limit)
		params.Offset = &offset
		params.Limit = &limit
	}

	resp, err := c.genClient.Operations.GetOrgInstalls(params, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, false, err
	}

	if query != nil {
		items, hasMore := handlePagination(resp.Payload, int64(query.Offset), int64(query.Limit))
		return items, hasMore, nil
	}

	return resp.Payload, false, nil
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

func (c *client) DeleteInstall(ctx context.Context, installID string) (bool, error) {
	resp, err := c.genClient.Operations.DeleteInstall(&operations.DeleteInstallParams{
		InstallID: installID,
		Context:   ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return false, err
	}

	return resp.IsSuccess(), nil
}

func (c *client) ForgetInstall(ctx context.Context, installID string) (bool, error) {
	resp, err := c.genClient.Operations.ForgetInstall(&operations.ForgetInstallParams{
		InstallID: installID,
		Context:   ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return false, err
	}

	return resp.IsSuccess(), nil
}

func (c *client) ReprovisionInstall(ctx context.Context, installID string) error {
	resp, err := c.genClient.Operations.ReprovisionInstall(&operations.ReprovisionInstallParams{
		InstallID: installID,
		Context:   ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return err
	}

	if resp.Payload != "ok" {
		return statusErr{resp.Payload}
	}

	return nil
}

func (c *client) DeprovisionInstall(ctx context.Context, installID string) error {
	resp, err := c.genClient.Operations.DeprovisionInstall(&operations.DeprovisionInstallParams{
		InstallID: installID,
		Context:   ctx,
		// TODO(jm): make this configurable
		Req: &models.ServiceDeprovisionInstallRequest{},
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return err
	}

	if resp.Payload != "ok" {
		return statusErr{resp.Payload}
	}

	return nil
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

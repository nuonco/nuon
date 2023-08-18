package client

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/pkg/api/client/client/operations"
	"github.com/powertoolsdev/mono/pkg/api/client/models"
)

// installs
func (c *client) CreateInstall(ctx context.Context, appID string, req *models.ServiceCreateInstallRequest) (*models.AppInstall, error) {
	resp, err := c.genClient.Operations.PostV1AppsAppIDInstalls(&operations.PostV1AppsAppIDInstallsParams{
		AppID:   appID,
		Req:     req,
		Context: ctx,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create app install: %w", err)
	}

	return resp.Payload, nil
}

func (c *client) GetAppInstalls(ctx context.Context, appID string) ([]*models.AppInstall, error) {
	resp, err := c.genClient.Operations.GetV1AppsAppIDInstalls(&operations.GetV1AppsAppIDInstallsParams{
		AppID:   appID,
		Context: ctx,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to get app installs: %w", err)
	}

	return resp.Payload, nil
}

func (c *client) GetAllInstalls(ctx context.Context) ([]*models.AppInstall, error) {
	resp, err := c.genClient.Operations.GetV1Installs(&operations.GetV1InstallsParams{
		Context: ctx,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to get installs: %w", err)
	}

	return resp.Payload, nil
}

func (c *client) GetInstall(ctx context.Context, installID string) (*models.AppInstall, error) {
	resp, err := c.genClient.Operations.GetV1InstallsInstallID(&operations.GetV1InstallsInstallIDParams{
		InstallID: installID,
		Context:   ctx,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to get install: %w", err)
	}

	return resp.Payload, nil
}

func (c *client) UpdateInstall(ctx context.Context, installID string, req *models.ServiceUpdateInstallRequest) (*models.AppInstall, error) {
	resp, err := c.genClient.Operations.PatchV1InstallsInstallID(&operations.PatchV1InstallsInstallIDParams{
		Req:       req,
		InstallID: installID,
		Context:   ctx,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to update install: %w", err)
	}

	return resp.Payload, nil
}

func (c *client) DeleteInstall(ctx context.Context, installID string) (bool, error) {
	resp, err := c.genClient.Operations.DeleteV1InstallsInstallID(&operations.DeleteV1InstallsInstallIDParams{
		InstallID: installID,
		Context:   ctx,
	})
	if err != nil {
		return false, fmt.Errorf("unable to delete install: %w", err)
	}

	return resp.IsSuccess(), nil
}

// install deploys
func (c *client) GetInstallDeploys(ctx context.Context, installID string) ([]*models.AppInstallDeploy, error) {
	resp, err := c.genClient.Operations.GetV1InstallsInstallIDDeploys(&operations.GetV1InstallsInstallIDDeploysParams{
		InstallID: installID,
		Context:   ctx,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to get install deploys: %w", err)
	}

	return resp.Payload, nil
}

func (c *client) CreateInstallDeploy(ctx context.Context, installID string, req *models.ServiceCreateInstallDeployRequest) (*models.AppInstallDeploy, error) {
	resp, err := c.genClient.Operations.PostV1InstallsInstallIDDeploys(&operations.PostV1InstallsInstallIDDeploysParams{
		Req:     req,
		Context: ctx,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create install deploy: %w", err)
	}

	return resp.Payload, nil
}

func (c *client) GetInstallDeploy(ctx context.Context, installID string, deployID string) (*models.AppInstallDeploy, error) {
	resp, err := c.genClient.Operations.GetV1InstallsInstallIDDeploysDeployID(&operations.GetV1InstallsInstallIDDeploysDeployIDParams{
		InstallID: installID,
		DeployID:  deployID,
		Context:   ctx,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create install deploy: %w", err)
	}

	return resp.Payload, nil
}

func (c *client) GetInstallDeployLogs(ctx context.Context, installID string, deployID string) ([]models.ServiceDeployLog, error) {
	resp, err := c.genClient.Operations.GetV1InstallsInstallIDDeploysDeployIDLogs(&operations.GetV1InstallsInstallIDDeploysDeployIDLogsParams{
		InstallID: installID,
		DeployID:  deployID,
		Context:   ctx,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to get install deploy logs: %w", err)
	}

	return resp.Payload, nil
}

// install components
func (c *client) GetInstallComponents(ctx context.Context, installID string) ([]*models.AppInstallComponent, error) {
	resp, err := c.genClient.Operations.GetV1InstallsInstallIDComponents(&operations.GetV1InstallsInstallIDComponentsParams{
		InstallID: installID,
		Context:   ctx,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to get install components: %w", err)
	}

	return resp.Payload, nil
}

func (c *client) GetInstallComponentDeploys(ctx context.Context, installID string, componentID string) ([]*models.AppInstallDeploy, error) {
	resp, err := c.genClient.Operations.GetV1InstallsInstallIDComponentsComponentIDDeploys(&operations.GetV1InstallsInstallIDComponentsComponentIDDeploysParams{
		ComponentID: componentID,
		InstallID:   installID,
		Context:     ctx,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to get install components: %w", err)
	}

	return resp.Payload, nil
}

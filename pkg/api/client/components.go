package client

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/pkg/api/client/client/operations"
	"github.com/powertoolsdev/mono/pkg/api/client/models"
)

// components
func (c *client) GetAllComponents(ctx context.Context) ([]*models.AppComponent, error) {
	resp, err := c.genClient.Operations.GetV1Components(&operations.GetV1ComponentsParams{
		Context: ctx,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to get components: %w", err)
	}

	return resp.Payload, nil
}

func (c *client) GetAppComponents(ctx context.Context, appID string) ([]*models.AppComponent, error) {
	resp, err := c.genClient.Operations.GetV1AppsAppIDComponents(&operations.GetV1AppsAppIDComponentsParams{
		AppID:   appID,
		Context: ctx,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to get app components: %w", err)
	}

	return resp.Payload, nil
}

func (c *client) CreateComponent(ctx context.Context, appID string, req *models.ServiceCreateComponentRequest) (*models.AppComponent, error) {
	resp, err := c.genClient.Operations.PostV1AppsAppIDComponents(&operations.PostV1AppsAppIDComponentsParams{
		AppID:   appID,
		Req:     req,
		Context: ctx,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create app component: %w", err)
	}

	return resp.Payload, nil
}

func (c *client) GetComponent(ctx context.Context, componentID string) (*models.AppComponent, error) {
	resp, err := c.genClient.Operations.GetV1ComponentsComponentID(&operations.GetV1ComponentsComponentIDParams{
		ComponentID: componentID,
		Context:     ctx,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to get component: %w", err)
	}

	return resp.Payload, nil
}

func (c *client) UpdateComponent(ctx context.Context, componentID string, req *models.ServiceUpdateComponentRequest) (*models.AppComponent, error) {
	resp, err := c.genClient.Operations.PatchV1ComponentsComponentID(&operations.PatchV1ComponentsComponentIDParams{
		Req:         req,
		ComponentID: componentID,
		Context:     ctx,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to update component: %w", err)
	}

	return resp.Payload, nil
}

func (c *client) DeleteComponent(ctx context.Context, componentID string) (bool, error) {
	resp, err := c.genClient.Operations.DeleteV1ComponentsComponentID(&operations.DeleteV1ComponentsComponentIDParams{
		ComponentID: componentID,
		Context:     ctx,
	})
	if err != nil {
		return false, fmt.Errorf("unable to delete component: %w", err)
	}

	return resp.Payload, nil
}

// component configs
func (c *client) CreateTerraformModuleComponentConfig(ctx context.Context, componentID string, req *models.ServiceCreateTerraformModuleComponentConfigRequest) (*models.AppComponentConfigConnection, error) {
	resp, err := c.genClient.Operations.PostV1ComponentsComponentIDConfigsTerraformModule(&operations.PostV1ComponentsComponentIDConfigsTerraformModuleParams{
		ComponentID: componentID,
		Req:         req,
		Context:     ctx,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create terraform component config: %w", err)
	}

	return resp.Payload, nil
}

func (c *client) CreateHelmComponentConfig(ctx context.Context, componentID string, req *models.ServiceCreateHelmComponentConfigRequest) (*models.AppComponentConfigConnection, error) {
	resp, err := c.genClient.Operations.PostV1ComponentsComponentIDConfigsHelm(&operations.PostV1ComponentsComponentIDConfigsHelmParams{
		ComponentID: componentID,
		Req:         req,
		Context:     ctx,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create helm component config: %w", err)
	}

	return resp.Payload, nil
}

func (c *client) CreateDockerBuildComponentConfig(ctx context.Context, componentID string, req *models.ServiceCreateDockerBuildComponentConfigRequest) (*models.AppComponentConfigConnection, error) {
	resp, err := c.genClient.Operations.PostV1ComponentsComponentIDConfigsDockerBuild(&operations.PostV1ComponentsComponentIDConfigsDockerBuildParams{
		ComponentID: componentID,
		Req:         req,
		Context:     ctx,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create docker build component config: %w", err)
	}

	return resp.Payload, nil
}

func (c *client) CreateExternalImageComponentConfig(ctx context.Context, componentID string, req *models.ServiceCreateExternalImageComponentConfigRequest) (*models.AppComponentConfigConnection, error) {
	resp, err := c.genClient.Operations.PostV1ComponentsComponentIDConfigsExternalImage(&operations.PostV1ComponentsComponentIDConfigsExternalImageParams{
		ComponentID: componentID,
		Req:         req,
		Context:     ctx,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create external image component config: %w", err)
	}

	return resp.Payload, nil
}

func (c *client) GetComponentConfigs(ctx context.Context, componentID string) ([]*models.AppComponentConfigConnection, error) {
	resp, err := c.genClient.Operations.GetV1ComponentsComponentIDConfigs(&operations.GetV1ComponentsComponentIDConfigsParams{
		ComponentID: componentID,
		Context:     ctx,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to get component configs: %w", err)
	}

	return resp.Payload, nil
}

func (c *client) GetComponentLatestConfig(ctx context.Context, componentID string) (*models.AppComponentConfigConnection, error) {
	resp, err := c.genClient.Operations.GetV1ComponentsComponentIDConfig(&operations.GetV1ComponentsComponentIDConfigParams{
		ComponentID: componentID,
		Context:     ctx,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to get latest component config: %w", err)
	}

	return resp.Payload, nil
}

// builds
func (c *client) CreateComponentBuild(ctx context.Context, componentID string, req *models.ServiceCreateComponentBuildRequest) (*models.AppComponentBuild, error) {
	resp, err := c.genClient.Operations.PostV1ComponentsComponentIDBuilds(&operations.PostV1ComponentsComponentIDBuildsParams{
		ComponentID: componentID,
		Req:         req,
		Context:     ctx,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create component build: %w", err)
	}

	return resp.Payload, nil
}

func (c *client) GetComponentBuilds(ctx context.Context, componentID string) ([]*models.AppComponentBuild, error) {
	resp, err := c.genClient.Operations.GetV1ComponentsComponentIDBuilds(&operations.GetV1ComponentsComponentIDBuildsParams{
		ComponentID: componentID,
		Context:     ctx,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to get component builds: %w", err)
	}

	return resp.Payload, nil
}

func (c *client) GetComponentLatestBuild(ctx context.Context, componentID string) (*models.AppComponentBuild, error) {
	resp, err := c.genClient.Operations.GetV1ComponentsComponentIDLatestBuild(&operations.GetV1ComponentsComponentIDLatestBuildParams{
		ComponentID: componentID,
		Context:     ctx,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to get latest build for component: %w", err)
	}

	return resp.Payload, nil
}

func (c *client) GetComponentBuild(ctx context.Context, componentID string, buildID string) (*models.AppComponentBuild, error) {
	resp, err := c.genClient.Operations.GetV1ComponentsComponentIDBuildBuildID(&operations.GetV1ComponentsComponentIDBuildBuildIDParams{
		BuildID:     buildID,
		ComponentID: componentID,
		Context:     ctx,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to get latest build for component: %w", err)
	}

	return resp.Payload, nil
}

func (c *client) GetComponentBuildLogs(ctx context.Context, componentID string, buildID string) ([]models.ServiceBuildLog, error) {
	resp, err := c.genClient.Operations.GetV1ComponentsComponentIDBuildsBuildIDLogs(&operations.GetV1ComponentsComponentIDBuildsBuildIDLogsParams{
		ComponentID: componentID,
		BuildID:     buildID,
		Context:     ctx,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to get build logs: %w", err)
	}

	return resp.Payload, nil
}

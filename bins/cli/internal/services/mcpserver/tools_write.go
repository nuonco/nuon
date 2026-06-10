package mcpserver

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/nuonco/nuon/bins/cli/internal/lookup"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

type createInstallInput struct {
	App    string            `json:"app" jsonschema:"app name or ID to create the install for"`
	Name   string            `json:"name" jsonschema:"name for the new install"`
	Region string            `json:"region,omitempty" jsonschema:"cloud region to provision in"`
	Inputs map[string]string `json:"inputs,omitempty" jsonschema:"app input values"`
}

type deployComponentInput struct {
	Install   string `json:"install" jsonschema:"install name or ID to deploy to"`
	Component string `json:"component" jsonschema:"component name or ID to deploy"`
	BuildID   string `json:"build_id,omitempty" jsonschema:"build to deploy; omit to use the component's latest build"`
}

func (s *Service) registerWriteTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_install",
		Description: "Create a new install of an app. Mutates state; only available when the server runs with --allow-writes.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in createInstallInput) (*mcp.CallToolResult, any, error) {
		appID, err := lookup.AppID(ctx, s.api, in.App)
		if err != nil {
			return nil, nil, err
		}
		install, err := s.api.CreateInstall(ctx, appID, &models.ServiceCreateInstallRequest{
			Name: &in.Name,
			AwsAccount: &models.ServiceCreateInstallRequestAwsAccount{
				Region: in.Region,
			},
			Inputs: in.Inputs,
		})
		if err != nil {
			return nil, nil, err
		}
		return jsonResult(toInstallSummary(install))
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "deploy_component",
		Description: "Deploy a component build to an install. Mutates state; only available when the server runs with --allow-writes.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in deployComponentInput) (*mcp.CallToolResult, any, error) {
		installID, err := lookup.InstallID(ctx, s.api, in.Install)
		if err != nil {
			return nil, nil, err
		}

		buildID := in.BuildID
		if buildID == "" {
			install, err := s.api.GetInstall(ctx, installID)
			if err != nil {
				return nil, nil, err
			}
			componentID, err := lookup.ComponentID(ctx, s.api, install.AppID, in.Component)
			if err != nil {
				return nil, nil, err
			}
			build, err := s.api.GetComponentLatestBuild(ctx, componentID)
			if err != nil {
				return nil, nil, fmt.Errorf("unable to resolve latest build for component %s: %w", in.Component, err)
			}
			buildID = build.ID
		}

		deploy, err := s.api.CreateInstallDeploy(ctx, installID, &models.ServiceCreateInstallDeployRequest{
			BuildID: buildID,
		})
		if err != nil {
			return nil, nil, err
		}
		return jsonResult(map[string]string{
			"deploy_id": deploy.ID,
			"status":    deploy.Status,
			"build_id":  buildID,
		})
	})
}

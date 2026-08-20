package mcpserver

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/nuonco/nuon/bins/cli/internal/lookup"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

type createInstallInput struct {
	App                 string            `json:"app,omitempty" jsonschema:"app name or ID to create the install for; defaults to the currently selected app"`
	Name                string            `json:"name" jsonschema:"name for the new install"`
	Region              string            `json:"region,omitempty" jsonschema:"cloud region to provision in"`
	AWSAccountID        string            `json:"aws_account_id,omitempty" jsonschema:"AWS account ID this install targets; required when phone home authentication is enabled for the org"`
	AzureSubscriptionID string            `json:"azure_subscription_id,omitempty" jsonschema:"Azure subscription ID this install targets; required when phone home authentication is enabled for the org"`
	GCPProjectID        string            `json:"gcp_project_id,omitempty" jsonschema:"GCP project ID this install targets; required when phone home authentication is enabled for the org"`
	Inputs              map[string]string `json:"inputs,omitempty" jsonschema:"app input values"`
}

type deployComponentInput struct {
	Install   string `json:"install,omitempty" jsonschema:"install name or ID to deploy to; defaults to the currently selected install"`
	Component string `json:"component" jsonschema:"component name or ID to deploy"`
	BuildID   string `json:"build_id,omitempty" jsonschema:"build to deploy; omit to use the component's latest build"`
}

func (s *Service) registerWriteTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_install",
		Description: "Create a new install of an app. Mutates state; only available when the server runs with --allow-writes.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in createInstallInput) (*mcp.CallToolResult, any, error) {
		appID, err := s.resolveApp(ctx, in.App)
		if err != nil {
			return nil, nil, err
		}
		req := &models.ServiceCreateInstallRequest{Name: &in.Name, Inputs: in.Inputs}
		// Only one account block may be set, and sending the AWS one unconditionally made
		// Azure and GCP apps impossible to install through this tool.
		switch {
		case in.AzureSubscriptionID != "":
			req.AzureAccount = &models.HelpersCreateInstallAzureAccountParams{
				SubscriptionID: in.AzureSubscriptionID,
				Location:       in.Region,
			}
		case in.GCPProjectID != "":
			req.GcpAccount = &models.HelpersCreateInstallGCPAccountParams{
				ProjectID: in.GCPProjectID,
				Region:    in.Region,
			}
		default:
			req.AwsAccount = &models.HelpersCreateInstallAWSAccountParams{
				Region:    in.Region,
				AccountID: in.AWSAccountID,
			}
		}

		install, err := s.api.CreateInstall(ctx, appID, req)
		if err != nil {
			return nil, nil, err
		}
		return jsonResult(toInstallSummary(install))
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "deploy_component",
		Description: "Deploy a component build to an install. Mutates state; only available when the server runs with --allow-writes.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in deployComponentInput) (*mcp.CallToolResult, any, error) {
		installID, err := s.resolveInstall(ctx, in.Install)
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

package activities

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type CreateManagedAzureARMStackRequest struct {
	InstallID      string `json:"install_id" validate:"required"`
	StackVersionID string `json:"stack_version_id" validate:"required"`
	ConnectionID   string `json:"connection_id" validate:"required"`
}

// @temporal-gen-v2 activity
func (a *Activities) CreateManagedAzureARMStack(ctx context.Context, req *CreateManagedAzureARMStackRequest) error {
	var install app.Install
	if result := a.db.WithContext(ctx).Preload("AzureAccount").Preload("CloudConnection").First(&install, "id = ?", req.InstallID); result.Error != nil {
		return fmt.Errorf("load install: %w", result.Error)
	}
	if install.CloudConnectionID == nil || *install.CloudConnectionID != req.ConnectionID || install.CloudConnection == nil {
		return fmt.Errorf("install does not use cloud connection %s", req.ConnectionID)
	}
	if install.CloudConnection.Status != app.CloudConnectionStatusVerified || !install.CloudConnection.HasCapability(app.CloudConnectionCapabilityStacks) {
		return fmt.Errorf("cloud connection %s is not verified for install stacks", install.CloudConnection.ID)
	}
	if install.AzureAccount == nil || install.AzureAccount.Location == "" {
		return fmt.Errorf("install %s has no Azure location", install.ID)
	}

	var version app.InstallStackVersion
	if result := a.db.WithContext(ctx).First(&version, "id = ? AND install_id = ?", req.StackVersionID, install.ID); result.Error != nil {
		return fmt.Errorf("load install stack version: %w", result.Error)
	}
	if len(version.Contents) == 0 {
		return fmt.Errorf("install stack version %s has no template contents", version.ID)
	}
	var template map[string]any
	if err := json.Unmarshal(version.Contents, &template); err != nil {
		return fmt.Errorf("parse ARM template: %w", err)
	}

	credential, err := a.cloudConnectionsHelpers.AzureCredential(install.CloudConnection)
	if err != nil {
		return fmt.Errorf("resolve Azure cloud connection credentials: %w", err)
	}
	client, err := armresources.NewDeploymentsClient(install.CloudConnection.TargetID, credential, nil)
	if err != nil {
		return fmt.Errorf("create ARM deployments client: %w", err)
	}
	mode := armresources.DeploymentModeIncremental
	deployment := armresources.Deployment{Location: to.Ptr(install.AzureAccount.Location), Properties: &armresources.DeploymentProperties{Mode: &mode, Template: template}}
	schema, _ := template["$schema"].(string)
	if strings.Contains(schema, "subscriptionDeploymentTemplate") {
		if _, err := client.BeginCreateOrUpdateAtSubscriptionScope(ctx, version.ID, deployment, nil); err != nil {
			return fmt.Errorf("create subscription ARM deployment: %w", err)
		}
		return nil
	}

	resourceGroupName := install.ID + "-rg"
	groups, err := armresources.NewResourceGroupsClient(install.CloudConnection.TargetID, credential, nil)
	if err != nil {
		return fmt.Errorf("create Azure resource groups client: %w", err)
	}
	if _, err := groups.CreateOrUpdate(ctx, resourceGroupName, armresources.ResourceGroup{Location: to.Ptr(install.AzureAccount.Location)}, nil); err != nil {
		return fmt.Errorf("create resource group %s: %w", resourceGroupName, err)
	}
	if _, err := client.BeginCreateOrUpdate(ctx, resourceGroupName, version.ID, deployment, nil); err != nil {
		return fmt.Errorf("create resource-group ARM deployment: %w", err)
	}
	return nil
}

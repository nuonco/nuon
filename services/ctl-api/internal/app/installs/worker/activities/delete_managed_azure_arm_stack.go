package activities

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type DeleteManagedAzureARMStackRequest struct {
	InstallID    string `json:"install_id" validate:"required"`
	ConnectionID string `json:"connection_id" validate:"required"`
}

// @temporal-gen-v2 activity
func (a *Activities) DeleteManagedAzureARMStack(ctx context.Context, req *DeleteManagedAzureARMStackRequest) error {
	var install app.Install
	if result := a.db.WithContext(ctx).Preload("CloudConnection").First(&install, "id = ?", req.InstallID); result.Error != nil {
		return fmt.Errorf("load install: %w", result.Error)
	}
	if install.CloudConnectionID == nil || *install.CloudConnectionID != req.ConnectionID || install.CloudConnection == nil {
		return fmt.Errorf("install does not use cloud connection %s", req.ConnectionID)
	}
	if !install.CloudConnection.HasCapability(app.CloudConnectionCapabilityStacks) {
		return fmt.Errorf("cloud connection %s is not capable of managing install stacks", install.CloudConnection.ID)
	}
	credential, err := a.cloudConnectionsHelpers.AzureCredential(install.CloudConnection)
	if err != nil {
		return fmt.Errorf("resolve Azure cloud connection credentials: %w", err)
	}
	groups, err := armresources.NewResourceGroupsClient(install.CloudConnection.TargetID, credential, nil)
	if err != nil {
		return fmt.Errorf("create Azure resource groups client: %w", err)
	}
	if _, err := groups.BeginDelete(ctx, install.ID+"-rg", nil); err != nil {
		return fmt.Errorf("delete Azure resource group %s-rg: %w", install.ID, err)
	}
	return nil
}

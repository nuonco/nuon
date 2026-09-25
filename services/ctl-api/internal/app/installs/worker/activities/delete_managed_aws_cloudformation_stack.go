package activities

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"

	awscredentials "github.com/nuonco/nuon/pkg/aws/credentials"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type DeleteManagedAWSCloudFormationStackRequest struct {
	InstallID      string `json:"install_id" validate:"required"`
	StackVersionID string `json:"stack_version_id" validate:"required"`
	ConnectionID   string `json:"connection_id" validate:"required"`
}

// @temporal-gen-v2 activity
func (a *Activities) DeleteManagedAWSCloudFormationStack(ctx context.Context, req *DeleteManagedAWSCloudFormationStackRequest) error {
	var install app.Install
	if result := a.db.WithContext(ctx).Preload("AWSAccount").Preload("CloudConnection").First(&install, "id = ?", req.InstallID); result.Error != nil {
		return fmt.Errorf("load install: %w", result.Error)
	}
	if install.CloudConnectionID == nil || *install.CloudConnectionID != req.ConnectionID || install.CloudConnection == nil {
		return fmt.Errorf("install does not use cloud connection %s", req.ConnectionID)
	}
	if !install.CloudConnection.HasCapability(app.CloudConnectionCapabilityStacks) {
		return fmt.Errorf("cloud connection %s is not capable of managing install stacks", install.CloudConnection.ID)
	}
	if install.AWSAccount == nil {
		return fmt.Errorf("install %s has no AWS account", install.ID)
	}
	var version app.InstallStackVersion
	if result := a.db.WithContext(ctx).First(&version, "id = ? AND install_id = ?", req.StackVersionID, install.ID); result.Error != nil {
		return fmt.Errorf("load install stack version: %w", result.Error)
	}
	credentialsConfig, err := a.cloudConnectionsHelpers.Credentials(ctx, install.CloudConnection, "nuon-install-stack-delete")
	if err != nil {
		return fmt.Errorf("resolve cloud connection credentials: %w", err)
	}
	credentialsConfig.Region = install.AWSAccount.Region
	awsConfig, err := awscredentials.Fetch(ctx, credentialsConfig)
	if err != nil {
		return fmt.Errorf("assume cloud connection role: %w", err)
	}
	if _, err := cloudformation.NewFromConfig(awsConfig).DeleteStack(ctx, &cloudformation.DeleteStackInput{StackName: aws.String(version.StackName), ClientRequestToken: aws.String(version.ID)}); err != nil {
		return fmt.Errorf("delete cloudformation stack %q: %w", version.StackName, err)
	}
	return nil
}

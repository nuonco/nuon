package activities

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	cloudformationtypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/aws/smithy-go"

	awscredentials "github.com/nuonco/nuon/pkg/aws/credentials"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type DeleteManagedAWSCloudFormationStackRequest struct {
	InstallID      string `json:"install_id" validate:"required"`
	StackVersionID string `json:"stack_version_id" validate:"required"`
	ConnectionID   string `json:"connection_id" validate:"required"`
}

type cloudFormationDeleteStackAPI interface {
	DeleteStack(context.Context, *cloudformation.DeleteStackInput, ...func(*cloudformation.Options)) (*cloudformation.DeleteStackOutput, error)
	DescribeStackEvents(context.Context, *cloudformation.DescribeStackEventsInput, ...func(*cloudformation.Options)) (*cloudformation.DescribeStackEventsOutput, error)
}

func deleteManagedStack(ctx context.Context, client cloudFormationDeleteStackAPI, stackName, versionID string) error {
	token := versionID + "-delete"
	_, err := client.DeleteStack(ctx, &cloudformation.DeleteStackInput{StackName: aws.String(stackName), ClientRequestToken: aws.String(token)})
	if err == nil {
		return nil
	}
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) && apiErr.ErrorCode() == "ValidationError" && strings.Contains(apiErr.ErrorMessage(), "does not exist") {
		return nil
	}
	var tokenAlreadyExists *cloudformationtypes.TokenAlreadyExistsException
	if !errors.As(err, &tokenAlreadyExists) {
		return err
	}
	events, describeErr := client.DescribeStackEvents(ctx, &cloudformation.DescribeStackEventsInput{StackName: aws.String(stackName)})
	if describeErr != nil {
		return fmt.Errorf("reconcile delete stack after duplicate response: %v: %w", describeErr, err)
	}
	for _, event := range events.StackEvents {
		if aws.ToString(event.ClientRequestToken) == token && (event.ResourceStatus == cloudformationtypes.ResourceStatusDeleteInProgress || event.ResourceStatus == cloudformationtypes.ResourceStatusDeleteComplete) {
			return nil
		}
	}
	return err
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
	if install.CloudConnection.Status != app.CloudConnectionStatusVerified || install.CloudConnection.Platform != app.CloudPlatformAWS {
		return fmt.Errorf("cloud connection %s is not a verified AWS connection", install.CloudConnection.ID)
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
	if err := deleteManagedStack(ctx, cloudformation.NewFromConfig(awsConfig), version.StackName, version.ID); err != nil {
		return fmt.Errorf("delete cloudformation stack %q: %w", version.StackName, err)
	}
	return nil
}

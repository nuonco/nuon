package activities

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	cloudformationtypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"

	assumerole "github.com/nuonco/nuon/pkg/aws/assume-role"
	awscredentials "github.com/nuonco/nuon/pkg/aws/credentials"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type CreateManagedAWSCloudFormationStackRequest struct {
	InstallID      string `json:"install_id" validate:"required"`
	StackVersionID string `json:"stack_version_id" validate:"required"`
	ConnectionID   string `json:"connection_id" validate:"required"`
}

type cloudFormationCreateStackAPI interface {
	CreateStack(context.Context, *cloudformation.CreateStackInput, ...func(*cloudformation.Options)) (*cloudformation.CreateStackOutput, error)
	DescribeStackEvents(context.Context, *cloudformation.DescribeStackEventsInput, ...func(*cloudformation.Options)) (*cloudformation.DescribeStackEventsOutput, error)
}

func managedCreateStackInput(stackName, templateURL, requestToken string) *cloudformation.CreateStackInput {
	return &cloudformation.CreateStackInput{
		StackName:          aws.String(stackName),
		TemplateURL:        aws.String(templateURL),
		ClientRequestToken: aws.String(requestToken),
		Capabilities:       []cloudformationtypes.Capability{cloudformationtypes.CapabilityCapabilityNamedIam},
	}
}

func createManagedStack(ctx context.Context, client cloudFormationCreateStackAPI, input *cloudformation.CreateStackInput) error {
	if _, err := client.CreateStack(ctx, input); err != nil {
		var alreadyExists *cloudformationtypes.AlreadyExistsException
		var tokenAlreadyExists *cloudformationtypes.TokenAlreadyExistsException
		if !errors.As(err, &alreadyExists) && !errors.As(err, &tokenAlreadyExists) {
			return err
		}

		events, describeErr := client.DescribeStackEvents(ctx, &cloudformation.DescribeStackEventsInput{StackName: input.StackName})
		if describeErr != nil {
			return fmt.Errorf("reconcile create stack after duplicate response: %v: %w", describeErr, err)
		}
		for _, event := range events.StackEvents {
			if aws.ToString(event.ClientRequestToken) == aws.ToString(input.ClientRequestToken) {
				return nil
			}
		}
		return err
	}
	return nil
}

// @temporal-gen-v2 activity
func (a *Activities) CreateManagedAWSCloudFormationStack(ctx context.Context, req *CreateManagedAWSCloudFormationStackRequest) error {
	if a.cfg.ManagementIAMRoleARN == "" {
		return fmt.Errorf("management IAM role ARN is not configured")
	}

	var install app.Install
	if result := a.db.WithContext(ctx).Preload("AWSAccount").First(&install, "id = ?", req.InstallID); result.Error != nil {
		return fmt.Errorf("load install: %w", result.Error)
	}
	if install.AWSAccount == nil || install.AWSAccount.AWSAccountConnectionID == nil || *install.AWSAccount.AWSAccountConnectionID != req.ConnectionID {
		return fmt.Errorf("install does not use aws account connection %s", req.ConnectionID)
	}

	var connection app.AWSAccountConnection
	if result := a.db.WithContext(ctx).Where("id = ? AND org_id = ?", req.ConnectionID, install.OrgID).First(&connection); result.Error != nil {
		return fmt.Errorf("load aws account connection: %w", result.Error)
	}
	if connection.VerificationStatus != app.AWSAccountConnectionVerificationVerified || connection.RoleARN == "" {
		return fmt.Errorf("aws account connection %s is not verified and usable", connection.ID)
	}

	var version app.InstallStackVersion
	if result := a.db.WithContext(ctx).First(&version, "id = ? AND install_id = ?", req.StackVersionID, install.ID); result.Error != nil {
		return fmt.Errorf("load install stack version: %w", result.Error)
	}
	if version.TemplateURL == "" {
		return fmt.Errorf("install stack version %s has no template URL", version.ID)
	}

	if version.StackName == "" {
		return fmt.Errorf("install stack version %s has no stack name", version.ID)
	}

	awsConfig, err := awscredentials.Fetch(ctx, &awscredentials.Config{
		Region: install.AWSAccount.Region,
		AssumeRole: &awscredentials.AssumeRoleConfig{
			RoleARN:                connection.RoleARN,
			ExternalID:             connection.ExternalID,
			SessionName:            "nuon-install-stack",
			SessionDurationSeconds: int(time.Hour.Seconds()),
			TwoStepConfig:          &assumerole.TwoStepConfig{IAMRoleARN: a.cfg.ManagementIAMRoleARN},
		},
	})
	if err != nil {
		return fmt.Errorf("assume aws account connection role: %w", err)
	}

	if err := createManagedStack(ctx, cloudformation.NewFromConfig(awsConfig), managedCreateStackInput(version.StackName, version.TemplateURL, version.ID)); err != nil {
		return fmt.Errorf("create cloudformation stack %q (existing stacks are not updated): %w", version.StackName, err)
	}
	return nil
}

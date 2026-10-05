package activities

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	cloudformationtypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
)

type CreateManagedAWSCloudFormationStackRequest struct {
	InstallID      string `json:"install_id" validate:"required"`
	StackVersionID string `json:"stack_version_id" validate:"required"`
	ConnectionID   string `json:"connection_id" validate:"required"`
}

type cloudFormationCreateStackAPI interface {
	CreateStack(context.Context, *cloudformation.CreateStackInput, ...func(*cloudformation.Options)) (*cloudformation.CreateStackOutput, error)
	UpdateStack(context.Context, *cloudformation.UpdateStackInput, ...func(*cloudformation.Options)) (*cloudformation.UpdateStackOutput, error)
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
		if errors.As(err, &alreadyExists) {
			_, updateErr := client.UpdateStack(ctx, &cloudformation.UpdateStackInput{
				StackName:          input.StackName,
				TemplateURL:        input.TemplateURL,
				ClientRequestToken: input.ClientRequestToken,
				Capabilities:       input.Capabilities,
			})
			if updateErr != nil && !strings.Contains(updateErr.Error(), "No updates are to be performed") {
				return updateErr
			}
			return nil
		}
		if !errors.As(err, &tokenAlreadyExists) {
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
	version, awsConfig, err := a.managedStackSession(ctx, req.InstallID, req.StackVersionID, req.ConnectionID, "nuon-install-stack")
	if err != nil {
		return err
	}
	if version.TemplateURL == "" {
		return fmt.Errorf("install stack version %s has no template URL", version.ID)
	}
	if version.StackName == "" {
		return fmt.Errorf("install stack version %s has no stack name", version.ID)
	}
	if err := createManagedStack(ctx, cloudformation.NewFromConfig(awsConfig), managedCreateStackInput(version.StackName, version.TemplateURL, version.ID)); err != nil {
		return fmt.Errorf("create or update cloudformation stack %q: %w", version.StackName, err)
	}
	return nil
}

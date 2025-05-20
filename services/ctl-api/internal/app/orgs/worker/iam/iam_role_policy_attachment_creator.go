package orgiam

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/go-playground/validator/v10"
	assumerole "github.com/powertoolsdev/mono/pkg/aws/assume-role"
)

type CreateIAMRolePolicyAttachmentRequest struct {
	AssumeRoleARN string `validate:"required" json:"assume_role_arn"`

	PolicyArn string `validate:"required" json:"policy_arn"`
	RoleName  string `validate:"required" json:"role_name"`
}

func (r CreateIAMRolePolicyAttachmentRequest) validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

type CreateIAMRolePolicyAttachmentResponse struct{}

// @temporal-gen activity
// @schedule-to-close-timeout 1m
func (a *Activities) CreateIAMRolePolicyAttachment(ctx context.Context, req CreateIAMRolePolicyAttachmentRequest) (CreateIAMRolePolicyAttachmentResponse, error) {
	var resp CreateIAMRolePolicyAttachmentResponse
	if err := req.validate(); err != nil {
		return resp, fmt.Errorf("unable to validate request: %w", err)
	}

	assumer, err := assumerole.New(a.validator, assumerole.WithRoleARN(req.AssumeRoleARN), assumerole.WithRoleSessionName("workers-orgs-iam-policy-creator"))
	if err != nil {
		return resp, fmt.Errorf("unable to create role assumer: %w", err)
	}
	cfg, err := assumer.LoadConfigWithAssumedRole(ctx)
	if err != nil {
		return resp, fmt.Errorf("unable to load config with assumed role: %w", err)
	}

	client := iam.NewFromConfig(cfg)
	if err := a.iamRolePolicyAttachmentCreator.createIAMRolePolicyAttachment(ctx, client, req.PolicyArn, req.RoleName); err != nil {
		return resp, fmt.Errorf("unable to create IAM role policy attachment: %w", err)
	}

	return resp, nil
}

type iamRolePolicyAttachmentCreator interface {
	createIAMRolePolicyAttachment(context.Context, awsClientIAMRolePolicyAttacher, string, string) error
}

var _ iamRolePolicyAttachmentCreator = (*iamRolePolicyAttachmentCreatorImpl)(nil)

type iamRolePolicyAttachmentCreatorImpl struct{}

type awsClientIAMRolePolicyAttacher interface {
	AttachRolePolicy(context.Context, *iam.AttachRolePolicyInput, ...func(*iam.Options)) (*iam.AttachRolePolicyOutput, error)
}

func (o *iamRolePolicyAttachmentCreatorImpl) createIAMRolePolicyAttachment(ctx context.Context, client awsClientIAMRolePolicyAttacher, policyArn, roleName string) error {
	params := &iam.AttachRolePolicyInput{
		PolicyArn: &policyArn,
		RoleName:  &roleName,
	}
	_, err := client.AttachRolePolicy(ctx, params)
	if err != nil {
		return fmt.Errorf("unable to create role policy attachment: %w", err)
	}

	return nil
}

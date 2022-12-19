package iam

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/go-playground/validator/v10"
)

type CreateIAMRolePolicyAttachmentRequest struct {
	AssumeRoleARN string `validate:"required" json:"assume_role_arn"`

	PolicyArn string `validate:"required" json:"policy_arn"`
	RoleArn   string `validate:"required" json:"role_arn"`
}

func (r CreateIAMRolePolicyAttachmentRequest) validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

type CreateIAMRolePolicyAttachmentResponse struct{}

func (a *Activities) CreateIAMRolePolicyAttachment(ctx context.Context, req CreateIAMRolePolicyAttachmentRequest) (CreateIAMRolePolicyAttachmentResponse, error) {
	var resp CreateIAMRolePolicyAttachmentResponse
	if err := req.validate(); err != nil {
		return resp, fmt.Errorf("unable to validate request: %w", err)
	}

	cfg, err := a.loadConfigWithAssumeRole(ctx, req.AssumeRoleARN)
	if err != nil {
		return resp, fmt.Errorf("unable to assume role: %w", err)
	}

	client := iam.NewFromConfig(cfg)
	if err := a.createIAMRolePolicyAttachment(ctx, client, req.PolicyArn, req.RoleArn); err != nil {
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

func (o *iamRolePolicyAttachmentCreatorImpl) createIAMRolePolicyAttachment(ctx context.Context, client awsClientIAMRolePolicyAttacher, policyArn, roleArn string) error {
	params := &iam.AttachRolePolicyInput{
		PolicyArn: &policyArn,
		RoleName:  &roleArn,
	}
	_, err := client.AttachRolePolicy(ctx, params)
	if err != nil {
		return fmt.Errorf("unable to create role policy attachment: %w", err)
	}

	return nil
}

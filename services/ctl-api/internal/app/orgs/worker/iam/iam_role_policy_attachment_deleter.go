package orgiam

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/go-playground/validator/v10"
	assumerole "github.com/powertoolsdev/mono/pkg/aws/assume-role"
)

type DeleteIAMRolePolicyAttachmentRequest struct {
	AssumeRoleARN string `validate:"required" json:"assume_role_arn"`

	PolicyArn string `validate:"required" json:"policy_arn"`
	RoleName  string `validate:"required" json:"role_name"`
}

func (r DeleteIAMRolePolicyAttachmentRequest) validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

type DeleteIAMRolePolicyAttachmentResponse struct{}

// @temporal-gen activity
// @schedule-to-close-timeout 1m
// @max-retries 2
func (a *Activities) DeleteIAMRolePolicyAttachment(ctx context.Context, req DeleteIAMRolePolicyAttachmentRequest) (DeleteIAMRolePolicyAttachmentResponse, error) {
	var resp DeleteIAMRolePolicyAttachmentResponse
	if err := req.validate(); err != nil {
		return resp, fmt.Errorf("unable to validate request: %w", err)
	}

	assumer, err := assumerole.New(a.validator,
		assumerole.WithRoleARN(req.AssumeRoleARN),
		assumerole.WithRoleSessionName("workers-orgs-iam-policy-deleter"))
	if err != nil {
		return resp, fmt.Errorf("unable to delete role assumer: %w", err)
	}
	cfg, err := assumer.LoadConfigWithAssumedRole(ctx)
	if err != nil {
		return resp, fmt.Errorf("unable to load config with assumed role: %w", err)
	}

	client := iam.NewFromConfig(cfg)
	if err := a.iamRolePolicyAttachmentDeleter.deleteIAMRolePolicyAttachment(ctx, client, req.PolicyArn, req.RoleName); err != nil {
		return resp, fmt.Errorf("unable to delete IAM role policy attachment: %w", err)
	}

	return resp, nil
}

type iamRolePolicyAttachmentDeleter interface {
	deleteIAMRolePolicyAttachment(context.Context, awsClientIAMRolePolicyDetacher, string, string) error
}

var _ iamRolePolicyAttachmentDeleter = (*iamRolePolicyAttachmentDeleterImpl)(nil)

type iamRolePolicyAttachmentDeleterImpl struct{}

type awsClientIAMRolePolicyDetacher interface {
	DetachRolePolicy(context.Context, *iam.DetachRolePolicyInput, ...func(*iam.Options)) (*iam.DetachRolePolicyOutput, error)
}

func (o *iamRolePolicyAttachmentDeleterImpl) deleteIAMRolePolicyAttachment(ctx context.Context, client awsClientIAMRolePolicyDetacher, policyArn, roleName string) error {
	params := &iam.DetachRolePolicyInput{
		PolicyArn: &policyArn,
		RoleName:  &roleName,
	}
	_, err := client.DetachRolePolicy(ctx, params)
	if err != nil {
		if isNotFoundErr(err) {
			return nil
		}

		return fmt.Errorf("unable to create role policy attachment: %w", err)
	}

	return nil
}

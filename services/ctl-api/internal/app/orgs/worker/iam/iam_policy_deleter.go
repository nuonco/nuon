package orgiam

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/go-playground/validator/v10"
	assumerole "github.com/powertoolsdev/mono/pkg/aws/assume-role"
)

type DeleteIAMPolicyRequest struct {
	AssumeRoleARN string `validate:"required" json:"assume_role_arn"`

	PolicyARN string `validate:"required" json:"policy_arn"`
}

func (r DeleteIAMPolicyRequest) validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

type DeleteIAMPolicyResponse struct{}

// @temporal-gen activity
// @schedule-to-close-timeout 1m
// @max-retries 2
func (a *Activities) DeleteIAMPolicy(ctx context.Context, req DeleteIAMPolicyRequest) (DeleteIAMPolicyResponse, error) {
	var resp DeleteIAMPolicyResponse
	if err := req.validate(); err != nil {
		return resp, fmt.Errorf("unable to validate request: %w", err)
	}

	assumer, err := assumerole.New(a.validator,
		assumerole.WithRoleARN(req.AssumeRoleARN),
		assumerole.WithRoleSessionName("workers-orgs-iam-policy-creator"))
	if err != nil {
		return resp, fmt.Errorf("unable to delete role assumer: %w", err)
	}
	cfg, err := assumer.LoadConfigWithAssumedRole(ctx)
	if err != nil {
		return resp, fmt.Errorf("unable to load config with assumed role: %w", err)
	}

	client := iam.NewFromConfig(cfg)
	err = a.iamPolicyDeleter.deleteIAMPolicy(ctx, client, req)
	if err != nil {
		return resp, fmt.Errorf("unable to delete IAM policy: %w", err)
	}

	return resp, nil
}

type iamPolicyDeleter interface {
	deleteIAMPolicy(context.Context, awsClientIAMPolicyDeleter, DeleteIAMPolicyRequest) error
}

var _ iamPolicyDeleter = (*iamPolicyDeleterImpl)(nil)

type iamPolicyDeleterImpl struct{}

type awsClientIAMPolicyDeleter interface {
	DeletePolicy(context.Context, *iam.DeletePolicyInput, ...func(*iam.Options)) (*iam.DeletePolicyOutput, error)
}

func (o *iamPolicyDeleterImpl) deleteIAMPolicy(ctx context.Context, client awsClientIAMPolicyDeleter, req DeleteIAMPolicyRequest) error {
	params := &iam.DeletePolicyInput{
		PolicyArn: &req.PolicyARN,
	}

	_, err := client.DeletePolicy(ctx, params)
	if err != nil {
		if isNotFoundErr(err) {
			return nil
		}

		return fmt.Errorf("unable to delete policy: %w", err)
	}

	return nil
}

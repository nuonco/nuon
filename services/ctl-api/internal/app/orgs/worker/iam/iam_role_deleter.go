package orgiam

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/go-playground/validator/v10"
	assumerole "github.com/powertoolsdev/mono/pkg/aws/assume-role"
	"github.com/powertoolsdev/mono/pkg/generics"
)

type DeleteIAMRoleRequest struct {
	AssumeRoleARN string `validate:"required" json:"assume_role_arn"`

	RoleName string `validate:"required" json:"role_name"`
}

type DeleteIAMRoleResponse struct {
	RoleArn string `json:"role_arn" validate:"required"`
}

// @temporal-gen activity
// @schedule-to-close-timeout 1m
// @max-retries 2
func (a *Activities) DeleteIAMRole(ctx context.Context, req DeleteIAMRoleRequest) (DeleteIAMRoleResponse, error) {
	var resp DeleteIAMRoleResponse
	if err := req.validate(); err != nil {
		return resp, fmt.Errorf("unable to validate request: %w", err)
	}

	assumer, err := assumerole.New(a.validator,
		assumerole.WithRoleARN(req.AssumeRoleARN),
		assumerole.WithRoleSessionName("workers-orgs-iam-role-deleter"))
	if err != nil {
		return resp, fmt.Errorf("unable to create role assumer: %w", err)
	}
	cfg, err := assumer.LoadConfigWithAssumedRole(ctx)
	if err != nil {
		return resp, fmt.Errorf("unable to load config with assumed role: %w", err)
	}

	client := iam.NewFromConfig(cfg)
	err = a.iamRoleDeleter.deleteIAMRole(ctx, client, req)
	if err != nil {
		return resp, fmt.Errorf("unable to delete IAM role: %w", err)
	}
	return resp, nil
}

func (r DeleteIAMRoleRequest) validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

type iamRoleDeleter interface {
	deleteIAMRole(context.Context, awsClientIAMRoleDeleter, DeleteIAMRoleRequest) error
}

var _ iamRoleDeleter = (*iamRoleDeleterImpl)(nil)

type iamRoleDeleterImpl struct{}

type awsClientIAMRoleDeleter interface {
	DeleteRole(context.Context, *iam.DeleteRoleInput, ...func(*iam.Options)) (*iam.DeleteRoleOutput, error)
}

func (o *iamRoleDeleterImpl) deleteIAMRole(ctx context.Context, client awsClientIAMRoleDeleter, req DeleteIAMRoleRequest) error {
	params := &iam.DeleteRoleInput{
		RoleName: generics.ToPtr(req.RoleName),
	}

	_, err := client.DeleteRole(ctx, params)
	if err != nil {
		if isNotFoundErr(err) {
			return nil
		}

		return fmt.Errorf("unable to create IAM role: %w", err)
	}

	return nil
}

package iam

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	iam_types "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/go-playground/validator/v10"
	assumerole "github.com/powertoolsdev/mono/pkg/aws/assume-role"
	"github.com/powertoolsdev/mono/pkg/generics"
)

const (
	defaultIAMRoleSessionDuration time.Duration = time.Minute * 60
)

type CreateIAMRoleRequest struct {
	AssumeRoleARN string `validate:"required" json:"assume_role_arn"`

	RoleARN             string      `validate:"required" json:"role_arn"`
	RoleName            string      `validate:"required" json:"role_name"`
	RolePath            string      `validate:"required" json:"role_path"`
	TrustPolicyDocument string      `validate:"required" json:"trust_policy_document"`
	RoleTags            [][2]string `validate:"required" json:"role_tags"`
}

type CreateIAMRoleResponse struct {
	RoleArn string `json:"role_arn" validate:"required"`
}

func (a *Activities) CreateIAMRole(ctx context.Context, req CreateIAMRoleRequest) (CreateIAMRoleResponse, error) {
	var resp CreateIAMRoleResponse
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
	roleArn, err := a.iamRoleCreator.createIAMRole(ctx, client, req)
	if err == nil {
		resp.RoleArn = roleArn
		return resp, nil
	}
	if !isEntityExistsException(err) {
		return resp, fmt.Errorf("unable to create odr IAM role: %w", err)
	}

	resp.RoleArn = req.RoleARN
	return resp, nil
}

func (r CreateIAMRoleRequest) validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

type iamRoleCreator interface {
	createIAMRole(context.Context, awsClientIAMRoleCreator, CreateIAMRoleRequest) (string, error)
}

var _ iamRoleCreator = (*iamRoleCreatorImpl)(nil)

type iamRoleCreatorImpl struct{}

type awsClientIAMRoleCreator interface {
	CreateRole(context.Context, *iam.CreateRoleInput, ...func(*iam.Options)) (*iam.CreateRoleOutput, error)
}

func (o *iamRoleCreatorImpl) createIAMRole(ctx context.Context, client awsClientIAMRoleCreator, req CreateIAMRoleRequest) (string, error) {
	tags := make([]iam_types.Tag, 0, len(req.RoleTags)+1)
	for _, pair := range req.RoleTags {
		tags = append(tags, iam_types.Tag{
			Key:   generics.ToPtr(pair[0]),
			Value: generics.ToPtr(pair[1]),
		})
	}

	params := &iam.CreateRoleInput{
		AssumeRolePolicyDocument: &req.TrustPolicyDocument,
		RoleName:                 &req.RoleName,
		MaxSessionDuration:       generics.ToPtr(int32(defaultIAMRoleSessionDuration.Seconds())),
		Path:                     &req.RolePath,
		Tags:                     tags,
	}

	resp, err := client.CreateRole(ctx, params)
	if err != nil {
		return "", fmt.Errorf("unable to create IAM role: %w", err)
	}

	return *resp.Role.Arn, nil
}

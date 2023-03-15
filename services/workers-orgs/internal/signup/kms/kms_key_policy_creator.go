package kms

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	iam_types "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/go-playground/validator/v10"
	assumerole "github.com/powertoolsdev/mono/pkg/aws-assume-role"
	"github.com/powertoolsdev/mono/pkg/generics"
)

const (
	defaultIAMRoleSessionDuration time.Duration = time.Minute * 60
)

type CreateKMSKeyPolicyRequest struct {
	AssumeRoleARN string `validate:"required" json:"assume_role_arn"`

	KeyArn              string      `validate:"required" json:"key_arn"`
	RoleName            string      `validate:"required" json:"role_name"`
	RolePath            string      `validate:"required" json:"role_path"`
	TrustPolicyDocument string      `validate:"required" json:"trust_policy_document"`
	RoleTags            [][2]string `validate:"required" json:"role_tags"`
}

type CreateKMSKeyPolicyResponse struct {
	RoleArn string `json:"role_arn" validate:"required"`
}

func (a *Activities) CreateKMSKeyPolicy(ctx context.Context, req CreateKMSKeyPolicyRequest) (CreateKMSKeyPolicyResponse, error) {
	var resp CreateKMSKeyPolicyResponse
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
	roleArn, err := a.kmsKeyPolicyCreator.createIAMRole(ctx, client, req)
	if err != nil {
		return resp, fmt.Errorf("unable to create odr IAM role: %w", err)
	}
	resp.RoleArn = roleArn

	return resp, nil
}

func (r CreateKMSKeyPolicyRequest) validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

type kmsKeyPolicyCreator interface {
	createIAMRole(context.Context, awsClientIAMRoleCreator, CreateKMSKeyPolicyRequest) (string, error)
}

var _ kmsKeyPolicyCreator = (*kmsKeyPolicyCreatorImpl)(nil)

type kmsKeyPolicyCreatorImpl struct{}

type awsClientIAMRoleCreator interface {
	CreateRole(context.Context, *iam.CreateRoleInput, ...func(*iam.Options)) (*iam.CreateRoleOutput, error)
}

func (o *kmsKeyPolicyCreatorImpl) createIAMRole(ctx context.Context, client awsClientIAMRoleCreator, req CreateKMSKeyPolicyRequest) (string, error) {
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

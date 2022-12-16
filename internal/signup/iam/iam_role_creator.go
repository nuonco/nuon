package iam

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	iam_types "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/go-playground/validator/v10"
)

const (
	defaultIAMPolicyVersion       string        = "2012-10-17"
	defaultIAMRoleSessionDuration time.Duration = time.Minute * 60
)

type CreateIAMRoleRequest struct {
	OrgsIAMAccessRoleArn string `validate:"required" json:"orgs_iam_access_role_arn"`

	RoleName            string      `validate:"required" json:"role_name"`
	RolePath            string      `validate:"required" json:"role_path"`
	TrustPolicyDocument string      `validate:"required" json:"trust_policy_document"`
	RoleTags            [][2]string `validate:"role_tags" json:"role_tags"`
}

type CreateIAMRoleResponse struct {
	IAMRoleArn string `json:"iam_role_arn" validate:"required"`
}

func (a *Activities) CreateIAMRole(ctx context.Context, req CreateIAMRoleRequest) (CreateIAMRoleResponse, error) {
	var resp CreateIAMRoleResponse
	if err := req.validate(); err != nil {
		return resp, fmt.Errorf("unable to validate request: %w", err)
	}

	cfg, err := a.loadConfigWithAssumeRole(ctx, req.OrgsIAMAccessRoleArn)
	if err != nil {
		return resp, fmt.Errorf("unable to assume role: %w", err)
	}

	client := iam.NewFromConfig(cfg)
	roleArn, err := a.createIAMRole(ctx, client, req)
	if err != nil {
		return resp, fmt.Errorf("unable to create odr IAM role: %w", err)
	}
	resp.IAMRoleArn = roleArn

	return resp, nil
}

func (r CreateIAMRoleRequest) validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

type iamRoleCreator interface {
	createIAMRole(context.Context, awsClientIAMRole, CreateIAMRoleRequest) (string, error)
}

var _ iamRoleCreator = (*iamRoleCreatorImpl)(nil)

type iamRoleCreatorImpl struct{}

type awsClientIAMRole interface {
	CreateRole(context.Context, *iam.CreateRoleInput, ...func(*iam.Options)) (*iam.CreateRoleOutput, error)
}

func (o *iamRoleCreatorImpl) createIAMRole(ctx context.Context, client awsClientIAMRole, req CreateIAMRoleRequest) (string, error) {
	tags := make([]iam_types.Tag, 0, len(req.RoleTags)+1)
	for _, pair := range req.RoleTags {
		tags = append(tags, iam_types.Tag{
			Key:   toPtr(pair[0]),
			Value: toPtr(pair[1]),
		})
	}

	params := &iam.CreateRoleInput{
		AssumeRolePolicyDocument: toPtr(req.TrustPolicyDocument),
		RoleName:                 toPtr(req.RoleName),
		MaxSessionDuration:       toPtr(int32(defaultIAMRoleSessionDuration.Seconds())),
		Path:                     toPtr(req.RolePath),
		Tags:                     tags,
	}

	resp, err := client.CreateRole(ctx, params)
	if err != nil {
		return "", fmt.Errorf("unable to create IAM role: %w", err)
	}

	return *resp.Role.Arn, nil
}

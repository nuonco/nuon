package iam

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	iam_types "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/go-playground/validator/v10"
	assumerole "github.com/powertoolsdev/mono/pkg/aws/assume-role"
	"github.com/powertoolsdev/mono/pkg/generics"
)

type CreateIAMPolicyRequest struct {
	AssumeRoleARN string `validate:"required" json:"assume_role_arn"`

	PolicyName     string `validate:"required" json:"policy_name"`
	PolicyARN      string `validate:"required" json:"policy_arn"`
	PolicyPath     string `validate:"required" json:"policy_path"`
	PolicyDocument string `validate:"required" json:"policy_document"`

	PolicyTags [][2]string `validate:"required" json:"policy_tags"`
}

func (r CreateIAMPolicyRequest) validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

type CreateIAMPolicyResponse struct {
	PolicyArn string `validate:"required" json:"policy_arn"`
}

func (a *Activities) CreateIAMPolicy(ctx context.Context, req CreateIAMPolicyRequest) (CreateIAMPolicyResponse, error) {
	var resp CreateIAMPolicyResponse
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
	policyArn, err := a.iamPolicyCreator.createIAMPolicy(ctx, client, req)
	if err == nil {
		resp.PolicyArn = policyArn
		return resp, nil
	}
	if !isEntityExistsException(err) {
		return resp, fmt.Errorf("unable to create odr IAM policy: %w", err)
	}

	// default case where the policy exists already
	resp.PolicyArn = req.PolicyARN
	return resp, nil
}

type iamPolicyCreator interface {
	createIAMPolicy(context.Context, awsClientIAMPolicy, CreateIAMPolicyRequest) (string, error)
}

var _ iamPolicyCreator = (*iamPolicyCreatorImpl)(nil)

type iamPolicyCreatorImpl struct{}

type awsClientIAMPolicy interface {
	CreatePolicy(context.Context, *iam.CreatePolicyInput, ...func(*iam.Options)) (*iam.CreatePolicyOutput, error)
}

func (o *iamPolicyCreatorImpl) createIAMPolicy(ctx context.Context, client awsClientIAMPolicy, req CreateIAMPolicyRequest) (string, error) {
	tags := make([]iam_types.Tag, 0, len(req.PolicyTags)+1)
	for _, pair := range req.PolicyTags {
		tags = append(tags, iam_types.Tag{
			Key:   generics.ToPtr(pair[0]),
			Value: generics.ToPtr(pair[1]),
		})
	}

	params := &iam.CreatePolicyInput{
		PolicyDocument: &req.PolicyDocument,
		PolicyName:     &req.PolicyName,
		Path:           &req.PolicyPath,
		Tags:           tags,
	}

	output, err := client.CreatePolicy(ctx, params)
	if err != nil {
		return "", fmt.Errorf("unable to create policy: %w", err)
	}

	return *output.Policy.Arn, nil
}

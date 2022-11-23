package runner

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/go-playground/validator/v10"
)

func odrIAMPolicyName(orgID string) string {
	return fmt.Sprintf("org-odr-policy-%s", orgID)
}

type CreateOdrIAMPolicyRequest struct {
	OrgID string `validate:"required" json:"org_id"`

	OrgsIAMAccessRoleArn string `validate:"required" json:"orgs_iam_access_role_arn"`

	// the following fields are used to configure the IAM role
	ECRRegistryArn string `validate:"required" json:"ecr_registry_arn"`
}

func (r CreateOdrIAMPolicyRequest) validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

type CreateOdrIAMPolicyResponse struct {
	PolicyArn string `validate:"required" json:"policy_arn"`
}

func (a *Activities) CreateOdrIAMPolicy(ctx context.Context, req CreateOdrIAMPolicyRequest) (CreateOdrIAMPolicyResponse, error) {
	var resp CreateOdrIAMPolicyResponse
	if err := req.validate(); err != nil {
		return resp, fmt.Errorf("unable to validate request: %w", err)
	}

	cfg, err := a.loadConfigWithAssumeRole(ctx, req.OrgsIAMAccessRoleArn)
	if err != nil {
		return resp, fmt.Errorf("unable to assume role: %w", err)
	}

	client := iam.NewFromConfig(cfg)
	policyArn, err := a.createOdrIAMPolicy(ctx, client, req)
	if err != nil {
		return resp, fmt.Errorf("unable to create odr IAM policy: %w", err)
	}

	resp.PolicyArn = policyArn
	return resp, nil
}

type odrIAMPolicyCreator interface {
	createOdrIAMPolicy(context.Context, awsClientIAMPolicy, CreateOdrIAMPolicyRequest) (string, error)
}

var _ odrIAMPolicyCreator = (*odrIAMPolicyCreatorImpl)(nil)

type odrIAMPolicyCreatorImpl struct{}

type odrIAMRolePolicy struct {
	Version   string             `json:"Version"`
	Statement []IAMRoleStatement `json:"Statement"`
}

type awsClientIAMPolicy interface {
	CreatePolicy(context.Context, *iam.CreatePolicyInput, ...func(*iam.Options)) (*iam.CreatePolicyOutput, error)
}

func (o *odrIAMPolicyCreatorImpl) createOdrIAMPolicy(ctx context.Context, client awsClientIAMPolicy, req CreateOdrIAMPolicyRequest) (string, error) {
	policy := odrIAMRolePolicy{
		Version: defaultIAMPolicyVersion,
		Statement: []IAMRoleStatement{
			{
				Effect: "Allow",
				Action: []string{
					"ecr:*",
				},
				Resource: fmt.Sprintf("%s/%s/*", req.ECRRegistryArn, req.OrgID),
			},
		},
	}
	byts, err := json.Marshal(policy)
	if err != nil {
		return "", fmt.Errorf("unable to convert policy to json: %w", err)
	}

	params := &iam.CreatePolicyInput{
		PolicyDocument: toPtr(string(byts)),
		PolicyName:     toPtr(odrIAMPolicyName(req.OrgID)),
		Path:           toPtr("/orgs/"),
	}

	output, err := client.CreatePolicy(ctx, params)
	if err != nil {
		return "", fmt.Errorf("unable to create policy: %w", err)
	}
	return *output.Policy.Arn, nil
}

package runner

import (
	"context"
	"encoding/json"
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

func odrIAMRoleName(orgID string) string {
	return fmt.Sprintf("org-odr-role-%s", orgID)
}

type CreateOdrIAMRoleRequest struct {
	OrgID string `validate:"required" json:"org_id"`

	OrgsIAMAccessRoleArn string `validate:"required" json:"orgs_iam_access_role_arn"`

	// the following fields are used to configure the IAM role
	OrgsIAMOidcProviderArn string `validate:"required" json:"orgs_iam_oidc_provider_arn"`
	OrgsIAMOidcProviderURL string `validate:"required" json:"orgs_iam_oidc_provider_url"`
	ECRRegistryArn         string `validate:"required" json:"ecr_registry_arn"`
	PolicyArn              string `validate:"required" json:"policy_arn"`
}

type CreateOdrIAMRoleResponse struct {
	IAMRoleArn string `json:"iam_role_arn" validate:"required"`
}

func (a *Activities) CreateOdrIAMRole(ctx context.Context, req CreateOdrIAMRoleRequest) (CreateOdrIAMRoleResponse, error) {
	var resp CreateOdrIAMRoleResponse
	if err := req.validate(); err != nil {
		return resp, fmt.Errorf("unable to validate request: %w", err)
	}

	cfg, err := a.loadConfigWithAssumeRole(ctx, req.OrgsIAMAccessRoleArn)
	if err != nil {
		return resp, fmt.Errorf("unable to assume role: %w", err)
	}

	client := iam.NewFromConfig(cfg)
	if err := a.createOdrIAMRole(ctx, client, req); err != nil {
		return resp, fmt.Errorf("unable to create odr IAM role: %w", err)
	}

	if err := a.createOdrIAMRolePolicyAttachment(ctx, client, req.PolicyArn, req); err != nil {
		return resp, fmt.Errorf("unable to create odr IAM role attachment: %w", err)
	}

	return resp, nil
}

func (r CreateOdrIAMRoleRequest) validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

type odrIAMRoleCreator interface {
	createOdrIAMRole(context.Context, awsClientIAMRole, CreateOdrIAMRoleRequest) error
	createOdrIAMRolePolicyAttachment(context.Context, awsClientIAMRole, string, CreateOdrIAMRoleRequest) error
}

var _ odrIAMRoleCreator = (*odrIAMRoleCreatorImpl)(nil)

type odrIAMRoleCreatorImpl struct{}

type awsClientIAMRole interface {
	CreateRole(context.Context, *iam.CreateRoleInput, ...func(*iam.Options)) (*iam.CreateRoleOutput, error)
	AttachRolePolicy(context.Context, *iam.AttachRolePolicyInput, ...func(*iam.Options)) (*iam.AttachRolePolicyOutput, error)
}

func (o *odrIAMRoleCreatorImpl) createOdrIAMRole(ctx context.Context, client awsClientIAMRole, req CreateOdrIAMRoleRequest) error {
	trustPolicy := odrIAMRoleTrustPolicy{
		Version: defaultIAMPolicyVersion,
		Statement: []IAMRoleTrustStatement{
			{
				Action: "sts:AssumeRoleWithWebIdentity",
				Effect: "Allow",
				Sid:    "",
				Principal: struct {
					Federated string `json:"Federated,omitempty"`
				}{
					Federated: req.OrgsIAMOidcProviderArn,
				},
				Condition: struct {
					StringEquals map[string]string `json:"StringEquals"`
				}{
					StringEquals: map[string]string{
						fmt.Sprintf("%s:sub", req.OrgsIAMOidcProviderURL): fmt.Sprintf("system:serviceaccount:%s:%s", req.OrgID, runnerOdrServiceAccountName(req.OrgID)),
					},
				},
			},
		},
	}

	trustPolicyDoc, err := json.Marshal(trustPolicy)
	if err != nil {
		return fmt.Errorf("unable to create IAM trust policy: %w", err)
	}

	params := &iam.CreateRoleInput{
		AssumeRolePolicyDocument: toPtr(string(trustPolicyDoc)),
		RoleName:                 toPtr(odrIAMRoleName(req.OrgID)),
		MaxSessionDuration:       toPtr(int32(defaultIAMRoleSessionDuration.Seconds())),
		Path:                     toPtr("/orgs/"),
		Tags: []iam_types.Tag{
			{
				Key:   toPtr("org-id"),
				Value: toPtr(req.OrgID),
			},
			{
				Key:   toPtr("managed-by"),
				Value: toPtr("workers-orgs"),
			},
		},
	}

	_, err = client.CreateRole(ctx, params)
	if err != nil {
		return fmt.Errorf("unable to create IAM role: %w", err)
	}

	return nil
}

func (o *odrIAMRoleCreatorImpl) createOdrIAMRolePolicyAttachment(ctx context.Context, client awsClientIAMRole, policyArn string, req CreateOdrIAMRoleRequest) error {
	params := &iam.AttachRolePolicyInput{
		PolicyArn: toPtr(policyArn),
		RoleName:  toPtr(odrIAMRoleName(req.OrgID)),
	}
	_, err := client.AttachRolePolicy(ctx, params)
	if err != nil {
		return fmt.Errorf("unable to attach role policy: %w", err)
	}

	return nil
}

type IAMRoleTrustStatement struct {
	Action    string `json:"Action,omitempty"`
	Effect    string `json:"Effect,omitempty"`
	Resource  string `json:"Resource,omitempty"`
	Sid       string `json:"Sid"`
	Principal struct {
		Federated string `json:"Federated,omitempty"`
	} `json:"Principal,omitempty"`
	Condition struct {
		StringEquals map[string]string `json:"StringEquals"`
	} `json:"Condition,omitempty"`
}

type IAMRoleStatement struct {
	Action   []string `json:"Action,omitempty"`
	Effect   string   `json:"Effect,omitempty"`
	Resource string   `json:"Resource,omitempty"`
	Sid      string   `json:"Sid"`
}

type odrIAMRoleTrustPolicy struct {
	Version   string                  `json:"Version"`
	Statement []IAMRoleTrustStatement `json:"Statement"`
}

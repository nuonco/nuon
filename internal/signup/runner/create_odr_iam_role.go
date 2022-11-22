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

	OrgsIAMOidcFederationRoleArn string `validate:"required" json:"orgs_iam_oidc_federation_role_arn"`
	OrgsIAMOidcProviderURL       string `validate:"required" json:"orgs_iam_oidc_provider_url"`
	OrgsIAMAccessRoleArn         string `validate:"required" json:"orgs_iam_access_role_arn"`
	ECRRegistryID                string `validate:"required" json:"ecr_registry_id"`
}

type CreateOdrIAMRoleResponse struct{}

func (a *Activities) CreateOdrIAMRole(ctx context.Context, req CreateOdrIAMRoleRequest) (CreateOdrIAMRoleResponse, error) {
	var resp CreateOdrIAMRoleResponse
	if err := req.validate(); err != nil {
		return resp, fmt.Errorf("unable to validate request: %w", err)
	}

	return resp, nil
}

func (r CreateOdrIAMRoleRequest) validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

type odrIAMRoleCreator interface {
	createOdrIAMPolicy(context.Context, awsClientIAM, *CreateOdrIAMRoleRequest) (string, error)
	createOdrIAMRole(context.Context, awsClientIAM, *CreateOdrIAMRoleRequest) error
	createOdrIAMRolePolicyAttachment(context.Context, awsClientIAM, string, *CreateOdrIAMRoleRequest) error
}

var _ odrIAMRoleCreator = (*odrIAMRoleCreatorImpl)(nil)

type odrIAMRoleCreatorImpl struct{}

type awsClientIAM interface {
	CreatePolicy(context.Context, *iam.CreatePolicyInput, ...func(iam.Options)) (*iam.CreatePolicyOutput, error)
	CreateRole(context.Context, *iam.CreateRoleInput, ...func(iam.Options)) (*iam.CreateRoleOutput, error)
	AttachRolePolicy(context.Context, *iam.AttachRolePolicyInput, ...func(iam.Options)) (*iam.AttachRolePolicyOutput, error)
}

func (o *odrIAMRoleCreatorImpl) createOdrIAMPolicy(ctx context.Context, client awsClientIAM, req *CreateOdrIAMRoleRequest) (string, error) {
	policy := odrIAMRolePolicy{
		Version: defaultIAMPolicyVersion,
		Statement: []IAMRoleStatement{
			{
				Effect: "Allow",
				Action: []string{
					"ecr:*",
				},
				Resource: fmt.Sprintf("%s/%s/*", req.ECRRegistryID, req.OrgID),
			},
		},
	}
	byts, err := json.Marshal(policy)
	if err != nil {
		return "", fmt.Errorf("unable to convert policy to json: %w", err)
	}

	params := &iam.CreatePolicyInput{
		PolicyDocument: toPtr(string(byts)),
	}

	output, err := client.CreatePolicy(ctx, params)
	if err != nil {
		return "", fmt.Errorf("unable to create policy: %w", err)
	}
	return *output.Policy.Arn, nil
}

func (o *odrIAMRoleCreatorImpl) createOdrIAMRole(ctx context.Context, client awsClientIAM, req *CreateOdrIAMRoleRequest) error {
	trustPolicy := odrIAMRoleTrustPolicy{
		Version: defaultIAMPolicyVersion,
		Statement: []IAMRoleStatement{
			{
				Action: []string{
					"sts:AssumeRoleWithWebIdentity",
				},
				Effect: "Allow",
				Sid:    "",
				Principal: struct {
					Federated string `json:"Federated,omitempty"`
				}{
					Federated: req.OrgsIAMOidcFederationRoleArn,
				},
				Condition: struct {
					StringEquals map[string]string `json:"StringEquals"`
				}{
					StringEquals: map[string]string{
						req.OrgsIAMOidcProviderURL: runnerOdrServiceAccountName(req.OrgID),
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

func (o *odrIAMRoleCreatorImpl) createOdrIAMRolePolicyAttachment(ctx context.Context, client awsClientIAM, policyArn string, req *CreateOdrIAMRoleRequest) error {
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

type IAMRoleStatement struct {
	Action    []string `json:"Action,omitempty"`
	Effect    string   `json:"Effect,omitempty"`
	Resource  string   `json:"Resource,omitempty"`
	Sid       string   `json:"Sid"`
	Principal struct {
		Federated string `json:"Federated,omitempty"`
	} `json:"Principal,omitempty"`
	Condition struct {
		StringEquals map[string]string `json:"StringEquals"`
	} `json:"Condition,omitempty"`
}

type odrIAMRoleTrustPolicy struct {
	Version   string             `json:"Version"`
	Statement []IAMRoleStatement `json:"statement"`
}

type odrIAMRolePolicy struct {
	Version   string             `json:"Version"`
	Statement []IAMRoleStatement `json:"Statement"`
}

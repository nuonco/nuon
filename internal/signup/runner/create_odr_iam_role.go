package runner

import (
	"context"

	"github.com/go-playground/validator/v10"
)

type CreateOdrIAMRoleRequest struct {
	OrgID string `validate:"required" json:"org_id"`

	OrgsIAMOidcProviderURL string `validate:"required" json:"orgs_iam_oidc_provider_url"`
	OrgsIAMAccessRoleArn   string `validate:"required" json:"orgs_iam_access_role_arn"`
	ECRRegistryID          string `validate:"required" json:"ecr_registry_id"`
}

type CreateOdrIAMRoleResponse struct{}

func (a *Activities) CreateOdrIAMRole(ctx context.Context, req CreateOdrIAMRoleRequest) (CreateOdrIAMRoleResponse, error) {
	var resp CreateOdrIAMRoleResponse

	return resp, nil
}

func (r CreateOdrIAMRoleRequest) validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

type odrIAMRoleCreator interface{}

type odrIAMRoleCreatorImpl struct{}

var _ odrIAMRoleCreator = (*odrIAMRoleCreatorImpl)(nil)

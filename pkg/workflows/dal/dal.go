package dal

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/aws/credentials"
	appsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/apps/v1"
	installsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/installs/v1"
	orgsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/orgs/v1"
)

var (
	requestFilename  string = "request.json"
	responseFilename string = "response.json"
)

//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=mock_dal.go -source=dal.go -package=dal
type Repo interface {
	// workflow request and responses
	GetInstallProvisionRequest(ctx context.Context, orgID, appID, installID string) (*installsv1.ProvisionRequest, error)
	GetInstallProvisionResponse(ctx context.Context, orgID, appID, installID string) (*installsv1.ProvisionResponse, error)

	GetOrgProvisionRequest(ctx context.Context, orgID string) (*orgsv1.SignupRequest, error)
	GetOrgProvisionResponse(ctx context.Context, orgID string) (*orgsv1.SignupResponse, error)

	GetAppProvisionRequest(ctx context.Context, orgID, appID string) (*appsv1.ProvisionRequest, error)
	GetAppProvisionResponse(ctx context.Context, orgID, appID string) (*appsv1.ProvisionResponse, error)

	// executors
}

type repo struct {
	v *validator.Validate

	Settings Settings `validate:"required"`
	Auth     *credentials.Config
	OrgId    string `validate:"unless Auth 1"`
}

var _ Repo = (*repo)(nil)

// New returns a default repo with the default orgcontext getter
func New(v *validator.Validate, opts ...repoOption) (*repo, error) {
	r := &repo{
		v: v,
	}
	for idx, opt := range opts {
		if err := opt(r); err != nil {
			return nil, fmt.Errorf("option %d failed: %w", idx, err)
		}
	}

	if err := r.v.Struct(r); err != nil {
		return nil, fmt.Errorf("unable to validate workflows repo: %w", err)
	}

	return r, nil
}

type repoOption func(*repo) error

type Settings struct {
	InstallsBucket                string
	InstallsBucketIAMRoleTemplate string

	OrgsBucket                string
	OrgsBucketIAMRoleTemplate string

	DeploymentsBucket                string
	DeploymentsBucketIAMRoleTemplate string

	AppsBucket                string
	AppsBucketIAMRoleTemplate string
}

// WithAuth is used to override the authentication, and use something like static credentials for instance
func WithAuth(auth *credentials.Config) repoOption {
	return func(r *repo) error {
		r.Auth = auth
		return nil
	}
}

// WithOrgId is used to set the org id, which will be used to create IAM roles
func WithOrgID(orgID string) repoOption {
	return func(r *repo) error {
		r.OrgId = orgID
		return nil
	}
}

// WithSettings adds the settings into the dal
func WithSettings(set Settings) repoOption {
	return func(r *repo) error {
		r.Settings = set
		return nil
	}
}

package workflows

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/nuonctl/internal"
	installsv1 "github.com/powertoolsdev/protos/workflows/generated/types/installs/v1"
	sharedv1 "github.com/powertoolsdev/protos/workflows/generated/types/shared/v1"
)

var (
	// TODO(jm): this should come from config and be set to user's email
	assumeRoleSessionName string = "nuonctl-workflows"

	requestFilename  string = "request.json"
	responseFilename string = "response.json"
)

type Repo interface {
	GetInstallProvisionRequest(ctx context.Context, installID string) (*installsv1.ProvisionRequest, error)

	GetInstanceProvisionRequest(context.Context, string, string, string, string, string) (*sharedv1.Request, error)
	GetInstanceProvisionResponse(context.Context, string, string, string, string, string) (*sharedv1.Response, error)

	GetOrgProvisionRequest(ctx context.Context, orgID string) (*sharedv1.Request, error)
	GetOrgProvisionResponse(ctx context.Context, orgID string) (*sharedv1.Response, error)

	GetAppProvisionRequest(ctx context.Context, orgID, appID string) (*sharedv1.Request, error)
	GetAppProvisionResponse(ctx context.Context, orgID, appID string) (*sharedv1.Response, error)

	GetDeploymentsRequest(ctx context.Context, key string) (*sharedv1.Request, error)
	GetDeploymentsResponse(ctx context.Context, key string) (*sharedv1.Response, error)
}

type repo struct {
	v *validator.Validate

	InstallsBucket    string `validate:"required"`
	AppsBucket        string `validate:"required"`
	OrgsBucket        string `validate:"required"`
	DeploymentsBucket string `validate:"required"`
	IAMRoleARN        string `validate:"required"`
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
		return nil, fmt.Errorf("unable to validate temporal: %w", err)
	}

	return r, nil
}

type repoOption func(*repo) error

func WithConfig(cfg *internal.Config) repoOption {
	return func(r *repo) error {
		r.InstallsBucket = cfg.InstallsBucket
		r.OrgsBucket = cfg.OrgsBucket
		r.AppsBucket = cfg.OrgsBucket
		r.DeploymentsBucket = cfg.DeploymentsBucket
		r.IAMRoleARN = cfg.SupportIAMRoleArn

		return nil
	}
}

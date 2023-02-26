package workflows

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	sharedv1 "github.com/powertoolsdev/protos/workflows/generated/types/shared/v1"
)

const (
	requestFilename  string = "request.json"
	responseFilename string = "response.json"
)

//
//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=workflows_mock.go -source=workflows.go -package=workflows
type Repo interface {
	GetOrgProvisionRequest(ctx context.Context, orgID string) (*sharedv1.Request, error)
	GetOrgProvisionResponse(ctx context.Context, orgID string) (*sharedv1.Response, error)

	GetAppProvisionRequest(ctx context.Context, orgID, appID string) (*sharedv1.Request, error)
	GetAppProvisionResponse(ctx context.Context, orgID, appID string) (*sharedv1.Response, error)

	GetInstallProvisionRequest(ctx context.Context, orgID, appID, installID string) (*sharedv1.Request, error)
	GetInstallProvisionResponse(ctx context.Context, orgID, appID, installID string) (*sharedv1.Response, error)

	GetDeploymentProvisionRequest(ctx context.Context, orgID, appID, componentID, deploymentID string) (*sharedv1.Request, error)
	GetDeploymentProvisionResponse(ctx context.Context, orgID, appID, componentID, deploymentID string) (*sharedv1.Response, error)

	GetInstanceProvisionRequest(ctx context.Context, orgID, appID, componentID, deploymentID, installID string) (*sharedv1.Request, error)
	GetInstanceProvisionResponse(ctx context.Context, orgID, appID, componentID, deploymentID, installID string) (*sharedv1.Response, error)
}

type Bucket struct {
	Name               string `validate:"required"`
	IamRoleArn         string `validate:"required"`
	IamRoleSessionName string `validate:"required"`
}

func (b Bucket) Validate() error {
	validate := validator.New()
	return validate.Struct(b)
}

type repo struct {
	v *validator.Validate

	OrgsBucket        Bucket `validate:"required"`
	AppsBucket        Bucket `validate:"required"`
	DeploymentsBucket Bucket `validate:"required"`
	InstallsBucket    Bucket `validate:"required"`
	InstancesBucket   Bucket `validate:"required"`
}

var _ Repo = (*repo)(nil)

// New returns a default repo with the default orgcontext getter
func New(v *validator.Validate, opts ...repoOption) (*repo, error) {
	r := &repo{
		v: v,
	}

	for _, opt := range opts {
		if err := opt(r); err != nil {
			return nil, err
		}
	}

	if err := v.Struct(r); err != nil {
		return nil, err
	}

	return r, nil
}

type repoOption func(*repo) error

// WithOrgsBucket sets the provided org bucket object
func WithOrgsBucket(bkt Bucket) repoOption {
	return func(r *repo) error {
		if err := bkt.Validate(); err != nil {
			return fmt.Errorf("invalid org bucket: %w", err)
		}

		r.OrgsBucket = bkt
		return nil
	}
}

// WithAppsBucket sets the provided org bucket object
func WithAppsBucket(bkt Bucket) repoOption {
	return func(r *repo) error {
		if err := bkt.Validate(); err != nil {
			return fmt.Errorf("invalid app bucket: %w", err)
		}

		r.AppsBucket = bkt
		return nil
	}
}

// WithInstallsBucket sets the provided deployments bucket object
func WithInstallsBucket(bkt Bucket) repoOption {
	return func(r *repo) error {
		if err := bkt.Validate(); err != nil {
			return fmt.Errorf("invalid installs bucket: %w", err)
		}

		r.InstallsBucket = bkt
		return nil
	}
}

// WithDeploymentsBucket sets the provided deployments bucket object
func WithDeploymentsBucket(bkt Bucket) repoOption {
	return func(r *repo) error {
		if err := bkt.Validate(); err != nil {
			return fmt.Errorf("invalid deployments bucket: %w", err)
		}

		r.DeploymentsBucket = bkt
		return nil
	}
}

// WithInstancesBucket sets the provided deployments bucket object
func WithInstancesBucket(bkt Bucket) repoOption {
	return func(r *repo) error {
		if err := bkt.Validate(); err != nil {
			return fmt.Errorf("invalid instances bucket: %w", err)
		}

		r.InstancesBucket = bkt
		return nil
	}
}

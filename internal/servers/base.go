package servers

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/orgs-api/internal/orgcontext"
	"github.com/powertoolsdev/orgs-api/internal/repos/waypoint"
	"github.com/powertoolsdev/orgs-api/internal/repos/workflows"
)

// Base is a struct that exposes shared methods between each Base
type Base struct {
	v           *validator.Validate
	CtxProvider orgcontext.Provider `validate:"required"`
}

type BaseOption func(*Base) error

func New(v *validator.Validate, opts ...BaseOption) (*Base, error) {
	srv := &Base{
		v: v,
	}

	for idx, opt := range opts {
		if err := opt(srv); err != nil {
			return nil, fmt.Errorf("option %d failed: %w", idx, err)
		}
	}

	if err := srv.v.Struct(srv); err != nil {
		return nil, fmt.Errorf("unable to validate base server: %w", err)
	}
	return srv, nil
}

// WithContextProvider sets the context provider on the object
func WithContextProvider(ctxProvider orgcontext.Provider) BaseOption {
	return func(s *Base) error {
		s.CtxProvider = ctxProvider
		return nil
	}
}

// GetWorkflowsRepo returns a repo for interacting with an orgs workflows metadata
func (b *Base) WorkflowsRepo(ctx context.Context, orgID string) (workflows.Repo, error) {
	orgCtx, err := b.CtxProvider.Get(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("unable to get org context: %w", err)
	}

	repo, err := workflows.New(b.v,
		workflows.WithOrgsBucket(workflows.Bucket(orgCtx.Buckets[orgcontext.BucketTypeOrgs])),
		workflows.WithAppsBucket(workflows.Bucket(orgCtx.Buckets[orgcontext.BucketTypeApps])),
		workflows.WithInstallsBucket(workflows.Bucket(orgCtx.Buckets[orgcontext.BucketTypeInstalls])),
		workflows.WithDeploymentsBucket(workflows.Bucket(orgCtx.Buckets[orgcontext.BucketTypeDeployments])),
		workflows.WithInstancesBucket(workflows.Bucket(orgCtx.Buckets[orgcontext.BucketTypeInstances])),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to get workflows repo: %w", err)
	}

	return repo, nil
}

// GetWaypointRepo returns a repo for interacting with an org's waypoint Base
func (b *Base) WaypointRepo(ctx context.Context, orgID string) (waypoint.Repo, error) {
	orgCtx, err := b.CtxProvider.Get(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("unable to get org context: %w", err)
	}

	repo, err := waypoint.New(b.v,
		waypoint.WithAddress(orgCtx.WaypointServer.Address),
		waypoint.WithSecretName(orgCtx.WaypointServer.SecretName),
		waypoint.WithSecretNamespace(orgCtx.WaypointServer.SecretNamespace),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to get waypoint repo: %w", err)
	}

	return repo, nil
}

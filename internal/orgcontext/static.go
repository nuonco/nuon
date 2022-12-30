package orgcontext

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/go-waypoint"
	"github.com/powertoolsdev/orgs-api/internal"
)

type staticProvider struct {
	validate *validator.Validate `validate:"required"`

	DeploymentsBucketName               string `validate:"required"`
	DeploymentsBucketAssumeRoleTemplate string `validate:"required"`

	InstallationsBucketName               string `validate:"required"`
	InstallationsBucketAssumeRoleTemplate string `validate:"required"`

	OrgsBucketName               string `validate:"required"`
	OrgsBucketAssumeRoleTemplate string `validate:"required"`

	WaypointTokenSecretNamespace string `validate:"required"`
	WaypointTokenSecretTemplate  string `validate:"required"`
	WaypointServerRootDomain     string `validate:"required"`
}

func (s *staticProvider) createContext(orgID string) *Context {
	return &Context{
		Buckets: map[BucketType]Bucket{
			BucketTypeDeployments: {
				Prefix:         fmt.Sprintf("org=%s/", orgID),
				Name:           s.DeploymentsBucketName,
				AssumeRoleARN:  fmt.Sprintf(s.DeploymentsBucketAssumeRoleTemplate, orgID),
				AssumeRoleName: defaultAssumeRoleName,
			},
			BucketTypeOrgs: {
				Prefix:         fmt.Sprintf("org=%s/", orgID),
				Name:           s.OrgsBucketName,
				AssumeRoleARN:  fmt.Sprintf(s.OrgsBucketAssumeRoleTemplate, orgID),
				AssumeRoleName: defaultAssumeRoleName,
			},
			BucketTypeInstallations: {
				Name:           s.InstallationsBucketName,
				Prefix:         fmt.Sprintf("org=%s/", orgID),
				AssumeRoleARN:  fmt.Sprintf(s.InstallationsBucketAssumeRoleTemplate, orgID),
				AssumeRoleName: defaultAssumeRoleName,
			},
		},
		WaypointServer: WaypointServer{
			Address:         waypoint.DefaultOrgServerAddress(s.WaypointServerRootDomain, orgID),
			SecretNamespace: s.WaypointTokenSecretNamespace,
			SecretName:      fmt.Sprintf(s.WaypointTokenSecretTemplate, orgID),
		},
	}
}

func (s *staticProvider) SetContext(ctx context.Context, orgID string) (context.Context, error) {
	orgCtx := s.createContext(orgID)
	if err := s.validate.Struct(orgCtx); err != nil {
		return nil, fmt.Errorf("invalid context: %w", err)
	}

	for bktType, bkt := range orgCtx.Buckets {
		if err := s.validate.Struct(bkt); err != nil {
			return nil, fmt.Errorf("invalid %s bucket: %w", bktType, err)
		}
	}

	return context.WithValue(ctx, orgContextKey{}, orgCtx), nil
}

var _ provider = (*staticProvider)(nil)

func NewStaticProvider(opts ...staticOption) (*staticProvider, error) {
	spr := &staticProvider{
		validate: validator.New(),
	}

	for idx, opt := range opts {
		if err := opt(spr); err != nil {
			return nil, fmt.Errorf("option %d failed: %w", idx, err)
		}
	}

	if err := spr.validate.Struct(spr); err != nil {
		return nil, fmt.Errorf("unable to validate server: %w", err)
	}

	return spr, nil
}

type staticOption func(*staticProvider) error

// WithConfig is an option that allows us to encapsulate which fields we need on the provider itself, without actually
// needing to pass the config around or do a bunch of "hand wiring" outside of this context.
func WithConfig(cfg *internal.Config) staticOption {
	return func(s *staticProvider) error {
		s.DeploymentsBucketName = cfg.DeploymentsBucket
		s.DeploymentsBucketAssumeRoleTemplate = cfg.OrgsDeploymentsRoleTemplate

		s.InstallationsBucketName = cfg.InstallationsBucket
		s.InstallationsBucketAssumeRoleTemplate = cfg.OrgsInstallationsRoleTemplate

		s.OrgsBucketName = cfg.OrgsBucket
		s.OrgsBucketAssumeRoleTemplate = cfg.OrgsOrgsBucketRoleTemplate

		s.WaypointServerRootDomain = cfg.WaypointServerRootDomain
		s.WaypointTokenSecretTemplate = cfg.WaypointTokenSecretTemplate
		s.WaypointTokenSecretNamespace = cfg.WaypointTokenSecretNamespace

		return nil
	}
}

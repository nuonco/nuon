package orgcontext

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/go-waypoint"
	"github.com/powertoolsdev/orgs-api/internal"
)

type staticProvider struct {
	v *validator.Validate `validate:"required"`

	DeploymentsBucketName               string `validate:"required"`
	DeploymentsBucketAssumeRoleTemplate string `validate:"required"`

	InstallsBucketName          string `validate:"required"`
	InstallsBucketAssumeRoleARN string `validate:"required"`

	OrgsBucketName               string `validate:"required"`
	OrgsBucketAssumeRoleTemplate string `validate:"required"`

	AppsBucketName               string `validate:"required"`
	AppsBucketAssumeRoleTemplate string `validate:"required"`

	InstancesBucketName               string `validate:"required"`
	InstancesBucketAssumeRoleTemplate string `validate:"required"`

	WaypointTokenSecretNamespace string `validate:"required"`
	WaypointTokenSecretTemplate  string `validate:"required"`
	WaypointServerRootDomain     string `validate:"required"`
}

func (s *staticProvider) Get(_ context.Context, orgID string) (*Context, error) {
	return &Context{
		OrgID: orgID,
		Buckets: map[BucketType]Bucket{
			BucketTypeOrgs: {
				Name:               s.OrgsBucketName,
				IamRoleArn:         fmt.Sprintf(s.OrgsBucketAssumeRoleTemplate, orgID),
				IamRoleSessionName: defaultAssumeRoleName,
			},
			BucketTypeApps: {
				Name:               s.AppsBucketName,
				IamRoleArn:         fmt.Sprintf(s.AppsBucketAssumeRoleTemplate, orgID),
				IamRoleSessionName: defaultAssumeRoleName,
			},
			BucketTypeInstalls: {
				Name:               s.InstallsBucketName,
				IamRoleArn:         fmt.Sprintf(s.InstallsBucketAssumeRoleARN, orgID),
				IamRoleSessionName: defaultAssumeRoleName,
			},
			BucketTypeDeployments: {
				Name:               s.DeploymentsBucketName,
				IamRoleArn:         fmt.Sprintf(s.DeploymentsBucketAssumeRoleTemplate, orgID),
				IamRoleSessionName: defaultAssumeRoleName,
			},
			BucketTypeInstances: {
				Name:               s.InstancesBucketName,
				IamRoleArn:         fmt.Sprintf(s.InstancesBucketAssumeRoleTemplate, orgID),
				IamRoleSessionName: defaultAssumeRoleName,
			},
		},
		WaypointServer: WaypointServer{
			Address:         waypoint.DefaultOrgServerAddress(s.WaypointServerRootDomain, orgID),
			SecretNamespace: s.WaypointTokenSecretNamespace,
			SecretName:      fmt.Sprintf(s.WaypointTokenSecretTemplate, orgID),
		},
	}, nil
}

var _ provider = (*staticProvider)(nil)

func NewStaticProvider(v *validator.Validate, opts ...staticOption) (*staticProvider, error) {
	spr := &staticProvider{
		v: v,
	}

	for idx, opt := range opts {
		if err := opt(spr); err != nil {
			return nil, fmt.Errorf("option %d failed: %w", idx, err)
		}
	}

	if err := spr.v.Struct(spr); err != nil {
		return nil, fmt.Errorf("unable to validate server: %w", err)
	}

	return spr, nil
}

type staticOption func(*staticProvider) error

// WithConfig is an option that allows us to encapsulate which fields we need on the provider itself, without actually
// needing to pass the config around or do a bunch of "hand wiring" outside of this context.
func WithConfig(cfg *internal.Config) staticOption {
	return func(s *staticProvider) error {
		s.OrgsBucketName = cfg.OrgsBucket
		s.OrgsBucketAssumeRoleTemplate = cfg.OrgsOrgsBucketRoleTemplate

		s.AppsBucketName = cfg.OrgsBucket
		s.AppsBucketAssumeRoleTemplate = cfg.OrgsOrgsBucketRoleTemplate

		s.InstallsBucketName = cfg.InstallationsBucket
		s.InstallsBucketAssumeRoleARN = cfg.OrgsInstallationsRoleTemplate

		s.DeploymentsBucketName = cfg.DeploymentsBucket
		s.DeploymentsBucketAssumeRoleTemplate = cfg.OrgsDeploymentsRoleTemplate

		s.InstancesBucketName = cfg.DeploymentsBucket
		s.InstancesBucketAssumeRoleTemplate = cfg.OrgsDeploymentsRoleTemplate

		s.WaypointServerRootDomain = cfg.WaypointServerRootDomain
		s.WaypointTokenSecretTemplate = cfg.WaypointTokenSecretTemplate
		s.WaypointTokenSecretNamespace = cfg.WaypointTokenSecretNamespace

		return nil
	}
}

package orgcontext

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/powertoolsdev/go-generics"
	"github.com/powertoolsdev/orgs-api/internal"
	"github.com/stretchr/testify/assert"
)

func TestWithConfig(t *testing.T) {
	cfg := generics.GetFakeObj[*internal.Config]()
	s := &staticProvider{}

	assert.NoError(t, WithConfig(cfg)(s))

	// assert bucket values
	assert.Equal(t, cfg.DeploymentsBucket, s.DeploymentsBucketName)
	assert.Equal(t, cfg.OrgsDeploymentsRoleTemplate, s.DeploymentsBucketAssumeRoleTemplate)
	assert.Equal(t, s.DeploymentsBucketAssumeRoleTemplate, cfg.OrgsDeploymentsRoleTemplate)
	assert.Equal(t, s.InstallationsBucketName, cfg.InstallationsBucket)
	assert.Equal(t, s.InstallationsBucketAssumeRoleTemplate, cfg.OrgsInstallationsRoleTemplate)
	assert.Equal(t, s.OrgsBucketName, cfg.OrgsBucket)
	assert.Equal(t, s.OrgsBucketAssumeRoleTemplate, cfg.OrgsOrgsBucketRoleTemplate)

	// assert waypoint values
	assert.Equal(t, s.WaypointServerRootDomain, cfg.WaypointServerRootDomain)
	assert.Equal(t, s.WaypointTokenSecretTemplate, cfg.WaypointTokenSecretTemplate)
	assert.Equal(t, s.WaypointTokenSecretNamespace, cfg.WaypointTokenSecretNamespace)
}

func TestNewStaticProvider(t *testing.T) {
	testOptErr := fmt.Errorf("an option did not successfully return")
	cfg := generics.GetFakeObj[*internal.Config]()

	tests := map[string]struct {
		option      staticOption
		assertFn    func(*testing.T, *staticProvider)
		errExpected error
	}{
		"happy path": {
			option: WithConfig(cfg),
			assertFn: func(t *testing.T, prov *staticProvider) {
			},
		},
		"missing value": {
			option: func(s *staticProvider) error {
				assert.NoError(t, WithConfig(cfg)(s))
				s.DeploymentsBucketName = ""
				return nil
			},
			errExpected: fmt.Errorf("validate"),
		},
		"error": {
			option: func(s *staticProvider) error {
				return testOptErr
			},
			assertFn: func(t *testing.T, prov *staticProvider) {
			},
			errExpected: testOptErr,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			provider, err := NewStaticProvider(test.option)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
			test.assertFn(t, provider)
		})
	}
}

func Test_staticProvider_SetAndGetContext(t *testing.T) {
	provider := generics.GetFakeObj[*staticProvider]()
	provider.validate = validator.New()
	orgID := uuid.NewString()

	tests := map[string]struct {
		providerFn  func(*testing.T) *staticProvider
		assertFn    func(*testing.T, context.Context)
		errExpected error
	}{
		"happy path": {
			providerFn: func(t *testing.T) *staticProvider {
				return provider
			},
			assertFn: func(t *testing.T, ctx context.Context) {
				expected := provider.createContext(orgID)
				returned, err := Get(ctx)
				assert.NoError(t, err)
				assert.Equal(t, expected, returned)
			},
		},
		"missing value in provider": {
			providerFn: func(t *testing.T) *staticProvider {
				provider := generics.GetFakeObj[*staticProvider]()
				provider.validate = validator.New()
				provider.DeploymentsBucketName = ""
				return provider
			},
			errExpected: fmt.Errorf("Context.DeploymentsBucket.Name"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			provider := test.providerFn(t)
			ctx, err := provider.SetContext(context.Background(), orgID)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
			test.assertFn(t, ctx)
		})
	}
}

func Test_staticProvider_createContext(t *testing.T) {
	s := generics.GetFakeObj[*staticProvider]()
	s.validate = validator.New()
	assert.NoError(t, s.validate.Struct(s))
	orgID := uuid.NewString()

	ctx := s.createContext(orgID)
	assert.NoError(t, s.validate.Struct(ctx))

	expectedPrefix := fmt.Sprintf("org=%s/", orgID)
	assert.Equal(t, s.OrgsBucketName, ctx.OrgsBucket.Name)
	assert.Equal(t, expectedPrefix, ctx.OrgsBucket.Prefix)
	assert.Contains(t, ctx.OrgsBucket.AssumeRoleARN, orgID)

	assert.Equal(t, s.DeploymentsBucketName, ctx.DeploymentsBucket.Name)
	assert.Equal(t, expectedPrefix, ctx.DeploymentsBucket.Prefix)
	assert.Contains(t, ctx.DeploymentsBucket.AssumeRoleARN, orgID)

	assert.Equal(t, s.InstallationsBucketName, ctx.InstallationsBucket.Name)
	assert.Equal(t, expectedPrefix, ctx.InstallationsBucket.Prefix)
	assert.Contains(t, ctx.InstallationsBucket.AssumeRoleARN, orgID)

	assert.Equal(t, s.WaypointTokenSecretNamespace, ctx.WaypointServer.SecretNamespace)
	assert.Contains(t, ctx.WaypointServer.SecretName, orgID)
	assert.Contains(t, ctx.WaypointServer.Address, orgID)
	assert.Contains(t, ctx.WaypointServer.Address, s.WaypointServerRootDomain)
}

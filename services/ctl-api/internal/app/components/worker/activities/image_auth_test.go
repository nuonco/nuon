package activities

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"

	ecr "github.com/nuonco/nuon/pkg/aws/ecr-authorization"
	"github.com/nuonco/nuon/pkg/oci/metadata"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	cloudconnections "github.com/nuonco/nuon/services/ctl-api/internal/app/cloud-connections"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
)

type metadataAuthorizer struct{ mock.Mock }

func (m *metadataAuthorizer) ECRAuthorization(ctx context.Context, org string, cfg *app.AWSECRImageConfig, image string) (*ecr.Authorization, error) {
	args := m.Called(org, cfg, image)
	return &ecr.Authorization{Username: "AWS", RegistryToken: "acme-token", ServerAddress: "acme"}, args.Error(0)
}

func (m *metadataAuthorizer) GARAuthorization(ctx context.Context, org string, cfg *app.GCPGARImageConfig, image string) (*oauth2.Token, error) {
	args := m.Called(org, cfg, image)
	return &oauth2.Token{AccessToken: "acme-token"}, args.Error(0)
}

func (m *metadataAuthorizer) ACRAuthorization(ctx context.Context, org string, cfg *app.AzureACRImageConfig, image string) (*metadata.RegistryAuth, error) {
	args := m.Called(org, cfg, image)
	if err := args.Error(0); err != nil {
		return nil, err
	}
	return &metadata.RegistryAuth{Password: "acme-token"}, nil
}

func (m *metadataAuthorizer) GetACRAccessToken(ctx context.Context, req *sharedactivities.GetACRAccessTokenRequest) (*sharedactivities.ACRAccessToken, error) {
	args := m.Called(req)
	return &sharedactivities.ACRAccessToken{Password: "acme-token"}, args.Error(0)
}

func TestDirectMetadataAuthorization(t *testing.T) {
	for name, tc := range map[string]struct {
		registry, method string
		legacy           bool
	}{
		"legacy ECR":  {registry: "ecr", method: "ECRAuthorization", legacy: true},
		"pinned ECR":  {registry: "ecr", method: "ECRAuthorization"},
		"legacy GAR":  {registry: "gar", method: "GARAuthorization", legacy: true},
		"pinned GAR":  {registry: "gar", method: "GARAuthorization"},
		"ambient ACR": {registry: "acr", method: "GetACRAccessToken", legacy: true},
		"pinned ACR":  {registry: "acr", method: "ACRAuthorization"},
	} {
		for result, authError := range map[string]error{"success": nil, "failure": errors.New("registry authorization denied")} {
			t.Run(name+"/"+result, func(t *testing.T) {
				mockAuth := &metadataAuthorizer{}
				a := &Activities{cfg: &internal.Config{LegacyImageAuthEnabled: true}, cloudConnectionsHelpers: mockAuth, sharedActs: mockAuth}
				pin := "cc_saved"
				if tc.legacy {
					pin = ""
				}
				var auth *metadata.RegistryAuth
				var err error
				switch tc.registry {
				case "ecr":
					cfg := &app.AWSECRImageConfig{ID: "saved", CloudConnectionID: pin}
					if tc.legacy {
						cfg.IAMRoleARN = "role"
					}
					mockAuth.On(tc.method, "org_acme", cfg, "image").Return(authError).Once()
					auth, err = a.getECRAuth(context.Background(), "org_acme", cfg, "image")
				case "gar":
					cfg := &app.GCPGARImageConfig{ID: "saved", CloudConnectionID: pin}
					if tc.legacy {
						cfg.ServiceAccountEmail = "acme@example.com"
					}
					mockAuth.On(tc.method, "org_acme", cfg, "image").Return(authError).Once()
					auth, err = a.getGARAuth(context.Background(), "org_acme", cfg, "image")
				case "acr":
					cfg := &app.AzureACRImageConfig{ID: "saved", CloudConnectionID: pin, RegistryURL: "acme.azurecr.io"}
					if tc.legacy {
						mockAuth.On(tc.method, &sharedactivities.GetACRAccessTokenRequest{ComponentID: "component", LoginServer: cfg.RegistryURL}).Return(authError).Once()
					} else {
						mockAuth.On(tc.method, "org_acme", cfg, "image").Return(authError).Once()
					}
					auth, err = a.getACRAuth(context.Background(), "org_acme", cfg, "image", "component")
				}
				if authError != nil {
					require.ErrorIs(t, err, authError)
					require.Nil(t, auth)
				} else {
					require.NoError(t, err)
					require.Equal(t, "acme-token", auth.Password)
				}
				mockAuth.AssertExpectations(t)
			})
		}
	}
}

func TestDirectMetadataLegacyACRDisabled(t *testing.T) {
	for name, tc := range map[string]struct {
		config  app.AzureACRImageConfig
		allowed bool
	}{
		"ambient":     {},
		"secret":      {config: app.AzureACRImageConfig{TenantID: "tenant", ClientID: "client", ClientSecretName: "secret"}, allowed: true},
		"certificate": {config: app.AzureACRImageConfig{TenantID: "tenant", ClientID: "client", ClientCertificateName: "cert"}, allowed: true},
	} {
		t.Run(name, func(t *testing.T) {
			mockAuth := &metadataAuthorizer{}
			tc.config.RegistryURL = "acme.azurecr.io"
			if tc.allowed {
				mockAuth.On("GetACRAccessToken", mock.Anything).Return(nil).Once()
			}
			a := &Activities{cfg: &internal.Config{}, cloudConnectionsHelpers: mockAuth, sharedActs: mockAuth}
			auth, err := a.getACRAuth(context.Background(), "org_acme", &tc.config, "image", "component")
			if tc.allowed {
				require.NoError(t, err)
				require.Equal(t, "acme-token", auth.Password)
			} else {
				require.ErrorIs(t, err, cloudconnections.ErrLegacyImageAuthDisabled)
				mockAuth.AssertNotCalled(t, "GetACRAccessToken", mock.Anything)
			}
			mockAuth.AssertExpectations(t)
		})
	}
}

func TestMetadataACRPlanModes(t *testing.T) {
	for name, tc := range map[string]struct {
		config                   app.AzureACRImageConfig
		connection, registration bool
	}{
		"ambient":         {},
		"explicit opt in": {config: app.AzureACRImageConfig{UseCloudConnection: true}, connection: true},
		"saved pin":       {config: app.AzureACRImageConfig{CloudConnectionID: "cc_saved"}, connection: true},
		"secret":          {config: app.AzureACRImageConfig{TenantID: "tenant", ClientID: "client", ClientSecretName: "secret"}, registration: true},
		"certificate":     {config: app.AzureACRImageConfig{TenantID: "tenant", ClientID: "client", ClientCertificateName: "cert"}, registration: true},
	} {
		t.Run(name, func(t *testing.T) {
			tc.config.RegistryURL = "acme.azurecr.io"
			a := &Activities{cfg: &internal.Config{}}
			repo, err := a.getSourceRepository(&app.ExternalImageComponentConfig{ImageURL: "acme.azurecr.io/repo", AzureACRImageConfig: &tc.config}, "component")
			require.NoError(t, err)
			if tc.connection {
				require.Nil(t, repo.ACRAuth)
			} else {
				require.True(t, repo.ACRAuth.UseDefault)
			}
			require.Equal(t, tc.registration, repo.ACRAppRegistration != nil)
		})
	}
}

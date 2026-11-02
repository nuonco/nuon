package build

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	cloudconnections "github.com/nuonco/nuon/services/ctl-api/internal/app/cloud-connections"
)

func TestExternalImageAuthenticationConversion(t *testing.T) {
	for name, tc := range map[string]struct {
		config    *config.ExternalImageComponentConfig
		check     func(*testing.T, *app.ExternalImageComponentConfig)
		wantError bool
		legacy    bool
	}{
		"legacy ECR": {
			config: &config.ExternalImageComponentConfig{AWSECRImageConfig: &config.AWSECRConfig{IAMRoleARN: "arn:aws:iam::123456789012:role/acme", AWSRegion: "us-east-1", ImageURL: "123456789012.dkr.ecr.us-east-1.amazonaws.com/acme", Tag: "v1"}},
			legacy: true,
			check: func(t *testing.T, cfg *app.ExternalImageComponentConfig) {
				require.Equal(t, "arn:aws:iam::123456789012:role/acme", cfg.AWSECRImageConfig.IAMRoleARN)
			},
		},
		"ambient GAR": {
			config: &config.ExternalImageComponentConfig{GCPGARImageConfig: &config.GCPGARConfig{GCPRegion: "us-central1", GCPProjectID: "acme", ImageURL: "us-central1-docker.pkg.dev/acme/repo/image", Tag: "v1"}},
			legacy: true,
			check: func(t *testing.T, cfg *app.ExternalImageComponentConfig) {
				require.False(t, cfg.GCPGARImageConfig.UseCloudConnection)
				require.Empty(t, cfg.GCPGARImageConfig.ServiceAccountEmail)
				require.Empty(t, cfg.GCPGARImageConfig.WorkloadIdentityProvider)
			},
		},
		"connection GAR": {
			config: &config.ExternalImageComponentConfig{GCPGARImageConfig: &config.GCPGARConfig{UseCloudConnection: true, GCPRegion: "us-central1", GCPProjectID: "acme", ImageURL: "us-central1-docker.pkg.dev/acme/repo/image", Tag: "v1"}},
			check: func(t *testing.T, cfg *app.ExternalImageComponentConfig) {
				require.True(t, cfg.GCPGARImageConfig.UseCloudConnection)
			},
		},
		"mixed GAR account":  {config: &config.ExternalImageComponentConfig{GCPGARImageConfig: &config.GCPGARConfig{UseCloudConnection: true, ServiceAccountEmail: "acme@example.com"}}, wantError: true},
		"mixed GAR provider": {config: &config.ExternalImageComponentConfig{GCPGARImageConfig: &config.GCPGARConfig{UseCloudConnection: true, WorkloadIdentityProvider: "provider"}}, wantError: true},
		"legacy GAR impersonation": {
			config: &config.ExternalImageComponentConfig{GCPGARImageConfig: &config.GCPGARConfig{ServiceAccountEmail: "acme@example.com", GCPRegion: "us-central1", GCPProjectID: "acme", ImageURL: "us-central1-docker.pkg.dev/acme/repo/image", Tag: "v1"}},
			legacy: true,
			check: func(t *testing.T, cfg *app.ExternalImageComponentConfig) {
				require.Equal(t, "acme@example.com", cfg.GCPGARImageConfig.ServiceAccountEmail)
			},
		},
		"legacy GAR federation": {
			config: &config.ExternalImageComponentConfig{GCPGARImageConfig: &config.GCPGARConfig{WorkloadIdentityProvider: "projects/123/locations/global/workloadIdentityPools/acme/providers/acme", GCPRegion: "us-central1", GCPProjectID: "acme", ImageURL: "us-central1-docker.pkg.dev/acme/repo/image", Tag: "v1"}},
			legacy: true,
			check: func(t *testing.T, cfg *app.ExternalImageComponentConfig) {
				require.NotEmpty(t, cfg.GCPGARImageConfig.WorkloadIdentityProvider)
			},
		},
		"ambient ACR": {
			config: &config.ExternalImageComponentConfig{AzureACRImageConfig: &config.AzureACRConfig{RegistryURL: "acme.azurecr.io", ImageURL: "acme.azurecr.io/repo", Tag: "v1"}},
			legacy: true,
			check: func(t *testing.T, cfg *app.ExternalImageComponentConfig) {
				require.False(t, cfg.AzureACRImageConfig.UsesCloudConnection())
			},
		},
		"connection ACR": {
			config: &config.ExternalImageComponentConfig{AzureACRImageConfig: &config.AzureACRConfig{UseCloudConnection: true, RegistryURL: "acme.azurecr.io", ImageURL: "acme.azurecr.io/repo", Tag: "v1"}},
			check: func(t *testing.T, cfg *app.ExternalImageComponentConfig) {
				require.True(t, cfg.AzureACRImageConfig.UsesCloudConnection())
			},
		},
		"secret ACR": {
			config: &config.ExternalImageComponentConfig{AzureACRImageConfig: &config.AzureACRConfig{RegistryURL: "acme.azurecr.io", ImageURL: "acme.azurecr.io/repo", Tag: "v1", TenantID: "tenant", ClientID: "client", ClientSecretName: "acme-secret"}},
			legacy: true,
			check: func(t *testing.T, cfg *app.ExternalImageComponentConfig) {
				require.Equal(t, "acme-secret", cfg.AzureACRImageConfig.ClientSecretName)
				require.False(t, cfg.AzureACRImageConfig.UsesCloudConnection())
			},
		},
		"certificate ACR": {
			config: &config.ExternalImageComponentConfig{AzureACRImageConfig: &config.AzureACRConfig{RegistryURL: "acme.azurecr.io", ImageURL: "acme.azurecr.io/repo", Tag: "v1", TenantID: "tenant", ClientID: "client", ClientCertificateName: "acme-cert"}},
			legacy: true,
			check: func(t *testing.T, cfg *app.ExternalImageComponentConfig) {
				require.Equal(t, "acme-cert", cfg.AzureACRImageConfig.ClientCertificateName)
			},
		},
		"mixed ACR":      {config: &config.ExternalImageComponentConfig{AzureACRImageConfig: &config.AzureACRConfig{UseCloudConnection: true, TenantID: "tenant"}}, wantError: true},
		"incomplete ACR": {config: &config.ExternalImageComponentConfig{AzureACRImageConfig: &config.AzureACRConfig{ClientID: "client"}}, wantError: true},
	} {
		t.Run(name, func(t *testing.T) {
			validationErr := tc.config.Validate()
			cfg, err := ExternalImageComponentConfig(tc.config)
			if tc.wantError {
				require.Error(t, validationErr)
				require.Error(t, err)
				return
			}
			require.NoError(t, validationErr)
			require.NoError(t, err)
			tc.check(t, cfg)
			if tc.legacy {
				connection := &app.ComponentConfigConnection{Checksum: "unchanged", ExternalImageComponentConfig: cfg}
				for range 2 {
					require.NoError(t, cloudconnections.ResolveImageConnection(context.Background(), nil, "org_acme", connection))
					require.Equal(t, "unchanged", connection.Checksum)
				}
			}
		})
	}
}

package build

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestExternalImageAuthenticationConversion(t *testing.T) {
	for name, tc := range map[string]struct {
		config    *config.ExternalImageComponentConfig
		check     func(*testing.T, *app.ExternalImageComponentConfig)
		wantError bool
	}{
		"legacy ECR": {
			config: &config.ExternalImageComponentConfig{AWSECRImageConfig: &config.AWSECRConfig{IAMRoleARN: "arn:aws:iam::123456789012:role/acme", AWSRegion: "us-east-1", ImageURL: "123456789012.dkr.ecr.us-east-1.amazonaws.com/acme", Tag: "v1"}},
			check: func(t *testing.T, cfg *app.ExternalImageComponentConfig) {
				require.Equal(t, "arn:aws:iam::123456789012:role/acme", cfg.AWSECRImageConfig.IAMRoleARN)
			},
		},
		"ambient GAR": {
			config: &config.ExternalImageComponentConfig{GCPGARImageConfig: &config.GCPGARConfig{GCPRegion: "us-central1", GCPProjectID: "acme", ImageURL: "us-central1-docker.pkg.dev/acme/repo/image", Tag: "v1"}},
			check: func(t *testing.T, cfg *app.ExternalImageComponentConfig) {
				require.Empty(t, cfg.GCPGARImageConfig.ServiceAccountEmail)
				require.Empty(t, cfg.GCPGARImageConfig.WorkloadIdentityProvider)
			},
		},
		"legacy GAR impersonation": {
			config: &config.ExternalImageComponentConfig{GCPGARImageConfig: &config.GCPGARConfig{ServiceAccountEmail: "acme@example.com", GCPRegion: "us-central1", GCPProjectID: "acme", ImageURL: "us-central1-docker.pkg.dev/acme/repo/image", Tag: "v1"}},
			check: func(t *testing.T, cfg *app.ExternalImageComponentConfig) {
				require.Equal(t, "acme@example.com", cfg.GCPGARImageConfig.ServiceAccountEmail)
			},
		},
		"legacy GAR federation": {
			config: &config.ExternalImageComponentConfig{GCPGARImageConfig: &config.GCPGARConfig{WorkloadIdentityProvider: "projects/123/locations/global/workloadIdentityPools/acme/providers/acme", GCPRegion: "us-central1", GCPProjectID: "acme", ImageURL: "us-central1-docker.pkg.dev/acme/repo/image", Tag: "v1"}},
			check: func(t *testing.T, cfg *app.ExternalImageComponentConfig) {
				require.NotEmpty(t, cfg.GCPGARImageConfig.WorkloadIdentityProvider)
			},
		},
		"ambient ACR": {
			config: &config.ExternalImageComponentConfig{AzureACRImageConfig: &config.AzureACRConfig{RegistryURL: "acme.azurecr.io", ImageURL: "acme.azurecr.io/repo", Tag: "v1"}},
			check: func(t *testing.T, cfg *app.ExternalImageComponentConfig) {
				require.False(t, cfg.AzureACRImageConfig.UsesCloudConnection())
			},
		},
		"secret ACR": {
			config: &config.ExternalImageComponentConfig{AzureACRImageConfig: &config.AzureACRConfig{RegistryURL: "acme.azurecr.io", ImageURL: "acme.azurecr.io/repo", Tag: "v1", TenantID: "tenant", ClientID: "client", ClientSecretName: "acme-secret"}},
			check: func(t *testing.T, cfg *app.ExternalImageComponentConfig) {
				require.Equal(t, "acme-secret", cfg.AzureACRImageConfig.ClientSecretName)
				require.False(t, cfg.AzureACRImageConfig.UsesCloudConnection())
			},
		},
		"certificate ACR": {
			config: &config.ExternalImageComponentConfig{AzureACRImageConfig: &config.AzureACRConfig{RegistryURL: "acme.azurecr.io", ImageURL: "acme.azurecr.io/repo", Tag: "v1", TenantID: "tenant", ClientID: "client", ClientCertificateName: "acme-cert"}},
			check: func(t *testing.T, cfg *app.ExternalImageComponentConfig) {
				require.Equal(t, "acme-cert", cfg.AzureACRImageConfig.ClientCertificateName)
			},
		},
		"secret and certificate ACR": {config: &config.ExternalImageComponentConfig{AzureACRImageConfig: &config.AzureACRConfig{RegistryURL: "acme.azurecr.io", ImageURL: "acme.azurecr.io/repo", Tag: "v1", TenantID: "tenant", ClientID: "client", ClientSecretName: "secret", ClientCertificateName: "cert"}}, wantError: true},
		"incomplete ACR":             {config: &config.ExternalImageComponentConfig{AzureACRImageConfig: &config.AzureACRConfig{ClientID: "client"}}, wantError: true},
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
		})
	}
}

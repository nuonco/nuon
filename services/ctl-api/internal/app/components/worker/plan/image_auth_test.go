package plan

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestImageSourcePlanAuthentication(t *testing.T) {
	for name, tc := range map[string]struct {
		acr                      app.AzureACRImageConfig
		connection, registration bool
	}{
		"saved ambient":     {acr: app.AzureACRImageConfig{ID: "saved"}},
		"saved pin":         {acr: app.AzureACRImageConfig{ID: "saved", CloudConnectionID: "cc_saved"}, connection: true},
		"saved secret":      {acr: app.AzureACRImageConfig{ID: "saved", TenantID: "tenant", ClientID: "client", ClientSecretName: "secret"}, registration: true},
		"saved certificate": {acr: app.AzureACRImageConfig{ID: "saved", TenantID: "tenant", ClientID: "client", ClientCertificateName: "cert"}, registration: true},
	} {
		t.Run(name, func(t *testing.T) {
			tc.acr.RegistryURL = "acme.azurecr.io"
			repo, err := (&Planner{}).getSourceRepository(&app.ExternalImageComponentConfig{ImageURL: "acme.azurecr.io/repo", AzureACRImageConfig: &tc.acr}, "component")
			require.NoError(t, err)
			if tc.connection {
				require.Nil(t, repo.ACRAuth)
			} else {
				require.True(t, repo.ACRAuth.UseDefault)
			}
			require.Equal(t, tc.registration, repo.ACRAppRegistration != nil)
		})
	}
	for name, tc := range map[string]struct{ account, provider string }{
		"GAR account":    {account: "acme@example.com"},
		"GAR federation": {account: "acme@example.com", provider: "provider"},
	} {
		t.Run(name, func(t *testing.T) {
			repo, err := (&Planner{}).getSourceRepository(&app.ExternalImageComponentConfig{ImageURL: "us-central1-docker.pkg.dev/acme/repo/image", GCPGARImageConfig: &app.GCPGARImageConfig{ID: "saved", GCPRegion: "us-central1", ServiceAccountEmail: tc.account, WorkloadIdentityProvider: tc.provider}}, "component")
			require.NoError(t, err)
			require.Equal(t, tc.account, repo.ServiceAccountEmail)
			require.Equal(t, tc.provider, repo.WorkloadIdentityProvider)
		})
	}
}

package service

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/oidcissuer"
)

func TestAzureSetup(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	issuer, err := oidcissuer.New("https://api.example.com", key, "example-key")
	require.NoError(t, err)
	svc := &service{issuer: issuer}

	tests := map[string]struct {
		capabilities []app.CloudConnectionCapability
		repositories []string
		want         []string
		notWant      []string
	}{
		"stacks uses Owner rather than split roles": {
			capabilities: []app.CloudConnectionCapability{app.CloudConnectionCapabilityStacks},
			want:         []string{"--role Owner", "nuonco/acr-access/azure"},
			notWant:      []string{"Contributor", "User Access Administrator", "AcrPull"},
		},
		"images without repositories": {
			capabilities: []app.CloudConnectionCapability{app.CloudConnectionCapabilityImages},
			want:         []string{"--role AcrPull", "repositories    = []"},
			notWant:      []string{"--role Owner"},
		},
		"images with repositories": {
			capabilities: []app.CloudConnectionCapability{app.CloudConnectionCapabilityImages},
			repositories: []string{"backend", "worker"},
			want:         []string{"--role AcrPull", `repositories    = ["backend","worker"]`},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			connection := &app.CloudConnection{ID: "cc_example", OrgID: "org_example", Platform: app.CloudPlatformAzure, TargetID: "00000000-0000-0000-0000-000000000001", Capabilities: test.capabilities}
			got := svc.setup(connection, SetupOptions{Registry: "acme", Repositories: test.repositories})
			for _, value := range test.want {
				assert.Contains(t, got.CLI+got.Terraform, value)
			}
			for _, value := range test.notWant {
				assert.NotContains(t, got.CLI+got.Terraform, value)
			}
			assert.Contains(t, got.PortalJSON, azureTokenExchangeAudience)
		})
	}
}

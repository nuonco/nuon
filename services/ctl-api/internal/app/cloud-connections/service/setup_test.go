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

func TestGCPSetup(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	issuer, err := oidcissuer.New("https://api.example.com", key, "example-key")
	require.NoError(t, err)
	svc := &service{issuer: issuer}

	tests := map[string]struct {
		identityProvider string
		repositories     []string
		want             []string
		notWant          []string
	}{
		"project binding with placeholder provider": {
			want:    []string{"projects/<project-number>/locations/global/workloadIdentityPools/<pool-id>/providers/<provider-id>", "gcloud projects add-iam-policy-binding", "nuonco/gar-access/google"},
			notWant: []string{"gcloud artifacts repositories add-iam-policy-binding"},
		},
		"repository bindings with supplied provider": {
			identityProvider: "projects/123456789/locations/global/workloadIdentityPools/nuon/providers/connection",
			repositories:     []string{"us-central1/backend", "europe-west1/worker"},
			want:             []string{"https://iam.googleapis.com/projects/123456789/locations/global/workloadIdentityPools/nuon/providers/connection", `repositories = ["us-central1/backend","europe-west1/worker"]`, `repositories add-iam-policy-binding "backend"`, `--location "us-central1"`, `repositories add-iam-policy-binding "worker"`, `--location "europe-west1"`},
			notWant:          []string{"gcloud projects add-iam-policy-binding"},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			connection := &app.CloudConnection{ID: "cc_example", OrgID: "org_example", Platform: app.CloudPlatformGCP, TargetID: "acme-project", Principal: "nuon-cloud-connection@acme-project.iam.gserviceaccount.com", IdentityProvider: test.identityProvider, Capabilities: []app.CloudConnectionCapability{app.CloudConnectionCapabilityImages}}
			got := svc.setup(connection, SetupOptions{Repositories: test.repositories})
			for _, value := range test.want {
				assert.Contains(t, got.CLI+got.Terraform+got.Audience, value)
			}
			for _, value := range test.notWant {
				assert.NotContains(t, got.CLI+got.Terraform, value)
			}
		})
	}
}

package activities

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestGetSourceRepositoryGARImpersonation(t *testing.T) {
	const (
		ownerOrg = "orgowner0000000000000000000"
		otherOrg = "orgother0000000000000000000"
		vendorSA = "gar-reader@acme-artifacts.iam.gserviceaccount.com"
	)

	a := &Activities{cfg: &internal.Config{
		GCPGARImpersonationAllowlist: []string{ownerOrg + ":" + vendorSA},
	}}
	cfg := func(sa string) *app.ExternalImageComponentConfig {
		return &app.ExternalImageComponentConfig{
			ImageURL: "europe-docker.pkg.dev/acme-artifacts/runtime/sandbox",
			GCPGARImageConfig: &app.GCPGARImageConfig{
				GCPProjectID:        "acme-artifacts",
				GCPRegion:           "europe",
				ServiceAccountEmail: sa,
			},
		}
	}

	repo, err := a.getSourceRepository(cfg(vendorSA), "cmp1", ownerOrg)
	require.NoError(t, err)
	require.Equal(t, vendorSA, repo.ServiceAccountEmail)

	repo, err = a.getSourceRepository(cfg(""), "cmp1", otherOrg)
	require.NoError(t, err)
	require.Empty(t, repo.ServiceAccountEmail)

	_, err = a.getSourceRepository(cfg(vendorSA), "cmp1", otherOrg)
	require.ErrorContains(t, err, "not allowlisted for org "+otherOrg)
}

package activities

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal"
)

func TestGetOrgECRAccessInfoRefusesControlPlaneGARToken(t *testing.T) {
	info, err := (&Activities{cfg: &internal.Config{CloudProvider: "gcp"}}).GetOrgECRAccessInfo(context.Background(), &GetOrgECRAccessInfoRequest{
		OrgID: "org",
		AppID: "app",
	})
	require.Nil(t, info)
	require.ErrorContains(t, err, "control plane token")
}

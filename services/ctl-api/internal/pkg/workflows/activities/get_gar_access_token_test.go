package activities

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetGARAccessTokenRefusesControlPlaneIdentity(t *testing.T) {
	token, err := (&Activities{}).GetGARAccessToken(context.Background(), &GetGARAccessTokenRequest{})
	require.Nil(t, token)
	require.ErrorContains(t, err, "control plane token")
}

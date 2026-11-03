package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"

	"github.com/nuonco/nuon/services/ctl-api/internal"
)

type countingTokenSource struct {
	calls  *int
	expiry time.Duration
	err    error
}

func (c countingTokenSource) Token() (*oauth2.Token, error) {
	*c.calls++
	if c.err != nil {
		return nil, c.err
	}
	return &oauth2.Token{AccessToken: "token", Expiry: time.Now().Add(c.expiry)}, nil
}

func testIssuer(calls *int, expiry time.Duration, err error) *registryAuthIssuer {
	return &registryAuthIssuer{
		serviceAccount: "reader@example.iam.gserviceaccount.com",
		newTokenSource: func(context.Context, string) (oauth2.TokenSource, error) {
			return countingTokenSource{calls: calls, expiry: expiry, err: err}, nil
		},
	}
}

func TestRegistryAuthIssuerDisabledWithoutServiceAccount(t *testing.T) {
	require.Nil(t, newRegistryAuthIssuer(nil))
	issuer := newRegistryAuthIssuer(&internal.Config{})
	require.Nil(t, issuer)

	auth, err := issuer.auth("europe-west4-docker.pkg.dev/acme/acme-runner/runner")
	require.NoError(t, err)
	require.Nil(t, auth)
}

func TestRegistryAuthIssuerSkipsNonArtifactRegistryImages(t *testing.T) {
	calls := 0
	issuer := testIssuer(&calls, time.Hour, nil)

	auth, err := issuer.auth("public.ecr.aws/p7e3r5y0/runner")
	require.NoError(t, err)
	require.Nil(t, auth)
	require.Zero(t, calls)
}

func TestRegistryAuthIssuerReturnsArtifactRegistryAuth(t *testing.T) {
	calls := 0
	issuer := testIssuer(&calls, time.Hour, nil)

	auth, err := issuer.auth("europe-west4-docker.pkg.dev/acme/acme-runner/runner")
	require.NoError(t, err)
	require.Equal(t, "europe-west4-docker.pkg.dev", auth.Registry)
	require.Equal(t, registryAuthUsername, auth.Username)
	require.Equal(t, "token", auth.Password)
	require.WithinDuration(t, time.Now().Add(time.Hour), auth.ExpiresAt, time.Minute)
}

func TestRegistryAuthIssuerCachesFreshTokens(t *testing.T) {
	calls := 0
	issuer := testIssuer(&calls, time.Hour, nil)

	for range 3 {
		_, err := issuer.auth("europe-west4-docker.pkg.dev/acme/acme-runner/runner")
		require.NoError(t, err)
	}
	require.Equal(t, 1, calls)
}

func TestRegistryAuthIssuerRemintsNearExpiry(t *testing.T) {
	calls := 0
	issuer := testIssuer(&calls, registryAuthMinRemaining-time.Minute, nil)

	for range 2 {
		_, err := issuer.auth("europe-west4-docker.pkg.dev/acme/acme-runner/runner")
		require.NoError(t, err)
	}
	require.Equal(t, 2, calls)
}

func TestRegistryAuthIssuerReturnsMintErrors(t *testing.T) {
	calls := 0
	issuer := testIssuer(&calls, time.Hour, errors.New("permission denied"))

	auth, err := issuer.auth("europe-west4-docker.pkg.dev/acme/acme-runner/runner")
	require.ErrorContains(t, err, "permission denied")
	require.Nil(t, auth)
}

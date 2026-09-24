package credentials

import (
	"context"
	"fmt"

	"google.golang.org/api/impersonate"
)

// ImpersonatedAccessToken mints an OAuth2 access token for targetSA using the
// caller's own credentials as the source.
func ImpersonatedAccessToken(ctx context.Context, targetSA string) (string, error) {
	ts, err := impersonate.CredentialsTokenSource(ctx, impersonate.CredentialsConfig{
		TargetPrincipal: targetSA,
		Scopes:          []string{"https://www.googleapis.com/auth/cloud-platform"},
	})
	if err != nil {
		return "", fmt.Errorf("unable to impersonate %s: %w", targetSA, err)
	}

	token, err := ts.Token()
	if err != nil {
		return "", fmt.Errorf("unable to get an access token for %s: %w", targetSA, err)
	}

	return token.AccessToken, nil
}

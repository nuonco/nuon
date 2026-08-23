// Package auth resolves the credential a Nuon SDK presents on a request: an
// explicit token, then NUON_API_TOKEN, then an ambient OIDC token exchanged for
// a short-lived one.
//
// The precedence lives here so it cannot drift between SDKs. The exchange does
// not: sdks/stack POSTs to the runner API directly while nuon-go goes through
// generated operations, which is what Exchanger abstracts.
package auth

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/nuonco/nuon/sdks/auth/oidctoken"
)

// APITokenEnvVar authenticates non-interactively, without a config file.
const APITokenEnvVar = "NUON_API_TOKEN"

// OrgIDEnvVar names the org whose trust policies the OIDC exchange evaluates.
const OrgIDEnvVar = "NUON_ORG_ID"

// Exchanger trades an OIDC ID token for a short-lived Nuon API token over the
// SDK's own transport. Called only with a non-empty orgID and jwt.
type Exchanger interface {
	ExchangeOIDCToken(ctx context.Context, orgID, jwt string) (string, error)
}

// Options are the inputs to Resolve; the environment supplies what is unset.
type Options struct {
	// APIToken wins over every other source.
	APIToken string

	// OrgID is required on the OIDC path only. Falls back to NUON_ORG_ID.
	OrgID string

	// Audience to request for an ambient OIDC token. The control plane compares
	// it literally against the trust policy, so there is no default — only the
	// caller knows which host it authenticates against. NUON_OIDC_AUDIENCE wins.
	Audience string
}

// Resolve produces the bearer token for a request. ex is consulted only on the
// OIDC path, so nil is valid for a caller that supports static tokens alone.
func Resolve(ctx context.Context, opts Options, ex Exchanger) (string, error) {
	if t := strings.TrimSpace(opts.APIToken); t != "" {
		return t, nil
	}
	if t := strings.TrimSpace(os.Getenv(APITokenEnvVar)); t != "" {
		return t, nil
	}

	if !oidctoken.Available() {
		return "", fmt.Errorf(
			"no credentials: set an api token, %s, or run somewhere an OIDC token is available "+
				"(GitHub Actions with `permissions: id-token: write`, %s, or %s)",
			APITokenEnvVar, "NUON_OIDC_TOKEN", "NUON_OIDC_TOKEN_FILE",
		)
	}

	if ex == nil {
		return "", fmt.Errorf("an OIDC token is available but this client cannot exchange one")
	}

	orgID := strings.TrimSpace(opts.OrgID)
	if orgID == "" {
		orgID = strings.TrimSpace(os.Getenv(OrgIDEnvVar))
	}
	if orgID == "" {
		// Checked before fetching a token: the exchange cannot succeed without an org.
		return "", fmt.Errorf("an OIDC token is available but no org id is set: set an org id or %s", OrgIDEnvVar)
	}

	jwt, source, ok, err := oidctoken.Detect(ctx, oidctoken.Audience("", opts.Audience))
	if err != nil {
		return "", fmt.Errorf("unable to get OIDC token from %s: %w", source, err)
	}
	if !ok {
		return "", fmt.Errorf("no OIDC token found")
	}

	token, err := ex.ExchangeOIDCToken(ctx, orgID, jwt)
	if err != nil {
		return "", fmt.Errorf("exchange OIDC token (from %s): %w", source, err)
	}

	return token, nil
}

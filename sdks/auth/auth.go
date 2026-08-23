// Package auth resolves the credential a Nuon SDK presents on a request.
//
// The precedence — an explicit token, then NUON_API_TOKEN, then an ambient OIDC
// token exchanged for a short-lived Nuon token — is the part that must not differ
// between SDKs. A customer who sets NUON_API_TOKEN and gets one behavior from the
// CLI and another from the stack SDK has found a bug, so the order and the error
// messages live here rather than in each SDK.
//
// The exchange itself does not, because the two SDKs reach the same endpoint by
// different routes: sdks/stack POSTs to the runner API directly, while nuon-go
// calls it through generated go-swagger operations. That difference is the
// Exchanger interface.
package auth

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/nuonco/nuon/sdks/auth/oidctoken"
)

// APITokenEnvVar is the usual way to authenticate non-interactively without
// writing a token into a config file.
const APITokenEnvVar = "NUON_API_TOKEN"

// OrgIDEnvVar is consulted only on the OIDC path, where the exchange has to name
// the org whose trust policies should be evaluated.
const OrgIDEnvVar = "NUON_ORG_ID"

// Exchanger trades an OIDC ID token for a short-lived Nuon API token.
//
// Each SDK implements this over its own transport. Implementations are called
// only when Resolve reaches the OIDC path, and only with a non-empty orgID and
// jwt, so they need not re-check either.
type Exchanger interface {
	ExchangeOIDCToken(ctx context.Context, orgID, jwt string) (string, error)
}

// Options are the inputs to Resolve. Every field is optional except as noted;
// the environment supplies the rest.
type Options struct {
	// APIToken authenticates the caller directly and wins over everything else.
	APIToken string

	// OrgID is required on the OIDC path only. Falls back to NUON_ORG_ID.
	OrgID string

	// Audience is the audience to request for an ambient OIDC token. It has to
	// equal the audience recorded on the trust policy being matched, so there is
	// deliberately no default: only the caller knows which host it authenticates
	// against. NUON_OIDC_AUDIENCE overrides it.
	//
	// Getting this wrong is not a redirect or a retry — the control plane compares
	// it literally and the exchange fails.
	Audience string
}

// Resolve produces the bearer token for a request.
//
// ex is consulted only if there is no static token to use; passing nil is
// therefore valid for a caller that supports static tokens alone.
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
		// Checked before fetching the ID token rather than after: the exchange
		// cannot succeed without an org, and failing here says why.
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

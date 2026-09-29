package auth

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/nuonco/nuon/sdks/auth/oidctoken"
)

const APITokenEnvVar = "NUON_API_TOKEN"

const OrgIDEnvVar = "NUON_ORG_ID"

type Exchanger interface {
	ExchangeOIDCToken(ctx context.Context, orgID, jwt string) (string, error)
}

type Options struct {
	APIToken string

	OrgID string

	Audience string
}

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
				"(GitHub Actions with `permissions: id-token: write`, HCP Terraform workload identity, %s, or %s)",
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

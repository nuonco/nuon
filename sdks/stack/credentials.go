package stack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/nuonco/nuon/sdks/stack/oidctoken"
)

// apiTokenEnvVar mirrors the CLI: a token in the environment is the usual way to
// run non-interactively without writing it into a config file.
const apiTokenEnvVar = "NUON_API_TOKEN"

// orgIDEnvVar is only consulted on the OIDC path, where the exchange has to name the
// org whose trust policies should be evaluated.
const orgIDEnvVar = "NUON_ORG_ID"

// resolveToken produces the bearer token for a request, following the same
// precedence the CLI uses: an explicit token, then the environment, then an ambient
// OIDC token exchanged for a short-lived Nuon token.
//
// The OIDC path exists so a customer applying the module from CI never has to store
// a long-lived credential: GitHub Actions mints an ID token per run, and the control
// plane trades it for a Nuon token if it satisfies one of the org's trust policies.
func resolveToken(ctx context.Context, opts Options) (string, error) {
	if t := strings.TrimSpace(opts.APIToken); t != "" {
		return t, nil
	}
	if t := strings.TrimSpace(os.Getenv(apiTokenEnvVar)); t != "" {
		return t, nil
	}

	if !oidctoken.Available() {
		return "", fmt.Errorf(
			"no credentials: set api_token, %s, or run somewhere an OIDC token is available "+
				"(GitHub Actions with `permissions: id-token: write`, %s, or %s)",
			apiTokenEnvVar, "NUON_OIDC_TOKEN", "NUON_OIDC_TOKEN_FILE",
		)
	}

	orgID := strings.TrimSpace(opts.OrgID)
	if orgID == "" {
		orgID = strings.TrimSpace(os.Getenv(orgIDEnvVar))
	}
	if orgID == "" {
		// Checked before fetching the ID token rather than after: the exchange
		// cannot succeed without an org, and failing here says why.
		return "", fmt.Errorf("an OIDC token is available but no org id is set: set org_id or %s", orgIDEnvVar)
	}

	jwt, source, ok, err := oidctoken.Detect(ctx, oidctoken.Audience("", opts.APIURL))
	if err != nil {
		return "", fmt.Errorf("unable to get OIDC token from %s: %w", source, err)
	}
	if !ok {
		return "", fmt.Errorf("no OIDC token found")
	}

	token, err := exchangeOIDCToken(ctx, opts, orgID, jwt)
	if err != nil {
		return "", fmt.Errorf("exchange OIDC token (from %s): %w", source, err)
	}

	return token, nil
}

type exchangeRequest struct {
	OrgID string `json:"org_id"`
	Token string `json:"token"`
}

type exchangeResponse struct {
	Authenticated bool   `json:"authenticated"`
	Token         string `json:"token,omitempty"`
}

// exchangeOIDCToken trades an OIDC ID token for a short-lived Nuon API token.
//
// Deliberately not routed through runClient: that client attaches a bearer token,
// and this is the call made precisely because there isn't one yet.
func exchangeOIDCToken(ctx context.Context, opts Options, orgID, jwt string) (string, error) {
	body, err := json.Marshal(exchangeRequest{OrgID: orgID, Token: jwt})
	if err != nil {
		return "", fmt.Errorf("marshal exchange request: %w", err)
	}

	url := fmt.Sprintf("%s/v1/oidc/token", strings.TrimSuffix(opts.APIURL, "/"))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("build exchange request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	hc := opts.HTTPClient
	if hc == nil {
		hc = &http.Client{Timeout: 10 * time.Second}
	}

	resp, err := hc.Do(req)
	if err != nil {
		return "", fmt.Errorf("call token exchange: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("token exchange returned %d: %s", resp.StatusCode, string(respBody))
	}

	var out exchangeResponse
	if err := json.Unmarshal(respBody, &out); err != nil {
		return "", fmt.Errorf("decode exchange response: %w", err)
	}
	if !out.Authenticated || out.Token == "" {
		// The control plane returns a deliberately uniform error for every auth
		// failure, so there is nothing more specific to report here.
		return "", fmt.Errorf("token exchange did not authenticate: check the org's OIDC trust policies")
	}

	return out.Token, nil
}

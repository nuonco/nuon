package stack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/nuonco/nuon/sdks/auth"
)

// resolveToken produces the bearer token for a request.
//
// The precedence and its error messages live in sdks/auth so every Nuon SDK
// behaves the same way; this only supplies the exchange, which is transport-
// specific.
//
// The audience passed through is opts.APIURL — the runner API. It has to equal the
// audience recorded on the trust policy, and a policy created for this SDK names the
// runner API rather than the public one the CLI talks to. NUON_OIDC_AUDIENCE
// overrides it; the dashboard's OIDC directions print the policy's audience for
// exactly this reason.
func resolveToken(ctx context.Context, opts Options) (string, error) {
	return auth.Resolve(ctx, auth.Options{
		APIToken: opts.APIToken,
		OrgID:    opts.OrgID,
		Audience: opts.APIURL,
	}, exchanger{opts: opts})
}

// exchanger implements auth.Exchanger against the runner API.
type exchanger struct {
	opts Options
}

type exchangeRequest struct {
	OrgID string `json:"org_id"`
	Token string `json:"token"`
}

type exchangeResponse struct {
	Authenticated bool   `json:"authenticated"`
	Token         string `json:"token,omitempty"`
}

// ExchangeOIDCToken trades an OIDC ID token for a short-lived Nuon API token.
//
// Deliberately not routed through runClient: that client attaches a bearer token,
// and this is the call made precisely because there isn't one yet.
func (e exchanger) ExchangeOIDCToken(ctx context.Context, orgID, jwt string) (string, error) {
	opts := e.opts

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

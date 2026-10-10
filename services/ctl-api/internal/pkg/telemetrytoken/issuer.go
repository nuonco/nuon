package telemetrytoken

import (
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/oidcissuer"
)

const (
	LegacyAudience = "urn:nuon:telemetry"
	Scope          = "telemetry:write"
	Lifetime       = 10 * time.Minute
	maxJWKSSize    = 64 * 1024
)

type Principal struct {
	OrgID         string
	AppID         string
	InstallID     string
	AccountID     string
	RelayEndpoint string
}

type AccessTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
}

type Issuer struct {
	signer *oidcissuer.Issuer
}

func New(cfg *internal.Config) (*Issuer, error) {
	if cfg == nil || cfg.TelemetryJWKS == "" {
		return nil, nil
	}
	if len(cfg.TelemetryJWKS) > maxJWKSSize {
		return nil, fmt.Errorf("telemetry JWKS exceeds maximum size")
	}

	issuer := strings.TrimRight(cfg.TelemetryJWTIssuer, "/")
	if issuer == "" {
		issuer = strings.TrimRight(cfg.PublicAPIURL, "/")
	}
	signer, _, err := oidcissuer.NewFromJWKS(issuer, cfg.TelemetryJWKS)
	if err != nil {
		return nil, err
	}

	return &Issuer{signer: signer}, nil
}

func (i *Issuer) Issue(principal Principal) (string, error) {
	claims, err := principal.claims()
	if err != nil {
		return "", err
	}
	return i.signer.MintAccessToken(principal.AccountID, principal.RelayEndpoint, Lifetime, claims)
}

func (i *Issuer) IssueLegacyRunner(principal Principal, runnerID string, endpointBound bool) (string, error) {
	if runnerID == "" {
		return "", fmt.Errorf("legacy telemetry token requires a runner")
	}
	claims, err := principal.claims()
	if err != nil {
		return "", err
	}
	audience := LegacyAudience
	if endpointBound {
		audience = principal.RelayEndpoint
	}
	claims["client_id"] = runnerID
	claims["nuon_runner_id"] = runnerID
	subject := fmt.Sprintf("org:%s:install:%s:runner:%s", principal.OrgID, principal.InstallID, runnerID)
	return i.signer.MintAccessToken(subject, audience, Lifetime, claims)
}

func (p Principal) claims() (jwt.MapClaims, error) {
	if p.OrgID == "" || p.AppID == "" || p.InstallID == "" || p.AccountID == "" || p.RelayEndpoint == "" {
		return nil, fmt.Errorf("telemetry principal requires an org, app, install, account and relay endpoint")
	}
	return jwt.MapClaims{
		"client_id":       p.AccountID,
		"scope":           Scope,
		"nuon_org_id":     p.OrgID,
		"nuon_app_id":     p.AppID,
		"nuon_install_id": p.InstallID,
	}, nil
}

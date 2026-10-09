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
	RunnerID      string
	CollectorID   string
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

func (i *Issuer) Issue(principal Principal, endpointBound bool) (string, error) {
	if principal.OrgID == "" || principal.AppID == "" || principal.InstallID == "" || principal.RelayEndpoint == "" || (principal.RunnerID == "") == (principal.CollectorID == "") {
		return "", fmt.Errorf("telemetry principal requires an org, app, install, relay endpoint and exactly one runner or collector")
	}
	if principal.CollectorID != "" && !endpointBound {
		return "", fmt.Errorf("collector telemetry tokens must be relay-bound")
	}

	audience := LegacyAudience
	if endpointBound {
		audience = principal.RelayEndpoint
	}
	kind, clientID := "runner", principal.RunnerID
	if principal.CollectorID != "" {
		kind, clientID = "collector", principal.CollectorID
	}
	claims := jwt.MapClaims{
		"client_id":       clientID,
		"scope":           Scope,
		"nuon_org_id":     principal.OrgID,
		"nuon_app_id":     principal.AppID,
		"nuon_install_id": principal.InstallID,
	}
	if principal.RunnerID != "" {
		claims["nuon_runner_id"] = principal.RunnerID
	}
	if principal.CollectorID != "" {
		claims["nuon_collector_id"] = principal.CollectorID
	}
	subject := fmt.Sprintf("org:%s:install:%s:%s:%s", principal.OrgID, principal.InstallID, kind, clientID)
	return i.signer.MintAccessToken(subject, audience, Lifetime, claims)
}

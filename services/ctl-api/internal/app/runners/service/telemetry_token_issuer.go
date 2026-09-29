package service

import (
	"crypto/rsa"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/oidcissuer"
)

const (
	telemetryTokenAudience = "urn:nuon:telemetry"
	telemetryTokenScope    = "telemetry:write"
	telemetryTokenLifetime = 10 * time.Minute
	maxTelemetryJWKSSize   = 64 * 1024
)

type telemetryAccessTokenClaims struct {
	ClientID  string `json:"client_id"`
	Scope     string `json:"scope"`
	OrgID     string `json:"nuon_org_id"`
	AppID     string `json:"nuon_app_id"`
	InstallID string `json:"nuon_install_id"`
	RunnerID  string `json:"nuon_runner_id"`
	jwt.RegisteredClaims
}

type telemetryTokenIssuer struct {
	issuer     string
	keyID      string
	privateKey *rsa.PrivateKey
	now        func() time.Time
}

func newTelemetryTokenIssuer(cfg *internal.Config) (*telemetryTokenIssuer, error) {
	if cfg == nil || cfg.TelemetryJWKS == "" {
		return nil, nil
	}
	if len(cfg.TelemetryJWKS) > maxTelemetryJWKSSize {
		return nil, fmt.Errorf("telemetry JWKS exceeds maximum size")
	}

	issuer := strings.TrimRight(cfg.TelemetryJWTIssuer, "/")
	if issuer == "" {
		issuer = strings.TrimRight(cfg.PublicAPIURL, "/")
	}
	if err := oidcissuer.ValidateIssuer(issuer); err != nil {
		return nil, err
	}

	privateKey, keyID, _, err := oidcissuer.ParseJWKS(cfg.TelemetryJWKS)
	if err != nil {
		return nil, err
	}

	return &telemetryTokenIssuer{
		issuer:     issuer,
		keyID:      keyID,
		privateKey: privateKey,
		now:        time.Now,
	}, nil
}

func (i *telemetryTokenIssuer) issue(principal telemetryRunnerPrincipal) (string, error) {
	if principal.OrgID == "" || principal.AppID == "" || principal.InstallID == "" || principal.RunnerID == "" {
		return "", fmt.Errorf("telemetry runner principal is incomplete")
	}

	now := i.now().UTC()
	claims := telemetryAccessTokenClaims{
		ClientID:  principal.RunnerID,
		Scope:     telemetryTokenScope,
		OrgID:     principal.OrgID,
		AppID:     principal.AppID,
		InstallID: principal.InstallID,
		RunnerID:  principal.RunnerID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    i.issuer,
			Subject:   fmt.Sprintf("org:%s:install:%s:runner:%s", principal.OrgID, principal.InstallID, principal.RunnerID),
			Audience:  jwt.ClaimStrings{telemetryTokenAudience},
			ExpiresAt: jwt.NewNumericDate(now.Add(telemetryTokenLifetime)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        uuid.NewString(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = i.keyID
	token.Header["typ"] = "at+jwt"

	signed, err := token.SignedString(i.privateKey)
	if err != nil {
		return "", fmt.Errorf("sign telemetry access token: %w", err)
	}
	return signed, nil
}

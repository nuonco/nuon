package cloudconnections

import (
	"fmt"

	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/oidcissuer"
)

func NewVerifierFromConfig(cfg *internal.Config, l *zap.Logger) (Verifier, error) {
	issuer, _, err := IssuerFromConfig(cfg)
	if err != nil {
		return nil, err
	}
	return NewVerifier(issuer, l), nil
}

func IssuerFromConfig(cfg *internal.Config) (*oidcissuer.Issuer, oidcissuer.JWKS, error) {
	if cfg == nil || cfg.TelemetryJWKS == "" {
		return nil, oidcissuer.JWKS{}, nil
	}
	issuer, publicKeys, err := oidcissuer.NewFromJWKS(cfg.PublicAPIURL, cfg.TelemetryJWKS)
	if err != nil {
		return nil, oidcissuer.JWKS{}, fmt.Errorf("initialize cloud connection issuer: %w", err)
	}
	return issuer, publicKeys, nil
}

const ForeignSubject = "org:foreign:connection:foreign"

func Subject(connection *app.CloudConnection) string {
	return fmt.Sprintf("org:%s:connection:%s", connection.OrgID, connection.ID)
}

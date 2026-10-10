package cloudconnections

import (
	"fmt"

	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/oidcissuer"
)

func NewVerifierFromConfig(cfg *internal.Config, l *zap.Logger) (Verifier, error) {
	issuer, err := IssuerFromConfig(cfg)
	if err != nil {
		return nil, err
	}
	return NewAWSVerifier(issuer, l), nil
}

func IssuerFromConfig(cfg *internal.Config) (*oidcissuer.Issuer, error) {
	if cfg == nil || cfg.TelemetryJWKS == "" {
		return nil, nil
	}
	issuer, _, err := oidcissuer.NewFromJWKS(cfg.PublicAPIURL, cfg.TelemetryJWKS)
	if err != nil {
		return nil, fmt.Errorf("initialize cloud connection issuer: %w", err)
	}
	return issuer, nil
}

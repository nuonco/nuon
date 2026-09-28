package cloudconnections

import (
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/oidcissuer"
)

func NewVerifierFromConfig(cfg *internal.Config) (Verifier, error) {
	issuer, err := IssuerFromConfig(cfg)
	if err != nil {
		return nil, err
	}
	return NewAWSVerifier(issuer), nil
}

func IssuerFromConfig(cfg *internal.Config) (*oidcissuer.Issuer, error) {
	if cfg == nil || cfg.TelemetryJWKS == "" {
		return nil, nil
	}
	privateKey, keyID, _, err := oidcissuer.ParseJWKS(cfg.TelemetryJWKS)
	if err != nil {
		return nil, fmt.Errorf("initialize cloud connection issuer: %w", err)
	}
	issuer, err := oidcissuer.New(cfg.PublicAPIURL, privateKey, keyID)
	if err != nil {
		return nil, fmt.Errorf("initialize cloud connection issuer: %w", err)
	}
	return issuer, nil
}

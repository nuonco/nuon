package config

import (
	"testing"

	"github.com/pelletier/go-toml/v2"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/oci/signature"
)

func TestExternalImageSignatureVerificationDecode(t *testing.T) {
	var cfg ExternalImageComponentConfig
	require.NoError(t, toml.Unmarshal([]byte(`
[public]
image_url = "cgr.dev/acme/image"
tag = "latest"

[verification]
require_signature = true

[[verification.authorities]]
type = "keyless"
issuer = "https://issuer.example.com"
subject_regexp = "^https://example.com/.+$"

[[verification.authorities]]
type = "public_key"
public_key = "./cosign.pub"
`), &cfg))

	require.NotNil(t, cfg.Verification)
	require.True(t, cfg.Verification.RequireSignature)
	require.Equal(t, []signature.Authority{
		{
			Type:          signature.AuthorityTypeKeyless,
			Issuer:        "https://issuer.example.com",
			SubjectRegexp: "^https://example.com/.+$",
		},
		{
			Type:      signature.AuthorityTypePublicKey,
			PublicKey: "./cosign.pub",
		},
	}, cfg.Verification.Authorities)
	require.NoError(t, cfg.Validate())
}

func TestAzureACRConfigValidateCredentials(t *testing.T) {
	const (
		tenant = "00000000-0000-0000-0000-000000000000"
		client = "11111111-1111-1111-1111-111111111111"
		secret = "acr-secret"
		cert   = "acr-cert"
	)

	tests := []struct {
		name    string
		cfg     AzureACRConfig
		wantErr bool
	}{
		{
			name: "nothing set uses ambient credentials",
			cfg:  AzureACRConfig{},
		},
		{
			name: "complete with secret",
			cfg:  AzureACRConfig{TenantID: tenant, ClientID: client, ClientSecretName: secret},
		},
		{
			name: "complete with certificate",
			cfg:  AzureACRConfig{TenantID: tenant, ClientID: client, ClientCertificateName: cert},
		},
		{
			name:    "both credential kinds",
			cfg:     AzureACRConfig{TenantID: tenant, ClientID: client, ClientSecretName: secret, ClientCertificateName: cert},
			wantErr: true,
		},
		{
			name:    "secret without tenant or client",
			cfg:     AzureACRConfig{ClientSecretName: secret},
			wantErr: true,
		},
		{
			name:    "certificate without tenant or client",
			cfg:     AzureACRConfig{ClientCertificateName: cert},
			wantErr: true,
		},
		{
			name:    "tenant and client without a credential",
			cfg:     AzureACRConfig{TenantID: tenant, ClientID: client},
			wantErr: true,
		},
		{
			name:    "missing tenant only",
			cfg:     AzureACRConfig{ClientID: client, ClientSecretName: secret},
			wantErr: true,
		},
		{
			name:    "missing client only",
			cfg:     AzureACRConfig{TenantID: tenant, ClientSecretName: secret},
			wantErr: true,
		},
		{
			name:    "tenant alone",
			cfg:     AzureACRConfig{TenantID: tenant},
			wantErr: true,
		},
		{
			name:    "client alone",
			cfg:     AzureACRConfig{ClientID: client},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.ValidateCredentials()
			if tt.wantErr && err == nil {
				t.Fatalf("expected an error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}

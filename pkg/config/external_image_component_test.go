package config

import (
	"testing"

	"github.com/pelletier/go-toml/v2"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/hasher"
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

	tests := map[string]struct {
		cfg     AzureACRConfig
		wantErr bool
	}{
		"nothing set preserves ambient authentication": {
			cfg: AzureACRConfig{},
		},
		"explicit connection opt in": {
			cfg: AzureACRConfig{UseCloudConnection: true},
		},
		"mixed connection and secret": {
			cfg: AzureACRConfig{UseCloudConnection: true, TenantID: tenant, ClientID: client, ClientSecretName: secret}, wantErr: true,
		},
		"mixed connection and certificate": {
			cfg: AzureACRConfig{UseCloudConnection: true, TenantID: tenant, ClientID: client, ClientCertificateName: cert}, wantErr: true,
		},
		"complete with secret": {
			cfg: AzureACRConfig{TenantID: tenant, ClientID: client, ClientSecretName: secret},
		},
		"complete with certificate": {
			cfg: AzureACRConfig{TenantID: tenant, ClientID: client, ClientCertificateName: cert},
		},
		"both credential kinds": {
			cfg:     AzureACRConfig{TenantID: tenant, ClientID: client, ClientSecretName: secret, ClientCertificateName: cert},
			wantErr: true,
		},
		"secret without tenant or client": {
			cfg:     AzureACRConfig{ClientSecretName: secret},
			wantErr: true,
		},
		"certificate without tenant or client": {
			cfg:     AzureACRConfig{ClientCertificateName: cert},
			wantErr: true,
		},
		"tenant and client without a credential": {
			cfg:     AzureACRConfig{TenantID: tenant, ClientID: client},
			wantErr: true,
		},
		"missing tenant only": {
			cfg:     AzureACRConfig{ClientID: client, ClientSecretName: secret},
			wantErr: true,
		},
		"missing client only": {
			cfg:     AzureACRConfig{TenantID: tenant, ClientSecretName: secret},
			wantErr: true,
		},
		"tenant alone": {
			cfg:     AzureACRConfig{TenantID: tenant},
			wantErr: true,
		},
		"client alone": {
			cfg:     AzureACRConfig{ClientID: client},
			wantErr: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err := tt.cfg.ValidateCredentials()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAzureACRLegacyChecksum(t *testing.T) {
	for name, tc := range map[string]struct {
		cfg      AzureACRConfig
		checksum string
	}{
		"ambient": {cfg: AzureACRConfig{RegistryURL: "acme.azurecr.io", ImageURL: "acme.azurecr.io/repo", Tag: "v1"}, checksum: "185c4621059b7713146bff276b61fa0fc557f30ff6eeb2573dff4d00c20ef751"},
		"secret":  {cfg: AzureACRConfig{RegistryURL: "acme.azurecr.io", ImageURL: "acme.azurecr.io/repo", Tag: "v1", TenantID: "tenant", ClientID: "client", ClientSecretName: "secret"}, checksum: "202d555b66a5280914d057b82b8462d1c56ca1c280e9190c35606b4780377ef6"},
	} {
		t.Run(name, func(t *testing.T) {
			checksum, err := hasher.HashStruct(tc.cfg, hasher.StructHasherOptions{EnableOmitEmpty: true})
			require.NoError(t, err)
			require.Equal(t, tc.checksum, checksum)
			tc.cfg.UseCloudConnection = true
			checksum, err = hasher.HashStruct(tc.cfg, hasher.StructHasherOptions{EnableOmitEmpty: true})
			require.NoError(t, err)
			require.NotEqual(t, tc.checksum, checksum)
		})
	}
}

func TestGARLegacyChecksums(t *testing.T) {
	for name, tc := range map[string]struct {
		account, checksum string
	}{
		"ambient": {checksum: "127479ef396a774e2e6d6487f1c59acad73041f5b94eb1e20298c97e2326ad51"},
		"account": {account: "acme@example.com", checksum: "c3f58ad581e5c615f29de4d24a3142d058cf4cdeb38aa693bb3ab27f828bc5f4"},
	} {
		t.Run(name, func(t *testing.T) {
			cfg := GCPGARConfig{GCPProjectID: "acme", GCPRegion: "us-central1", ImageURL: "us-central1-docker.pkg.dev/acme/repo/image", Tag: "v1", ServiceAccountEmail: tc.account}
			checksum, err := hasher.HashStruct(cfg, hasher.StructHasherOptions{EnableOmitEmpty: true})
			require.NoError(t, err)
			require.Equal(t, tc.checksum, checksum)
			cfg.UseCloudConnection = true
			checksum, err = hasher.HashStruct(cfg, hasher.StructHasherOptions{EnableOmitEmpty: true})
			require.NoError(t, err)
			require.NotEqual(t, tc.checksum, checksum)
		})
	}
}

func TestGARAuthenticationDecode(t *testing.T) {
	for name, tc := range map[string]struct {
		body                  string
		connection, wantError bool
	}{
		"ambient":         {},
		"connection":      {body: "use_cloud_connection = true", connection: true},
		"legacy account":  {body: "service_account_email = 'acme@example.com'"},
		"legacy provider": {body: "workload_identity_provider = 'provider'"},
		"mixed account":   {body: "use_cloud_connection = true\nservice_account_email = 'acme@example.com'", connection: true, wantError: true},
		"mixed provider":  {body: "use_cloud_connection = true\nworkload_identity_provider = 'provider'", connection: true, wantError: true},
	} {
		t.Run(name, func(t *testing.T) {
			var cfg ExternalImageComponentConfig
			require.NoError(t, toml.Unmarshal([]byte("[gcp_gar]\ntag = 'v1'\n"+tc.body), &cfg))
			require.Equal(t, tc.connection, cfg.GCPGARImageConfig.UseCloudConnection)
			err := cfg.Validate()
			if tc.wantError {
				require.ErrorContains(t, err, "cannot be combined")
			} else {
				require.NoError(t, err)
			}
		})
	}
}

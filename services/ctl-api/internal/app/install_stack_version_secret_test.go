package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCustomerStackSecrets(t *testing.T) {
	tests := map[string]struct {
		secrets []AppSecretConfig
		want    []InstallStackVersionSecret
	}{
		"no secrets": {},
		"only auto-generated secrets": {
			secrets: []AppSecretConfig{{Name: "jwt_secret", AutoGenerate: true}},
		},
		"customer secrets keep name, description and required": {
			secrets: []AppSecretConfig{
				{Name: "acme_api_key", Description: "API key", Required: true},
				{Name: "optional_token"},
			},
			want: []InstallStackVersionSecret{
				{Name: "acme_api_key", Description: "API key", Required: true},
				{Name: "optional_token"},
			},
		},
		"auto-generated secrets are left out": {
			secrets: []AppSecretConfig{
				{Name: "jwt_secret", AutoGenerate: true},
				{Name: "acme_api_key"},
			},
			want: []InstallStackVersionSecret{{Name: "acme_api_key"}},
		},
		"default values are not served": {
			secrets: []AppSecretConfig{{Name: "acme_api_key", Default: "not-served"}},
			want:    []InstallStackVersionSecret{{Name: "acme_api_key"}},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.want, CustomerStackSecrets(tc.secrets))
		})
	}
}

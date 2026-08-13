package validate

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/config"
)

func TestValidateDefaultLabels(t *testing.T) {
	t.Run("static and valid templated values pass", func(t *testing.T) {
		require.NoError(t, ValidateDefaultLabels(&config.AppConfig{
			DefaultLabels: map[string]string{
				"tier":   "prod",
				"region": "{{ .nuon.cloud_account.aws.region }}",
			},
		}))
	})

	t.Run("templated key rejected", func(t *testing.T) {
		err := ValidateDefaultLabels(&config.AppConfig{
			DefaultLabels: map[string]string{"{{ .nuon.install.id }}": "v"},
		})
		require.ErrorContains(t, err, "only label values may be templated")
	})

	t.Run("unparsable template rejected", func(t *testing.T) {
		err := ValidateDefaultLabels(&config.AppConfig{
			DefaultLabels: map[string]string{"region": "{{ .nuon.cloud_account"},
		})
		require.ErrorContains(t, err, "template")
	})

	t.Run("install label colliding with default rejected", func(t *testing.T) {
		err := ValidateDefaultLabels(&config.AppConfig{
			DefaultLabels: map[string]string{"tier": "prod"},
			Installs: []*config.Install{
				{Name: "customer-a", Labels: map[string]string{"tier": "gold"}},
			},
		})
		require.ErrorContains(t, err, "collides with a default label")
	})

	t.Run("install labels without collisions pass", func(t *testing.T) {
		require.NoError(t, ValidateDefaultLabels(&config.AppConfig{
			DefaultLabels: map[string]string{"tier": "prod"},
			Installs: []*config.Install{
				{Name: "customer-a", Labels: map[string]string{"env": "prod"}},
			},
		}))
	})
}

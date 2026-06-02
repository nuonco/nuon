package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnvAccentConfig_Validate(t *testing.T) {
	t.Run("empty config is valid", func(t *testing.T) {
		cfg := EnvAccentConfig{}
		assert.NoError(t, cfg.Validate())
	})

	t.Run("label key without values is valid", func(t *testing.T) {
		cfg := EnvAccentConfig{LabelKey: "env"}
		assert.NoError(t, cfg.Validate())
	})

	t.Run("values without label key is rejected", func(t *testing.T) {
		cfg := EnvAccentConfig{
			Values: map[string]EnvAccentColor{"production": EnvAccentColorError},
		}
		assert.Error(t, cfg.Validate())
	})

	t.Run("default config is valid", func(t *testing.T) {
		assert.NoError(t, DefaultEnvAccentConfig().Validate())
	})

	t.Run("invalid color is rejected", func(t *testing.T) {
		cfg := EnvAccentConfig{
			LabelKey: "env",
			Values:   map[string]EnvAccentColor{"production": EnvAccentColor("purple")},
		}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid accent color")
	})

	t.Run("empty value key is rejected", func(t *testing.T) {
		cfg := EnvAccentConfig{
			LabelKey: "env",
			Values:   map[string]EnvAccentColor{"": EnvAccentColorError},
		}
		assert.Error(t, cfg.Validate())
	})

	t.Run("every allowed color passes", func(t *testing.T) {
		for _, color := range ValidEnvAccentColors() {
			cfg := EnvAccentConfig{
				LabelKey: "env",
				Values:   map[string]EnvAccentColor{"v": color},
			}
			assert.NoError(t, cfg.Validate(), "color %s should be valid", color)
		}
	})
}

package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestEnvConfig struct {
	Env Env `config:"env"`
}

func TestEnv(t *testing.T) {

	t.Run("not set", func(t *testing.T) {
		var cfg TestEnvConfig
		require.NoError(t, LoadInto(nil, &cfg))
		assert.Equal(t, Local, cfg.Env)
	})

	t.Run("valid", func(t *testing.T) {
		os.Setenv("ENV", "demo")

		var cfg TestEnvConfig
		require.NoError(t, LoadInto(nil, &cfg))
		assert.Equal(t, Demo, cfg.Env)

		os.Clearenv()
	})

	t.Run("invalid", func(t *testing.T) {

		os.Setenv("ENV", "bogus")

		var cfg TestEnvConfig
		require.NoError(t, LoadInto(nil, &cfg))
		assert.Equal(t, Local, cfg.Env)

		os.Clearenv()
	})
}

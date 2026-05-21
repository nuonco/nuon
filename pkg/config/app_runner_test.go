package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppRunnerConfig_Parse_AWS_RequiresInitScriptURL(t *testing.T) {
	cfg := &AppRunnerConfig{
		RunnerType: "aws",
	}
	err := cfg.parse()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "init_script_url is required when runner_type is aws")
}

func TestAppRunnerConfig_Parse_AWS_WithInitScriptURL(t *testing.T) {
	cfg := &AppRunnerConfig{
		RunnerType:    "aws",
		InitScriptURL: "https://example.com/init.sh",
	}
	require.NoError(t, cfg.parse())
}

func TestAppRunnerConfig_Parse_NonAWS_NoInitScriptURL(t *testing.T) {
	cfg := &AppRunnerConfig{
		RunnerType: "azure",
	}
	require.NoError(t, cfg.parse())
}

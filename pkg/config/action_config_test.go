package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestActionConfig_ParseImageSteps(t *testing.T) {
	t.Run("command is allowed with image", func(t *testing.T) {
		cfg := &ActionConfig{
			Name:    "db_migrate",
			Timeout: "10m",
			Image:   "ghcr.io/acme/migrate-tools:v1.4.0",
			Triggers: []*ActionTriggerConfig{
				{Type: "manual"},
			},
			Steps: []*ActionStepConfig{
				{Name: "migrate", Command: "migrate up"},
			},
		}
		require.NoError(t, cfg.parse())
	})

	t.Run("inline_contents is allowed with image", func(t *testing.T) {
		cfg := &ActionConfig{
			Name:    "db_migrate",
			Timeout: "10m",
			Image:   "ghcr.io/acme/migrate-tools:v1.4.0",
			Triggers: []*ActionTriggerConfig{
				{Type: "manual"},
			},
			Steps: []*ActionStepConfig{
				{Name: "migrate", InlineContents: "#!/bin/sh\nmigrate up\n"},
			},
		}
		require.NoError(t, cfg.parse())
	})

	t.Run("repo steps are allowed with image", func(t *testing.T) {
		cfg := &ActionConfig{
			Name:    "healthcheck",
			Timeout: "10s",
			Image:   "ghcr.io/acme/tools:v1",
			Triggers: []*ActionTriggerConfig{
				{Type: "manual"},
			},
			Steps: []*ActionStepConfig{
				{
					Name:    "check",
					Command: "./healthcheck",
					PublicRepo: &PublicRepoConfig{
						Repo:      "nuonco/actions",
						Directory: "common",
						Branch:    "main",
					},
				},
			},
		}
		require.NoError(t, cfg.parse())
	})

	t.Run("step with neither command, inline_contents, nor repo is rejected", func(t *testing.T) {
		cfg := &ActionConfig{
			Name:    "empty",
			Timeout: "10s",
			Image:   "ghcr.io/acme/tools:v1",
			Triggers: []*ActionTriggerConfig{
				{Type: "manual"},
			},
			Steps: []*ActionStepConfig{
				{Name: "noop"},
			},
		}
		err := cfg.parse()
		require.Error(t, err)
		require.Contains(t, err.Error(), "must use command, inline_contents, or a repo")
	})
}

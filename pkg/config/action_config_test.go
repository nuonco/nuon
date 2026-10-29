package config

import (
	"fmt"
	"slices"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
	"hegel.dev/go/hegel"
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

func TestActionConfig_DependencyProperties(t *testing.T) {
	hegel.Test(t, func(t *hegel.T) {
		names := drawDistinctStrings(t, hegel.FromRegex(`[a-z]{1,8}`, true), 0, 12)
		expected := append(slices.Clone(names), "explicit_dependency")
		sort.Strings(expected)
		command := "echo"
		for _, name := range names {
			command += fmt.Sprintf(" {{.nuon.components.%s.outputs.value}} {{.nuon.components.%s.outputs.other}}", name, name)
		}

		var previous *AppConfig
		for range 3 {
			cfg := baseConfig()
			action := &ActionConfig{
				Name:         "inspect",
				Dependencies: []string{"explicit_dependency", "explicit_dependency"},
				Steps:        []*ActionStepConfig{{Name: "inspect", Command: command}},
			}
			cfg.Actions = []*ActionConfig{action}
			require.NoError(t, cfg.Parse())
			require.Equal(t, expected, action.Dependencies)
			require.NoError(t, cfg.Parse())
			require.Equal(t, expected, action.Dependencies)

			if previous != nil {
				require.False(t, cfg.Diff(previous).Summary().HasChanged)
			}
			previous = cfg
		}
	})
}

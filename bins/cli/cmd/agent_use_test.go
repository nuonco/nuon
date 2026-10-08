package cmd

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/bins/cli/internal/config"
	"github.com/nuonco/nuon/pkg/agentclient"
)

func clearAgentEnv(t *testing.T) {
	t.Helper()
	for _, key := range append(
		[]string{agentclient.EnvVar, agentclient.AIAgentEnvVar},
		agentclient.ProductEnvKeys()...,
	) {
		t.Setenv(key, "")
	}
}

func TestApplyAgentMode(t *testing.T) {
	tests := map[string]struct {
		env             map[string]string
		wantAgent       string
		wantInteractive bool
	}{
		"cursor": {
			env:       map[string]string{"CURSOR_AGENT": "1"},
			wantAgent: agentclient.Cursor,
		},
		"ai agent": {
			env:       map[string]string{agentclient.AIAgentEnvVar: "amp@1.0.0"},
			wantAgent: "amp",
		},
		"off": {
			env:             map[string]string{agentclient.EnvVar: "off", "CURSOR_AGENT": "1"},
			wantInteractive: true,
		},
		"person": {
			wantInteractive: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			clearAgentEnv(t)
			for k, v := range tc.env {
				t.Setenv(k, v)
			}
			c := &cli{cfg: &config.Config{Interactive: true}}

			c.applyAgentMode()

			require.Equal(t, tc.wantAgent, c.cfg.Agent)
			require.Equal(t, tc.wantInteractive, c.cfg.Interactive)
		})
	}
}

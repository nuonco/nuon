package cmd

import (
	"strings"
	"testing"

	"github.com/nuonco/nuon/pkg/agentclient"
)

func TestAgentsSetupGuideDetectsAgent(t *testing.T) {
	t.Run("cursor", func(t *testing.T) {
		t.Setenv(agentclient.EnvVar, agentclient.Cursor)
		guide := agentsSetupGuide(nil)
		if !strings.Contains(guide, "agent (cursor) detected") {
			t.Fatalf("guide missing cursor detection:\n%s", guide)
		}
	})

	t.Run("claude", func(t *testing.T) {
		t.Setenv(agentclient.EnvVar, agentclient.Claude)
		guide := agentsSetupGuide(nil)
		if !strings.Contains(guide, "agent (claude) detected") {
			t.Fatalf("guide missing claude detection:\n%s", guide)
		}
	})

	t.Run("off", func(t *testing.T) {
		t.Setenv(agentclient.EnvVar, "off")
		guide := agentsSetupGuide(nil)
		if strings.Contains(guide, "agent (") {
			t.Fatalf("guide showed a detection line while detection was off:\n%s", guide)
		}
	})

	t.Run("neither", func(t *testing.T) {
		t.Setenv(agentclient.EnvVar, "")
		t.Setenv("CURSOR_AGENT", "")
		t.Setenv("CURSOR_INVOKED_AS", "")
		t.Setenv("CLAUDECODE", "")
		t.Setenv(agentclient.AIAgentEnvVar, "")
		guide := agentsSetupGuide(nil)
		if strings.Contains(guide, "agent (") {
			t.Fatalf("guide showed a detection line with no agent:\n%s", guide)
		}
	})
}

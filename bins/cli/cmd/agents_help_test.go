package cmd

import (
	"strings"
	"testing"

	"github.com/nuonco/nuon/bins/cli/internal/config"
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
		t.Setenv(agentclient.AIAgentEnvVar, "")
		for _, key := range agentclient.ProductEnvKeys() {
			t.Setenv(key, "")
		}
		guide := agentsSetupGuide(nil)
		if strings.Contains(guide, "agent (") {
			t.Fatalf("guide showed a detection line with no agent:\n%s", guide)
		}
	})
}

func TestAgentsHelpTextFollowsDetectedAgent(t *testing.T) {
	t.Run("agent", func(t *testing.T) {
		c := &cli{cfg: &config.Config{Agent: agentclient.Cursor, APIURL: "https://api.nuon.co"}}
		out := c.agentsHelpText()
		if !strings.Contains(out, "Nuon agent context") {
			t.Fatalf("agent help missing orientation markdown:\n%s", out)
		}
		if strings.Contains(out, "Drive Nuon with an LLM agent.") {
			t.Fatalf("agent help printed the human setup guide:\n%s", out)
		}
	})

	t.Run("human", func(t *testing.T) {
		c := &cli{cfg: &config.Config{}}
		out := c.agentsHelpText()
		if !strings.Contains(out, "Drive Nuon with an LLM agent.") {
			t.Fatalf("human help missing setup guide:\n%s", out)
		}
		if strings.Contains(out, "Nuon agent context") {
			t.Fatalf("human help printed orientation markdown:\n%s", out)
		}
	})
}

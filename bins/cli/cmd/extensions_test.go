package cmd

import (
	"testing"

	"github.com/spf13/viper"

	"github.com/nuonco/nuon/bins/cli/internal/config"
)

func TestExtensionEnvIncludesInstallIDFromConfig(t *testing.T) {
	v := viper.New()
	v.Set("install_id", "inst_123")

	c := &cli{cfg: &config.Config{Viper: v}}
	env := c.extensionEnv()

	if env["NUON_INSTALL_ID"] != "inst_123" {
		t.Fatalf("expected NUON_INSTALL_ID to be %q, got %q", "inst_123", env["NUON_INSTALL_ID"])
	}
}

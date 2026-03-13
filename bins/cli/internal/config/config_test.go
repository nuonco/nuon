package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewConfigLoadsInstallID(t *testing.T) {
	t.Setenv("NUON_INSTALL_ID", "inst_123")
	cfgPath := filepath.Join(t.TempDir(), "nuon.yml")
	if err := os.WriteFile(cfgPath, []byte("api_url: https://api.nuon.co\n"), 0o600); err != nil {
		t.Fatalf("expected to write config fixture: %v", err)
	}

	cfg, err := NewConfig(cfgPath)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.InstallID != "inst_123" {
		t.Fatalf("expected InstallID to be %q, got %q", "inst_123", cfg.InstallID)
	}
}

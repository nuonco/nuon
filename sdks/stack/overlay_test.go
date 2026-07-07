package stack

import (
	"testing"

	"github.com/nuonco/nuon/sdks/stack/internal/core"
)

func TestOverlayInputsAndSecrets(t *testing.T) {
	newInstaller := func() *Installer {
		return &Installer{
			opts: Options{InstallID: "inst123"},
			cfg: &Config{
				InstallInputs:       map[string]string{"domain": ""},
				RequiredInputs:      []string{"domain", "node_count"},
				AutoGenerateSecrets: []string{"auto_token"},
				Secrets: map[string]core.SecretInput{
					"db_password": {Required: true},
				},
			},
		}
	}

	t.Run("overlays declared inputs and secrets", func(t *testing.T) {
		i := newInstaller()
		i.opts.InstallInputs = map[string]string{"domain": "example.com", "node_count": "3"}
		i.opts.Secrets = map[string]string{"db_password": "hunter2"}
		if err := i.overlayInputsAndSecrets(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got := i.cfg.InstallInputs["domain"]; got != "example.com" {
			t.Errorf("domain = %q, want example.com", got)
		}
		if got := i.cfg.InstallInputs["node_count"]; got != "3" {
			t.Errorf("node_count = %q, want 3", got)
		}
		if got := i.cfg.Secrets["db_password"].Value; got != "hunter2" {
			t.Errorf("db_password = %q, want hunter2", got)
		}
	})

	t.Run("rejects unknown input", func(t *testing.T) {
		i := newInstaller()
		i.opts.InstallInputs = map[string]string{"nope": "x"}
		if err := i.overlayInputsAndSecrets(); err == nil {
			t.Fatal("expected error for unknown input")
		}
	})

	t.Run("rejects unknown secret", func(t *testing.T) {
		i := newInstaller()
		i.opts.Secrets = map[string]string{"nope": "x"}
		if err := i.overlayInputsAndSecrets(); err == nil {
			t.Fatal("expected error for unknown secret")
		}
	})

	t.Run("rejects value for auto-generated secret", func(t *testing.T) {
		i := newInstaller()
		i.opts.Secrets = map[string]string{"auto_token": "x"}
		if err := i.overlayInputsAndSecrets(); err == nil {
			t.Fatal("expected error setting auto-generated secret")
		}
	})
}

func TestApplyTerraformOptionsBackendDefaults(t *testing.T) {
	t.Run("aws defaults key and region", func(t *testing.T) {
		i := &Installer{
			opts: Options{
				InstallID: "inst123",
				Cloud:     core.CloudAWS,
				AWSRegion: "us-west-2",
				Backend:   TerraformBackend{Bucket: "my-state"},
			},
			cfg: &Config{Cloud: core.CloudAWS},
		}
		i.applyTerraformOptions()
		be := i.cfg.TerraformBackend
		if be == nil {
			t.Fatal("backend not set")
		}
		if be.Key != "nuon/inst123/terraform.tfstate" {
			t.Errorf("key = %q", be.Key)
		}
		if be.Region != "us-west-2" {
			t.Errorf("region = %q, want us-west-2", be.Region)
		}
	})

	t.Run("gcp defaults prefix", func(t *testing.T) {
		i := &Installer{
			opts: Options{
				InstallID: "inst123",
				Cloud:     core.CloudGCP,
				GCP:       GCPOptions{ProjectID: "p", Region: "us-central1"},
				Backend:   TerraformBackend{Bucket: "my-state"},
			},
			cfg: &Config{Cloud: core.CloudGCP},
		}
		i.applyTerraformOptions()
		be := i.cfg.TerraformBackend
		if be == nil {
			t.Fatal("backend not set")
		}
		if be.Prefix != "nuon/inst123" {
			t.Errorf("prefix = %q, want nuon/inst123", be.Prefix)
		}
	})

	t.Run("no bucket leaves backend nil", func(t *testing.T) {
		i := &Installer{
			opts: Options{InstallID: "inst123", Cloud: core.CloudAWS},
			cfg:  &Config{Cloud: core.CloudAWS},
		}
		i.applyTerraformOptions()
		if i.cfg.TerraformBackend != nil {
			t.Errorf("expected nil backend, got %+v", i.cfg.TerraformBackend)
		}
	})
}

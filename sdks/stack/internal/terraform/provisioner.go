// Package terraform provisions a Nuon install stack by applying the
// install-stacks/aws Terraform module. It downloads a terraform binary on
// first use (no install required), fetches the module over HTTPS, renders
// tfvars from the install config, and drives terraform via terraform-exec.
// It is one implementation of core.Provisioner; outputs are mapped back into
// the same core.Outputs the AWS SDK method produces.
package terraform

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/hashicorp/terraform-exec/tfexec"

	"github.com/nuonco/nuon/sdks/stack/internal/core"
)

const tfvarsFilename = "terraform.tfvars.json"

// Provisioner applies the install-stacks/aws Terraform module.
type Provisioner struct {
	region string
}

var _ core.Provisioner = (*Provisioner)(nil)

// ReportsOwnRun is true: the install-stacks/aws module's phone-home reports the
// run to ctl-api, so the stack package must not also report via run_client.
func (p *Provisioner) ReportsOwnRun() bool { return true }

// New constructs a Terraform provisioner for the given region.
func New(region string) *Provisioner {
	return &Provisioner{region: region}
}

// Provision fetches the module, renders tfvars, runs init + apply, and returns
// the resolved outputs.
func (p *Provisioner) Provision(ctx context.Context, log, sysLog *slog.Logger, cfg *core.Config, _ core.Kind) (*core.Outputs, error) {
	tf, err := p.prepare(ctx, log, sysLog, cfg)
	if err != nil {
		return nil, err
	}
	log.Info("terraform apply starting")
	if err := tf.Apply(ctx); err != nil {
		return nil, fmt.Errorf("terraform apply: %w", err)
	}
	log.Info("terraform apply complete")
	return p.readOutputs(ctx, tf)
}

// Deprovision runs init + destroy.
func (p *Provisioner) Deprovision(ctx context.Context, log, sysLog *slog.Logger, cfg *core.Config) error {
	tf, err := p.prepare(ctx, log, sysLog, cfg)
	if err != nil {
		return err
	}
	log.Info("terraform destroy starting")
	if err := tf.Destroy(ctx); err != nil {
		return fmt.Errorf("terraform destroy: %w", err)
	}
	log.Info("terraform destroy complete")
	return nil
}

// Status returns the persisted outputs without mutating resources. It still
// needs the module + provider present to read outputs, so it prepares the work
// dir but runs no apply.
func (p *Provisioner) Status(ctx context.Context, cfg *core.Config) (*core.Outputs, error) {
	tf, err := p.prepare(ctx, slog.Default(), slog.Default(), cfg)
	if err != nil {
		return nil, err
	}
	return p.readOutputs(ctx, tf)
}

// prepare assembles the work dir (module + tfvars), resolves the terraform
// binary, wires logging, and runs init.
func (p *Provisioner) prepare(ctx context.Context, log, sysLog *slog.Logger, cfg *core.Config) (*tfexec.Terraform, error) {
	workDir, err := workDirFor(cfg.InstallID)
	if err != nil {
		return nil, err
	}

	log.Info("fetching terraform module", "subdir", moduleSubdir(cfg))
	if err := fetchModule(ctx, cfg.TerraformModuleURL, cfg.TerraformModuleSubdir, workDir); err != nil {
		return nil, err
	}

	vars, err := renderTFVars(cfg)
	if err != nil {
		return nil, fmt.Errorf("render tfvars: %w", err)
	}
	if err := os.WriteFile(filepath.Join(workDir, tfvarsFilename), vars, 0o600); err != nil {
		return nil, fmt.Errorf("write tfvars: %w", err)
	}

	binDir, err := binCacheDir()
	if err != nil {
		return nil, err
	}
	execPath, err := resolveTerraform(ctx, cfg.TerraformVersion, binDir)
	if err != nil {
		return nil, err
	}

	tf, err := tfexec.NewTerraform(workDir, execPath)
	if err != nil {
		return nil, fmt.Errorf("init terraform-exec: %w", err)
	}
	tf.SetStdout(&slogWriter{l: log})
	tf.SetStderr(&slogWriter{l: sysLog})

	log.Info("terraform init starting")
	if err := tf.Init(ctx); err != nil {
		return nil, fmt.Errorf("terraform init: %w", err)
	}
	log.Info("terraform init complete")
	return tf, nil
}

func (p *Provisioner) readOutputs(ctx context.Context, tf *tfexec.Terraform) (*core.Outputs, error) {
	meta, err := tf.Output(ctx)
	if err != nil {
		return nil, fmt.Errorf("terraform output: %w", err)
	}
	return outputsToCore(meta)
}

func moduleSubdir(cfg *core.Config) string {
	if cfg.TerraformModuleSubdir != "" {
		return cfg.TerraformModuleSubdir
	}
	return defaultModuleSubdir
}

func workDirFor(installID string) (string, error) {
	cache, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(cache, "nuon", "installer-sdk", installID, "tf")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create terraform work dir: %w", err)
	}
	return dir, nil
}

func binCacheDir() (string, error) {
	cache, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(cache, "nuon", "installer-sdk", "bin", "terraform"), nil
}

// Package terraform provisions a Nuon install stack by applying an
// install-stacks Terraform module. It downloads a terraform binary on first
// use (no install required), fetches the module over HTTPS, renders tfvars
// from the install config, and drives terraform via terraform-exec. It is one
// implementation of core.Provisioner; outputs are mapped back into the same
// core.Outputs the SDK methods produce.
//
// The engine here is cloud-agnostic: the per-cloud specifics (module subdir,
// tfvars rendering, output mapping) live behind a moduleAdapter selected by
// cloud at construction (see adapter.go).
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

// Provisioner applies an install-stacks Terraform module for a single cloud,
// selected by the moduleAdapter built at construction.
type Provisioner struct {
	cloud   core.Cloud
	adapter moduleAdapter
}

var _ core.Provisioner = (*Provisioner)(nil)

// ReportsOwnRun is true: the install-stacks module's phone-home reports the
// run to ctl-api, so the stack package must not also report via run_client.
func (p *Provisioner) ReportsOwnRun() bool { return true }

// New constructs a Terraform provisioner for the given cloud. It returns an
// error when the cloud has no module adapter.
func New(cloud core.Cloud) (*Provisioner, error) {
	adapter, err := adapterFor(cloud)
	if err != nil {
		return nil, err
	}
	return &Provisioner{cloud: cloud, adapter: adapter}, nil
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

// prepare assembles the work dir (module + tfvars + backend), resolves the
// terraform binary, wires logging, and runs init.
func (p *Provisioner) prepare(ctx context.Context, log, sysLog *slog.Logger, cfg *core.Config) (*tfexec.Terraform, error) {
	workDir, err := p.workDir(cfg)
	if err != nil {
		return nil, err
	}

	subdir := p.moduleSubdir(cfg)
	log.Info("fetching terraform module", "subdir", subdir)
	if err := fetchModule(ctx, cfg.TerraformModuleURL, subdir, workDir); err != nil {
		return nil, err
	}

	vars, err := p.adapter.RenderTFVars(cfg)
	if err != nil {
		return nil, fmt.Errorf("render tfvars: %w", err)
	}
	if err := os.WriteFile(filepath.Join(workDir, tfvarsFilename), vars, 0o600); err != nil {
		return nil, fmt.Errorf("write tfvars: %w", err)
	}

	initOpts, err := p.writeBackend(cfg, workDir)
	if err != nil {
		return nil, err
	}

	execPath, err := p.resolveBinary(ctx, cfg)
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
	if err := tf.Init(ctx, initOpts...); err != nil {
		return nil, fmt.Errorf("terraform init: %w", err)
	}
	log.Info("terraform init complete")
	return tf, nil
}

// workDir returns the directory to assemble the module and run terraform in. A
// Config-supplied dir is used as-is (created if missing); otherwise a fresh temp
// dir — state lives in the remote backend, so the work dir is disposable.
func (p *Provisioner) workDir(cfg *core.Config) (string, error) {
	if cfg.TerraformWorkDir != "" {
		if err := os.MkdirAll(cfg.TerraformWorkDir, 0o755); err != nil {
			return "", fmt.Errorf("create terraform work dir: %w", err)
		}
		return cfg.TerraformWorkDir, nil
	}
	dir, err := os.MkdirTemp("", "nuon-stack-tf-")
	if err != nil {
		return "", fmt.Errorf("create terraform work dir: %w", err)
	}
	return dir, nil
}

// resolveBinary returns the terraform binary to run: a Config-supplied path is
// used as-is (no download); otherwise one is fetched/cached via hc-install.
func (p *Provisioner) resolveBinary(ctx context.Context, cfg *core.Config) (string, error) {
	if cfg.TerraformExecPath != "" {
		return cfg.TerraformExecPath, nil
	}
	binDir, err := binCacheDir()
	if err != nil {
		return "", err
	}
	return resolveTerraform(ctx, cfg.TerraformVersion, binDir)
}

// writeBackend renders a partial backend block into the work dir and returns the
// `terraform init` options carrying the backend config. When no remote backend
// is configured it writes nothing and terraform uses local state.
func (p *Provisioner) writeBackend(cfg *core.Config, workDir string) ([]tfexec.InitOption, error) {
	be := cfg.TerraformBackend
	if be == nil || be.Bucket == "" {
		return nil, nil
	}
	block := fmt.Sprintf("terraform {\n  backend %q {}\n}\n", p.adapter.BackendType())
	if err := os.WriteFile(filepath.Join(workDir, "nuon_backend.tf"), []byte(block), 0o600); err != nil {
		return nil, fmt.Errorf("write backend config: %w", err)
	}
	var opts []tfexec.InitOption
	for _, kv := range p.adapter.BackendConfigKV(be) {
		opts = append(opts, tfexec.BackendConfig(kv))
	}
	return opts, nil
}

func (p *Provisioner) readOutputs(ctx context.Context, tf *tfexec.Terraform) (*core.Outputs, error) {
	meta, err := tf.Output(ctx)
	if err != nil {
		return nil, fmt.Errorf("terraform output: %w", err)
	}
	return p.adapter.MapOutputs(meta)
}

// moduleSubdir resolves the install-stacks subdir for this run: an explicit
// Config override wins, otherwise the adapter's default for the cloud.
func (p *Provisioner) moduleSubdir(cfg *core.Config) string {
	if cfg.TerraformModuleSubdir != "" {
		return cfg.TerraformModuleSubdir
	}
	return p.adapter.ModuleSubdir()
}

func binCacheDir() (string, error) {
	cache, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(cache, "nuon", "installer-sdk", "bin", "terraform"), nil
}

package terraform

import (
	"context"
	"log"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
)

const (
	defaultTerraformVersion = "v1.3.7"
)

type terraformInstaller interface {
	initTerraformInstaller(*log.Logger) error
	installTerraform(context.Context) (string, error)
	removeTerraform(context.Context) error
}

// hcTerraformInstaller is the interface that _we_ use here to manage the terraform binary
type hcTerraformInstaller interface {
	SetLogger(*log.Logger)
	Install(context.Context) (string, error)
	Remove(context.Context) error
}

type tfInstaller struct {
	installer hcTerraformInstaller
}

var _ terraformInstaller = (*tfInstaller)(nil)

func (t *tfInstaller) initTerraformInstaller(l *log.Logger) error {
	ver, err := version.NewVersion(defaultTerraformVersion)
	if err != nil {
		return err
	}
	t.installer = &releases.ExactVersion{
		Product: product.Terraform,
		Version: ver,
	}
	t.installer.SetLogger(l)
	return nil
}

func (t *tfInstaller) installTerraform(ctx context.Context) (string, error) {
	return t.installer.Install(ctx)
}

func (t *tfInstaller) removeTerraform(ctx context.Context) error {
	return t.installer.Remove(ctx)
}

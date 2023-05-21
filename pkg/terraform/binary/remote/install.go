package remote

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
)

func (r *remote) Install(ctx context.Context, lg *log.Logger, dir string) (string, error) {
	installer := r.getInstaller(lg, dir)
	execPath, err := installer.Install(ctx)
	if err != nil {
		return "", fmt.Errorf("unable to install: %w", err)
	}

	return execPath, nil
}

func (r *remote) getInstaller(lg *log.Logger, dir string) *releases.ExactVersion {
	installer := &releases.ExactVersion{
		Product:    product.Terraform,
		Version:    r.Version,
		InstallDir: dir,
	}
	installer.SetLogger(lg)

	return installer
}

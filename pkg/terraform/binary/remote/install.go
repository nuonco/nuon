package remote

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
)

func (r *remote) Install(ctx context.Context, lg hclog.Logger, dir string) (string, error) {
	binLog := lg.StandardLogger(nil)
	installer := r.getInstaller(binLog, dir)
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

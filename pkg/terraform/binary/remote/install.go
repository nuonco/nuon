package remote

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const tracerName = "github.com/nuonco/nuon/pkg/terraform/binary/remote"

func (r *remote) Install(ctx context.Context, lg hclog.Logger, dir string) (string, error) {
	ctx, span := otel.Tracer(tracerName).Start(ctx, "terraform.binary_install",
		trace.WithAttributes(
			attribute.String("terraform.version", r.Version.String()),
			attribute.String("install.dir", dir),
		),
	)
	defer span.End()

	binLog := lg.StandardLogger(&hclog.StandardLoggerOptions{ForceLevel: hclog.Trace})
	installer := r.getInstaller(binLog, dir)

	execPath, err := installer.Install(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return "", fmt.Errorf("unable to install: %w", err)
	}

	span.SetAttributes(attribute.String("terraform.exec_path", execPath))
	r.version = installer
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

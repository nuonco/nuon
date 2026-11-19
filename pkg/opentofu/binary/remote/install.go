package remote

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/go-hclog"
	"github.com/opentofu/tofudl"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const tracerName = "github.com/nuonco/nuon/pkg/opentofu/binary/remote"

// tofuBinaryName is the on-disk filename the downloaded OpenTofu CLI is
// written as inside the install dir.
const tofuBinaryName = "tofu"

func (r *remote) Install(ctx context.Context, lg hclog.Logger, dir string) (string, error) {
	ctx, span := otel.Tracer(tracerName).Start(ctx, "tofu.binary_install",
		trace.WithAttributes(
			attribute.String("tofu.version", string(r.Version)),
			attribute.String("install.dir", dir),
		),
	)
	defer span.End()

	downloader, err := tofudl.New()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return "", fmt.Errorf("unable to create tofudl downloader: %w", err)
	}

	lg.Debug("downloading opentofu via tofudl", "version", string(r.Version))

	byts, err := downloader.Download(ctx, tofudl.DownloadOptVersion(r.Version))
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return "", fmt.Errorf("unable to download opentofu %s: %w", r.Version, err)
	}

	execPath := filepath.Join(dir, tofuBinaryName)
	if err := os.WriteFile(execPath, byts, 0o755); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return "", fmt.Errorf("unable to write %s: %w", execPath, err)
	}

	span.SetAttributes(attribute.String("tofu.exec_path", execPath))
	r.installedPath = execPath
	return execPath, nil
}

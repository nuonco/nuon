package workspace

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/go-hclog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const tracerName = "github.com/nuonco/nuon/pkg/opentofu/workspace"

// binarySource is the optional Source() method exposed by binary
// implementations (pkg/terraform/binary/local and pkg/opentofu/binary/remote)
// so LoadBinary can stamp the trace span with which path was taken without
// coupling pkg/terraform/binary's interface to observability concerns.
type binarySource interface {
	Source() string
}

// LoadBinary installs the binary using the provided binary
func (w *workspace) LoadBinary(ctx context.Context, log hclog.Logger) (retErr error) {
	source := "unknown"
	if s, ok := w.Binary.(binarySource); ok {
		source = s.Source()
	}

	ctx, span := otel.Tracer(tracerName).Start(ctx, "tofu.binary_load",
		trace.WithAttributes(
			attribute.String("nuon.binary.source", source),
		),
	)
	defer func() {
		if retErr != nil {
			span.RecordError(retErr)
			span.SetStatus(codes.Error, retErr.Error())
		}
		span.End()
	}()

	if err := w.Binary.Init(ctx); err != nil {
		return fmt.Errorf("unable to initialize binary: %w", err)
	}

	installPath := filepath.Join(w.root, "bins")
	if err := os.MkdirAll(installPath, defaultDirPermissions); err != nil {
		return fmt.Errorf("unable to create bins path: %w", err)
	}

	execPath, err := w.Binary.Install(ctx, log, installPath)
	if err != nil {
		return fmt.Errorf("unable to install binary: %w", err)
	}
	w.execPath = execPath
	span.SetAttributes(attribute.String("tofu.exec_path", execPath))

	return nil
}

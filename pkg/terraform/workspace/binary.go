package workspace

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/hashicorp/go-hclog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const tracerName = "github.com/nuonco/nuon/pkg/terraform/workspace"

type binarySource interface {
	Source() string
}

func (w *workspace) LoadBinary(ctx context.Context, log hclog.Logger) (retErr error) {
	source := "unknown"
	if s, ok := w.Binary.(binarySource); ok {
		source = s.Source()
	}

	ctx, span := otel.Tracer(tracerName).Start(ctx, "terraform.binary_load",
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
	span.SetAttributes(attribute.String("terraform.exec_path", execPath))

	return nil
}

func (w *workspace) loadLocalBinary(ctx context.Context) {
	terraformPath, err := exec.LookPath("terraform")
	if err != nil {
		panic(err)
	}

	err = os.MkdirAll(filepath.Join(w.root, "bins"), 0755)
	if err != nil {
		panic(err)
	}

	err = copyFile(terraformPath, filepath.Join(w.root, "/bins/terraform"))
	if err != nil {
		panic(err)
	}
}

func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	return os.Chmod(dst, srcInfo.Mode())
}

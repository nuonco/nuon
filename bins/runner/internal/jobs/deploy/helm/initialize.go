package helm

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	pkgctx "github.com/nuonco/nuon/pkg/runner/ctx"
	ociarchive "github.com/nuonco/nuon/pkg/runner/oci/archive"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

const (
	defaultChartPackageFilename string = "chart.tgz"
)

func (h *handler) Initialize(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) (retErr error) {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	if h.isRecovery() {
		l.Info("recovery job, no chart to unpack")
		return nil
	}

	arch := ociarchive.New()
	if err := arch.Initialize(ctx); err != nil {
		return fmt.Errorf("unable to initialize archive: %w", err)
	}
	h.state.arch = arch
	defer func() {
		if retErr != nil {
			retErr = errors.Join(retErr, h.Cleanup(context.WithoutCancel(ctx), job, jobExecution))
		}
	}()

	l.Info("unpacking archive...")
	if err := h.state.arch.Unpack(ctx, h.state.srcCfg, h.state.srcTag); err != nil {
		return fmt.Errorf("unable to unpack archive: %w", err)
	}

	h.state.chartPath = filepath.Join(h.state.arch.BasePath(), defaultChartPackageFilename)

	_, err = os.Stat(h.state.chartPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("invalid archive, the chart was not found after unpacking: %w", err)
		}

		return fmt.Errorf("error checking chart after initializing: %w", err)
	}

	return nil
}

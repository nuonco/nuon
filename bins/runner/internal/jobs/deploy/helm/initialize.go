package helm

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/nuonco/nuon-runner-go/models"
	pkgctx "github.com/powertoolsdev/mono/bins/runner/internal/pkg/ctx"
)

const (
	defaultChartPackageFilename string = "chart.tgz"
)

func (h *handler) Initialize(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	l.Info("initializing archive...")
	if err := h.state.arch.Initialize(ctx); err != nil {
		return fmt.Errorf("unable to initialize archive: %w", err)
	}

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

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
	defaultChartBundleName string = "helm"
)

func (h *handler) Initialize(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	l.Info("initializing archive...")

	h.state.chartPath = filepath.Join(h.cfg.BundleDir, defaultChartBundleName)
	_, err = os.Stat(h.state.chartPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("bundled chart was not found: %w", err)
		}

		return fmt.Errorf("error checking chart after initializing: %w", err)
	}

	return nil
}

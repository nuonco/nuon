package check

import (
	"context"

	pkgctx "github.com/nuonco/nuon/bins/runner/internal/pkg/ctx"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

const (
	defaultChartPackageFilename string = "chart.tgz"
)

func (h *handler) Initialize(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	// initialize empty state
	h.state = &handlerState{cfg: &HealthcheckConfig{}}
	l.Info("initializing...")
	return nil
}

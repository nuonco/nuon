package helm

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon-runner-go/models"
	"go.uber.org/zap"

	pkgctx "github.com/powertoolsdev/mono/bins/runner/internal/pkg/ctx"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/plan"
)

func (h *handler) Validate(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	cfg, err := plan.ParseConfig[WaypointConfig](h.state.plan)
	if err != nil {
		return fmt.Errorf("unable to parse plan: %w", err)
	}

	h.state.cfg = &cfg.App.Deploy.Use
	l.Error("config", zap.Any("cfg", h.state.cfg))
	if h.state.cfg.Namespace == "" {
		return fmt.Errorf("namespace must be set on the input config")
	}

	return nil
}

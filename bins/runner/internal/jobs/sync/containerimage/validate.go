package containerimage

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon-runner-go/models"

	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/plan"
)

func (h *handler) Validate(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	cfg, err := plan.ParseConfig[WaypointConfig](h.state.plan)
	if err != nil {
		return fmt.Errorf("unable to parse plan: %w", err)
	}

	h.state.cfg = &cfg.Build.Use
	h.state.regCfg = &cfg.Build.Registry

	return nil
}

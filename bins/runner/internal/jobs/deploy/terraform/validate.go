package terraform

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

	h.state.cfg = &cfg.App.Deploy.Use
	h.state.srcCfg = cfg.App.Deploy.Use.ArtifactRepo
	h.state.srcTag = cfg.App.Deploy.Use.ArtifactTag

	return nil
}

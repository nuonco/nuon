package helm

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon-runner-go/models"
	"go.uber.org/zap"

	pkgctx "github.com/powertoolsdev/mono/bins/runner/internal/pkg/ctx"
	ociarchive "github.com/powertoolsdev/mono/bins/runner/internal/pkg/oci/archive"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/plan"
)

const (
	defaultNamespace string = "default"
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
	h.state.srcCfg = cfg.App.Deploy.Use.ArtifactRepo
	h.state.srcTag = cfg.App.Deploy.Use.ArtifactTag
	arch := ociarchive.New()
	if err := arch.Initialize(ctx); err != nil {
		return fmt.Errorf("unable to initialize archive: %w", err)
	}
	h.state.arch = arch

	if h.state.cfg.Namespace == "" {
		l.Info("no namespace set, using default", zap.String("namespace", defaultNamespace))
		h.state.cfg.Namespace = defaultNamespace
	}

	return nil
}

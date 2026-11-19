package controlplanejob

import (
	"context"

	runnercontrolplane "github.com/nuonco/nuon/pkg/runner/controlplane"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (a *Activities) newExecutor(_ context.Context, _ *app.RunnerJob) (*runnercontrolplane.Executor, error) {
	return runnercontrolplane.NewExecutor(a, a.l, runnercontrolplane.Config{GitRef: a.cfg.GitRef})
}

package workflow

import (
	"context"

	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"go.uber.org/zap"

	nuonrunner "github.com/nuonco/nuon/sdks/nuon-runner-go"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"

	"github.com/nuonco/nuon/bins/runner/internal/pkg/launcher"
	"github.com/nuonco/nuon/pkg/runner/jobs"
	"github.com/nuonco/nuon/pkg/runner/settings"
	"github.com/nuonco/nuon/pkg/runner/workspace"
)

type handler struct {
	v         *validator.Validate
	apiClient nuonrunner.Client
	settings  *settings.Settings

	launcher launcher.Launcher

	state *handlerState
}

var _ jobs.JobHandler = (*handler)(nil)

type HandlerParams struct {
	fx.In

	V         *validator.Validate
	APIClient nuonrunner.Client
	Settings  *settings.Settings
	Launcher  launcher.Launcher `optional:"true"`
}

func New(params HandlerParams) *handler {
	return &handler{
		apiClient: params.APIClient,
		v:         params.V,
		settings:  params.Settings,
		launcher:  params.Launcher,
	}
}

func (h *handler) GracefulShutdown(ctx context.Context, job *models.AppRunnerJob, l *zap.Logger) error {
	return nil
}

// why: workspaceRoot returns the directory the job's workspace is created under. A
// launcher is only wired for the image-actions handler, which mng runs natively
// on the VM host, so that path gets the root volume instead of the host's
// RAM-backed /tmp. The in-process handler keeps the default, which resolves
// inside the runner container's own filesystem.
//
// The preferred root is not always writable (a developer machine, or an mng unit
// whose sandboxing leaves /opt read-only), so the choice is resolved per job and
// reported rather than silently degrading to a memory-backed directory.
func (h *handler) workspaceRoot(l *zap.Logger) string {
	if h.launcher == nil {
		return workspace.DefaultTmpRootDir
	}
	return workspace.ResolveHostActionRoot(l)
}

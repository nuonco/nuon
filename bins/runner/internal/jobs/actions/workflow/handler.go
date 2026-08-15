package workflow

import (
	"context"

	"github.com/go-playground/validator/v10"
	nuonrunner "github.com/nuonco/nuon/sdks/nuon-runner-go"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/bins/runner/internal/pkg/launcher"
	"github.com/nuonco/nuon/pkg/runner/jobs"
	"github.com/nuonco/nuon/pkg/runner/settings"
	"github.com/nuonco/nuon/pkg/runner/workspace"
)

type handler struct {
	v         *validator.Validate
	apiClient nuonrunner.Client
	settings  *settings.Settings

	// launcher is only set for the image-actions handler registered by the mng
	// process; it is nil for the in-process actions handler.
	launcher launcher.Launcher

	// state is reused between function calls, but can _not_ be reused with different jobs.
	//
	// the job loop ensures that no handler ever has more than one job at a time, but this guarantee should be made
	// stronger in the future.
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

// workspaceRoot returns the directory the job's workspace is created under. A
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

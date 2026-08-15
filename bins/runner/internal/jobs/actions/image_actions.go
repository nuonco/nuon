package actions

import (
	"context"
	"time"

	"go.uber.org/fx"
	"go.uber.org/zap"

	workflow "github.com/nuonco/nuon/bins/runner/internal/jobs/actions/workflow"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/jobloop"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/launcher"
	"github.com/nuonco/nuon/pkg/runner/jobs"
	"github.com/nuonco/nuon/pkg/runner/workspace"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

const (
	imageActionsJobGroup models.AppRunnerJobGroup = models.AppRunnerJobGroupImageDashActions

	// imageCollectionTimeout bounds a collection pass so a slow docker daemon
	// can't keep the job loop from claiming its next action.
	imageCollectionTimeout = 2 * time.Minute
)

type ImageActionJobLoopParams struct {
	jobloop.BaseParams

	Handlers   []jobs.JobHandler `group:"image-actions"`
	ImageCache *launcher.ImageCache
}

func NewImageActionJobLoop(params ImageActionJobLoopParams) jobloop.JobLoop {
	return jobloop.New(params.Handlers, imageActionsJobGroup, params.BaseParams,
		// Action images are retained for reuse, so something has to bound them.
		// Running collection from the idle hook is what keeps it from removing an
		// image between a job's pull and its first container: the loop runs a
		// single worker goroutine, so no job of its own can be in flight here.
		// Leases are what make that safe against other processes.
		//
		// Time-boxed because this goroutine is not claiming work while it runs.
		jobloop.WithIdleHook(func(ctx context.Context) {
			ctx, cancel := context.WithTimeout(ctx, imageCollectionTimeout)
			defer cancel()

			params.ImageCache.CollectGarbage(ctx, params.L)
		}),
	)
}

// GetImageActionJobs wires the image-actions job loop and its docker launcher.
// It is registered ONLY by the mng process (which runs natively on the VM host
// with docker access) — that registration is what gates image-backed actions
// to VM runners.
func GetImageActionJobs() []fx.Option {
	return []fx.Option{
		fx.Provide(launcher.NewImageCache),
		fx.Provide(fx.Annotate(launcher.NewDockerLauncher, fx.As(new(launcher.Launcher)))),
		fx.Provide(jobloop.AsJobLoop(NewImageActionJobLoop)),
		fx.Provide(jobs.AsJobHandler("image-actions", workflow.New)),
		fx.Invoke(fx.Annotate(prepareActionHost, fx.ParamTags(`name:"system"`))),
	}
}

// prepareActionHost readies the VM host for image-backed actions. Action
// workspaces live on the root volume rather than the host's tmpfs /tmp, so this
// creates that root, clears anything a previous process left in it, and reports
// a filesystem that would put action content in RAM instead of on disk.
//
// Everything here is best effort: the process also runs management jobs, so a
// host that can't host action workspaces must still start.
func prepareActionHost(l *zap.Logger) {
	// Resolving here rather than waiting for the first job means a host that
	// can't use the preferred root says so at startup, and it logs which root
	// jobs will actually land in.
	root := workspace.ResolveHostActionRoot(l)
	l.Info("image-backed action workspaces will use", zap.String("path", root))

	// Every root is swept, not just the resolved one: a previous process may
	// have been able to write a different one.
	for _, candidate := range workspace.HostActionRoots() {
		removed, err := workspace.SweepStale(candidate)
		if err != nil {
			l.Warn("unable to sweep stale action workspaces", zap.String("path", candidate), zap.Error(err))
		}
		if len(removed) > 0 {
			l.Info("removed action workspaces left by a previous process",
				zap.String("path", candidate),
				zap.Strings("workspaces", removed),
			)
		}
	}
}

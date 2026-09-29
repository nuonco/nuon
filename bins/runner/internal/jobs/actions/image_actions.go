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

	imageCollectionTimeout = 2 * time.Minute
)

type ImageActionJobLoopParams struct {
	jobloop.BaseParams

	Handlers   []jobs.JobHandler `group:"image-actions"`
	ImageCache *launcher.ImageCache
}

func NewImageActionJobLoop(params ImageActionJobLoopParams) jobloop.JobLoop {
	return jobloop.New(params.Handlers, imageActionsJobGroup, params.BaseParams,
		// why: Action images are retained for reuse, so something has to bound them.
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

func GetImageActionJobs() []fx.Option {
	return []fx.Option{
		fx.Provide(launcher.NewImageCache),
		fx.Provide(fx.Annotate(launcher.NewDockerLauncher, fx.As(new(launcher.Launcher)))),
		fx.Provide(jobloop.AsJobLoop(NewImageActionJobLoop)),
		fx.Provide(jobs.AsJobHandler("image-actions", workflow.New)),
		fx.Invoke(fx.Annotate(prepareActionHost, fx.ParamTags(`name:"system"`))),
	}
}

func prepareActionHost(l *zap.Logger) {
	root := workspace.ResolveHostActionRoot(l)
	l.Info("image-backed action workspaces will use", zap.String("path", root))

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

package actions

import (
	"go.uber.org/fx"

	workflow "github.com/nuonco/nuon/bins/runner/internal/jobs/actions/workflow"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/jobloop"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/launcher"
	"github.com/nuonco/nuon/pkg/runner/jobs"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

const (
	imageActionsJobGroup models.AppRunnerJobGroup = models.AppRunnerJobGroupImageDashActions
)

type ImageActionJobLoopParams struct {
	jobloop.BaseParams

	Handlers []jobs.JobHandler `group:"image-actions"`
}

func NewImageActionJobLoop(params ImageActionJobLoopParams) jobloop.JobLoop {
	return jobloop.New(params.Handlers, imageActionsJobGroup, params.BaseParams)
}

// GetImageActionJobs wires the image-actions job loop and its docker launcher.
// It is registered ONLY by the mng process (which runs natively on the VM host
// with docker access) — that registration is what gates image-backed actions
// to VM runners.
func GetImageActionJobs() []fx.Option {
	return []fx.Option{
		fx.Provide(fx.Annotate(launcher.NewDockerLauncher, fx.As(new(launcher.Launcher)))),
		fx.Provide(jobloop.AsJobLoop(NewImageActionJobLoop)),
		fx.Provide(jobs.AsJobHandler("image-actions", workflow.New)),
	}
}

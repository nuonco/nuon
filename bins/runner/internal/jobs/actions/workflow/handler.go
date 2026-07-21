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

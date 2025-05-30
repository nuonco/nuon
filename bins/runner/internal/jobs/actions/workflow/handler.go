package workflow

import (
	"context"

	"github.com/go-playground/validator/v10"
	nuonrunner "github.com/nuonco/nuon-runner-go"
	"github.com/nuonco/nuon-runner-go/models"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/bins/runner/internal/jobs"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/settings"
)

type handler struct {
	v         *validator.Validate
	apiClient nuonrunner.Client
	settings  *settings.Settings

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
}

func New(params HandlerParams) *handler {
	return &handler{
		apiClient: params.APIClient,
		v:         params.V,
		settings:  params.Settings,
	}
}

func (h *handler) GracefulShutdown(ctx context.Context, job *models.AppRunnerJob, l *zap.Logger) error {
	return nil
}

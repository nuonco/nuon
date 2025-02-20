package update

import (
	"github.com/go-playground/validator/v10"
	nuonrunner "github.com/nuonco/nuon-runner-go"
	"go.uber.org/fx"

	"github.com/powertoolsdev/mono/bins/runner/internal"
	"github.com/powertoolsdev/mono/bins/runner/internal/jobs"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/errs"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/settings"
)

type handler struct {
	v           *validator.Validate
	apiClient   nuonrunner.Client
	settings    *settings.Settings
	shutdowner  fx.Shutdowner
	errRecorder *errs.Recorder
	cfg         *internal.Config
	state       *handlerState
}

var _ jobs.JobHandler = (*handler)(nil)

type HandlerParams struct {
	fx.In

	V           *validator.Validate
	APIClient   nuonrunner.Client
	Settings    *settings.Settings
	Shutdowner  fx.Shutdowner
	ErrRecorder *errs.Recorder
	Config      *internal.Config
}

func New(params HandlerParams) *handler {
	return &handler{
		apiClient:   params.APIClient,
		v:           params.V,
		settings:    params.Settings,
		shutdowner:  params.Shutdowner,
		errRecorder: params.ErrRecorder,
		cfg:         params.Config,
	}
}

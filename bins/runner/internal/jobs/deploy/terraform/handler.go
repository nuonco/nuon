package terraform

import (
	"github.com/go-playground/validator/v10"
	nuonrunner "github.com/nuonco/nuon/sdks/nuon-runner-go"
	"go.uber.org/fx"
	"go.uber.org/zap"

	runnerconfig "github.com/nuonco/nuon/pkg/runner/config"
	"github.com/nuonco/nuon/pkg/runner/errs"
	"github.com/nuonco/nuon/pkg/runner/jobs"
)

// handler is the handler implementation
type handler struct {
	v           *validator.Validate
	apiClient   nuonrunner.Client
	errRecorder *errs.Recorder
	cfg         *runnerconfig.Config
	l           *zap.Logger

	// created on initialization of the plugin struct
	state *handlerState
}

var _ jobs.JobHandler = (*handler)(nil)

type HandlerParams struct {
	fx.In

	V           *validator.Validate
	APIClient   nuonrunner.Client
	Config      *runnerconfig.Config
	ErrRecorder *errs.Recorder
	L           *zap.Logger
}

func New(params HandlerParams) (*handler, error) {
	return &handler{
		v:           params.V,
		apiClient:   params.APIClient,
		cfg:         params.Config,
		errRecorder: params.ErrRecorder,
		l:           params.L,
	}, nil
}

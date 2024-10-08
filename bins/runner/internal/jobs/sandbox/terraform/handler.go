package terraform

import (
	"log/slog"

	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/go-hclog"
	nuonrunner "github.com/nuonco/nuon-runner-go"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/bins/runner/internal"
	"github.com/powertoolsdev/mono/bins/runner/internal/jobs"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/errs"
)

// handler is the handler implementation
type handler struct {
	v           *validator.Validate
	apiClient   nuonrunner.Client
	errRecorder *errs.Recorder
	cfg         *internal.Config
	log         *zap.Logger
	hclog       hclog.Logger

	// created on initialization of the plugin struct
	state *handlerState
}

var _ jobs.JobHandler = (*handler)(nil)

type HandlerParams struct {
	fx.In

	V           *validator.Validate
	APIClient   nuonrunner.Client
	Config      *internal.Config
	ErrRecorder *errs.Recorder
	Log         *zap.Logger
	HCLog       hclog.Logger
	SLog        *slog.Logger `name:"system"`
}

func New(params HandlerParams) (*handler, error) {
	return &handler{
		v:           params.V,
		apiClient:   params.APIClient,
		cfg:         params.Config,
		log:         params.Log,
		hclog:       params.HCLog,
		errRecorder: params.ErrRecorder,
	}, nil
}

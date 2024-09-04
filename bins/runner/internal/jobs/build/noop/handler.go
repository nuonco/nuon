package noop

import (
	"github.com/go-playground/validator/v10"
	nuonrunner "github.com/nuonco/nuon-runner-go"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/bins/runner/internal"
	"github.com/powertoolsdev/mono/bins/runner/internal/jobs"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/errs"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/workspace"
	"github.com/powertoolsdev/mono/pkg/plugins/configs"
	planv1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/plan/v1"
)

type InputConfig configs.App[configs.Build[configs.NoopBuild, configs.NoopRegistry], configs.NoopDeploy]

type handlerState struct {
	// state for an individual run, that can not be reused
	plan      *planv1.Plan
	cfg       *InputConfig
	workspace workspace.Workspace
}

type handler struct {
	v           *validator.Validate
	apiClient   nuonrunner.Client
	errRecorder *errs.Recorder
	cfg         *internal.Config
	log         *zap.Logger
}

var _ jobs.JobHandler = (*handler)(nil)

type HandlerParams struct {
	fx.In

	V         *validator.Validate
	APIClient nuonrunner.Client
	Config    *internal.Config
	Log       *zap.Logger
}

func New(params HandlerParams) (*handler, error) {
	return &handler{
		v:         params.V,
		apiClient: params.APIClient,
		cfg:       params.Config,
		log:       params.Log,
	}, nil
}

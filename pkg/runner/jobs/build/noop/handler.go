package noop

import (
	"context"

	"github.com/go-playground/validator/v10"
	nuonrunner "github.com/nuonco/nuon/sdks/nuon-runner-go"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/plugins/configs"
	runnerconfig "github.com/nuonco/nuon/pkg/runner/config"
	"github.com/nuonco/nuon/pkg/runner/errs"
	"github.com/nuonco/nuon/pkg/runner/jobs"
	"github.com/nuonco/nuon/pkg/runner/workspace"
)

type InputConfig configs.App[configs.Build[configs.NoopBuild, configs.NoopRegistry], configs.NoopDeploy]

type handlerState struct {
	// state for an individual run, that can not be reused
	cfg       *InputConfig
	workspace workspace.Workspace
}

type handler struct {
	v           *validator.Validate
	apiClient   nuonrunner.Client
	errRecorder *errs.Recorder
	cfg         *runnerconfig.Config
}

var _ jobs.JobHandler = (*handler)(nil)

type HandlerParams struct {
	fx.In

	V           *validator.Validate
	APIClient   nuonrunner.Client
	Config      *runnerconfig.Config
	ErrRecorder *errs.Recorder
}

func New(params HandlerParams) (*handler, error) {
	return &handler{
		v:           params.V,
		apiClient:   params.APIClient,
		cfg:         params.Config,
		errRecorder: params.ErrRecorder,
	}, nil
}

func (h *handler) GracefulShutdown(ctx context.Context, job *models.AppRunnerJob, l *zap.Logger) error {
	return nil
}

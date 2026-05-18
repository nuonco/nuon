package sandbox

import (
	"context"

	"github.com/go-playground/validator/v10"
	nuonrunner "github.com/nuonco/nuon/sdks/nuon-runner-go"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
	"go.uber.org/fx"
	"go.uber.org/zap"

	runnerconfig "github.com/nuonco/nuon/pkg/runner/config"
	"github.com/nuonco/nuon/pkg/runner/errs"
	"github.com/nuonco/nuon/pkg/runner/jobs"
	ocicopy "github.com/nuonco/nuon/pkg/runner/oci/copy"
)

type handler struct {
	v           *validator.Validate
	apiClient   nuonrunner.Client
	errRecorder *errs.Recorder
	cfg         *runnerconfig.Config
	ociCopy     ocicopy.Copier

	// state is populated per-job and must not be shared across jobs.
	state *handlerState
}

var _ jobs.JobHandler = (*handler)(nil)

type HandlerParams struct {
	fx.In

	V           *validator.Validate
	APIClient   nuonrunner.Client
	Config      *runnerconfig.Config
	ErrRecorder *errs.Recorder
	OCICopy     ocicopy.Copier
}

func New(params HandlerParams) (*handler, error) {
	return &handler{
		v:           params.V,
		apiClient:   params.APIClient,
		cfg:         params.Config,
		errRecorder: params.ErrRecorder,
		ociCopy:     params.OCICopy,
	}, nil
}

func (h *handler) Name() string {
	return "sandbox-build"
}

func (h *handler) JobType() models.AppRunnerJobType {
	return models.AppRunnerJobType("sandbox-build")
}

func (h *handler) JobStatus() models.AppRunnerJobStatus {
	return models.AppRunnerJobStatusAvailable
}

func (h *handler) GracefulShutdown(ctx context.Context, job *models.AppRunnerJob, l *zap.Logger) error {
	return nil
}

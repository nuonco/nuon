package vmshutdown

import (
	"context"
	"time"

	"github.com/go-playground/validator/v10"
	nuonrunner "github.com/nuonco/nuon/sdks/nuon-runner-go"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/bins/runner/internal/jobs"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/errs"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/jobloop"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/settings"
)

const drainTimeout = 60 * time.Minute

type handler struct {
	v           *validator.Validate
	apiClient   nuonrunner.Client
	settings    *settings.Settings
	errRecorder *errs.Recorder
	jobLoops    []jobloop.JobLoop
}

var _ jobs.JobHandler = (*handler)(nil)

type HandlerParams struct {
	fx.In

	V           *validator.Validate
	APIClient   nuonrunner.Client
	Settings    *settings.Settings
	ErrRecorder *errs.Recorder
	JobLoops    []jobloop.JobLoop `group:"job_loops"`
}

func New(params HandlerParams) *handler {
	return &handler{
		apiClient:   params.APIClient,
		v:           params.V,
		settings:    params.Settings,
		errRecorder: params.ErrRecorder,
		jobLoops:    params.JobLoops,
	}
}

func (h *handler) GracefulShutdown(ctx context.Context, job *models.AppRunnerJob, l *zap.Logger) error {
	return nil
}

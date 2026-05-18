package imagemetadata

import (
	"context"

	"github.com/go-playground/validator/v10"
	nuonrunner "github.com/nuonco/nuon/sdks/nuon-runner-go"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/oci/metadata"
	runnerconfig "github.com/nuonco/nuon/pkg/runner/config"
	"github.com/nuonco/nuon/pkg/runner/errs"
	"github.com/nuonco/nuon/pkg/runner/jobs"
)

type handler struct {
	v           *validator.Validate
	apiClient   nuonrunner.Client
	errRecorder *errs.Recorder
	cfg         *runnerconfig.Config

	state *handlerState
}

var _ jobs.JobHandler = (*handler)(nil)

type HandlerParams struct {
	fx.In

	V           *validator.Validate
	APIClient   nuonrunner.Client
	ErrRecorder *errs.Recorder
	Config      *runnerconfig.Config
}

func New(params HandlerParams) (*handler, error) {
	return &handler{
		v:           params.V,
		apiClient:   params.APIClient,
		errRecorder: params.ErrRecorder,
		cfg:         params.Config,
	}, nil
}

func (h *handler) GracefulShutdown(ctx context.Context, job *models.AppRunnerJob, l *zap.Logger) error {
	return nil
}

type handlerState struct {
	plan     *FetchImageMetadataPlan
	metadata *metadata.ImageMetadata

	jobID          string
	jobExecutionID string
}

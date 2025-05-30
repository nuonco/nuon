package docker

import (
	"context"

	"github.com/go-playground/validator/v10"
	nuonrunner "github.com/nuonco/nuon-runner-go"
	"github.com/nuonco/nuon-runner-go/models"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/bins/runner/internal"
	"github.com/powertoolsdev/mono/bins/runner/internal/jobs"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/errs"
	ocicopy "github.com/powertoolsdev/mono/bins/runner/internal/pkg/oci/copy"
)

type handler struct {
	v           *validator.Validate
	apiClient   nuonrunner.Client
	errRecorder *errs.Recorder
	cfg         *internal.Config
	ociCopy     ocicopy.Copier

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
	Config    *internal.Config
	OCICopy   ocicopy.Copier
}

func New(params HandlerParams) (*handler, error) {
	return &handler{
		v:         params.V,
		apiClient: params.APIClient,
		cfg:       params.Config,
		ociCopy:   params.OCICopy,
	}, nil
}

func (h *handler) GracefulShutdown(ctx context.Context, job *models.AppRunnerJob, l *zap.Logger) error {
	return nil
}

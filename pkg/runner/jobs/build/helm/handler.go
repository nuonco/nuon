package helm

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/afero"
	"go.uber.org/fx"
	"go.uber.org/zap"

	nuonrunner "github.com/nuonco/nuon/sdks/nuon-runner-go"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"

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
	fs          afero.Fs

	// state is reused between function calls, but can _not_ be reused with different jobs.
	//
	// the job loop ensures that no handler ever has more than one job at a time, but this guarantee should be made
	// stronger in the future.
	state *handlerState
}

var _ jobs.JobHandler = (*handler)(nil)

type HandlerParams struct {
	fx.In

	V           *validator.Validate
	APIClient   nuonrunner.Client
	Config      *runnerconfig.Config
	OCICopy     ocicopy.Copier
	ErrRecorder *errs.Recorder
	FS          afero.Fs `optional:"true"`
}

func New(params HandlerParams) (*handler, error) {
	if params.FS == nil {
		params.FS = afero.NewOsFs()
	}
	switch params.FS.(type) {
	case *afero.OsFs, *afero.MemMapFs:
	default:
		return nil, fmt.Errorf("unsupported helm build filesystem %T", params.FS)
	}
	return &handler{
		v:           params.V,
		apiClient:   params.APIClient,
		cfg:         params.Config,
		ociCopy:     params.OCICopy,
		errRecorder: params.ErrRecorder,
		fs:          params.FS,
	}, nil
}

func (h *handler) SetAPIClient(c nuonrunner.Client) nuonrunner.Client {
	prev := h.apiClient
	h.apiClient = c
	return prev
}

func (h *handler) GracefulShutdown(ctx context.Context, job *models.AppRunnerJob, l *zap.Logger) error {
	return nil
}

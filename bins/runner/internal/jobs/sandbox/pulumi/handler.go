package pulumi

import (
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"

	nuonrunner "github.com/nuonco/nuon/sdks/nuon-runner-go"

	runnerconfig "github.com/nuonco/nuon/pkg/runner/config"
	"github.com/nuonco/nuon/pkg/runner/errs"
	"github.com/nuonco/nuon/pkg/runner/jobs"
	ociarchive "github.com/nuonco/nuon/pkg/runner/oci/archive"
)

type handler struct {
	v             *validator.Validate
	apiClient     nuonrunner.Client
	errRecorder   *errs.Recorder
	cfg           *runnerconfig.Config
	archiveSource ociarchive.Source

	state *handlerState
}

var _ jobs.JobHandler = (*handler)(nil)

type HandlerParams struct {
	fx.In

	V           *validator.Validate
	APIClient   nuonrunner.Client
	Config      *runnerconfig.Config
	ErrRecorder *errs.Recorder
	// ArchiveSource is only provided by air-gapped runs, where OCI sources
	// must be served from the bundle instead of a registry.
	ArchiveSource ociarchive.Source `optional:"true"`
}

func New(params HandlerParams) (*handler, error) {
	return &handler{
		v:             params.V,
		apiClient:     params.APIClient,
		cfg:           params.Config,
		errRecorder:   params.ErrRecorder,
		archiveSource: params.ArchiveSource,
	}, nil
}

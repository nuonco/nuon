package ociprune

import (
	"context"

	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"go.uber.org/zap"

	nuonrunner "github.com/nuonco/nuon/sdks/nuon-runner-go"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"

	runnerconfig "github.com/nuonco/nuon/pkg/runner/config"
	"github.com/nuonco/nuon/pkg/runner/errs"
	"github.com/nuonco/nuon/pkg/runner/jobs"
	"github.com/nuonco/nuon/pkg/runner/oci/prune"
)

type handler struct {
	v           *validator.Validate
	apiClient   nuonrunner.Client
	errRecorder *errs.Recorder
	cfg         *runnerconfig.Config

	openRegistry openRegistryFn
	scanCluster  scanClusterFn

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

		openRegistry: openECRRegistry,
		scanCluster:  prune.ScanCluster,
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

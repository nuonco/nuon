package handler

import (
	"github.com/go-playground/validator/v10"
	nuonrunner "github.com/nuonco/nuon-runner-go"
	"go.uber.org/fx"

	"github.com/powertoolsdev/mono/bins/runner/internal/jobs"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/settings"
)

type handler struct {
	v         *validator.Validate
	apiClient nuonrunner.Client
	settings  *settings.Settings
}

var _ jobs.JobHandler = (*handler)(nil)

type Params struct {
	fx.In

	V         *validator.Validate
	APIClient nuonrunner.Client
	Settings  *settings.Settings
}

func New(params Params) *handler {
	return &handler{
		apiClient: params.APIClient,
		v:         params.V,
		settings:  params.Settings,
	}
}

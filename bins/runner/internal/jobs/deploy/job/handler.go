package job

import (
	"github.com/go-playground/validator/v10"
	nuonrunner "github.com/nuonco/nuon-runner-go"
	"github.com/nuonco/nuon-runner-go/models"
	"go.uber.org/fx"

	"github.com/powertoolsdev/mono/bins/runner/internal"
	"github.com/powertoolsdev/mono/bins/runner/internal/jobs"
	"github.com/powertoolsdev/mono/pkg/plugins/configs"
)

const (
	runnerJobGroup models.AppRunnerJobGroup = models.AppRunnerJobGroupDeploy
)

type handler struct {
	v *validator.Validate

	// internal fields
	Cfg configs.JobDeploy `validate:"required"`
}

var _ jobs.JobHandler = (*handler)(nil)

type HandlerParams struct {
        fx.In

	V         *validator.Validate
	APIClient nuonrunner.Client
	Config    *internal.Config
}

func New(params HandlerParams) (*handler, error) {
	return &handler{
		v: params.V,
	}, nil
}

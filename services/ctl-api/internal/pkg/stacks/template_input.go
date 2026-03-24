package stacks

import (
	"github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type TemplateInput struct {
	Install                    *app.Install             `validate:"required"`
	CloudFormationStackVersion *app.InstallStackVersion `validate:"required"`
	InstallState               *state.State             `validate:"required"`
	AppCfg                     *app.AppConfig           `validate:"required"`

	Runner   *app.Runner              `validate:"required"`
	Settings *app.RunnerGroupSettings `validate:"required"`
	APIToken string                   `validate:"required"`

	// subscripts and embedded templates
	RunnerInitScriptURL string `validate:"required"`
	PhonehomeScript     string `validate:"required"`

	// AWS Embeds
	VPCNestedStackTemplateURL    string `validate:"required"`
	RunnerNestedStackTemplateURL string `validate:"required"`
}

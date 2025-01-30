package actions

import (
	"github.com/go-playground/validator/v10"
	"github.com/nuonco/nuon-go"
	"github.com/powertoolsdev/mono/bins/cli/internal/config"
	"github.com/pterm/pterm"
)

type Service struct {
	v   *validator.Validate
	api nuon.Client
	cfg *config.Config
}

func New(v *validator.Validate, apiClient nuon.Client, cfg *config.Config) *Service {
	return &Service{
		v:   v,
		api: apiClient,
		cfg: cfg,
	}
}

func (s *Service) printAppNotSetMsg() {
	pterm.DefaultBasicText.Printfln("current app is not set, use %s to set one", pterm.LightMagenta("apps select"))
}

package actions

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/nuonco/nuon-go"
	"github.com/powertoolsdev/mono/bins/cli/internal/config"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui/bubbles"
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
	fmt.Printf("%s\n", bubbles.BaseStyle.Render("current app is not set, use apps select to set one"))
}

package builds

import (
	"github.com/nuonco/nuon-go"
	"github.com/powertoolsdev/mono/bins/cli/internal/config"
	"github.com/pterm/pterm"
)

type Service struct {
	api nuon.Client
	cfg *config.Config
}

func New(apiClient nuon.Client, cfg *config.Config) *Service {
	return &Service{
		api: apiClient,
		cfg: cfg,
	}
}

func (s *Service) printAppNotSetMsg() {
	pterm.DefaultBasicText.Printfln("current app is not set, use %s to set one", pterm.LightMagenta("apps select"))
}

func (s *Service) printOrgNotSetMsg() {
	pterm.DefaultBasicText.Printfln("current org is not set, use %s to set one", pterm.LightMagenta("orgs select"))
}

package builds

import (
	"fmt"

	"github.com/nuonco/nuon-go"
	"github.com/powertoolsdev/mono/bins/cli/internal/config"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui/bubbles"
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
	fmt.Printf("%s\n", bubbles.BaseStyle.Render("current app is not set, use apps select to set one"))
}

func (s *Service) printOrgNotSetMsg() {
	fmt.Printf("%s\n", bubbles.BaseStyle.Render("current org is not set, use orgs select to set one"))
}

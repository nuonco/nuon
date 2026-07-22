package roles

import (
	"fmt"

	"github.com/nuonco/nuon/bins/cli/internal/config"
	"github.com/nuonco/nuon/bins/cli/internal/ui/bubbles"
	"github.com/nuonco/nuon/sdks/nuon-go"
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

func (s *Service) printOrgNotSetMsg() {
	fmt.Printf("%s\n", bubbles.BaseStyle.Render("current org is not set, use orgs select to set one"))
}

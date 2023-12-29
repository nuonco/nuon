package orgs

import (
	"context"
	"errors"
	"fmt"

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

func (s *Service) setOrgInConfig(ctx context.Context, orgID string) error {
	s.cfg.Set("org_id", orgID)
	return s.cfg.WriteConfig()
}

func (s *Service) printOrgSetMsg(name, id string) {
	pterm.Info.Printfln("current org is now %s: %s", pterm.Green(name), pterm.Green(id))
}

func (s *Service) printNoOrgsMsg() {
	pterm.DefaultBasicText.Printfln("you don't have any orgs, create one using %s", pterm.LightMagenta("orgs create"))
}

func (s *Service) printOrgNotFoundMsg(id string) {
	pterm.DefaultBasicText.Printfln("can't find org %s, use %s to view all orgs or create one using %s", pterm.Green(id), pterm.LightMagenta("orgs list"), pterm.LightMagenta("orgs create"))
}

func (s *Service) notFoundErr(id string) error {
	return errors.New(fmt.Sprintf("org %s was not found", id))
}

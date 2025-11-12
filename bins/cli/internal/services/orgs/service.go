package orgs

import (
	"context"
	"errors"
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

func (s *Service) setOrgID(ctx context.Context, orgID string) error {
	getCurrentOrgID := s.cfg.GetString("org_id")
	if getCurrentOrgID == orgID {
		return nil
	}

	err := s.unsetOrgID(ctx)
	if err != nil {
		return err
	}

	s.cfg.Set("org_id", orgID)
	return s.cfg.WriteConfig()
}

func (s *Service) unsetOrgID(ctx context.Context) error {
	s.cfg.Set("install_id", "")
	s.cfg.Set("app_id", "")
	s.cfg.Set("org_id", "")
	fmt.Printf("%s\n", bubbles.InfoStyle.Render("current org is now unset"))
	return s.cfg.WriteConfig()
}

func (s *Service) printOrgSetMsg(name, id string) {
	fmt.Printf("%s\n", bubbles.InfoStyle.Render(fmt.Sprintf("current org is now %s: %s", name, id)))
}

func (s *Service) printNoOrgsMsg() {
	fmt.Printf("%s\n", bubbles.BaseStyle.Render("you don't have any orgs, create one using orgs create"))
}

func (s *Service) printOrgNotFoundMsg(id string) {
	fmt.Printf("%s\n", bubbles.BaseStyle.Render(fmt.Sprintf("can't find org %s, use orgs list to view all orgs or create one using orgs create", id)))
}

func (s *Service) notFoundErr(id string) error {
	return errors.New(fmt.Sprintf("org %s was not found", id))
}

func (s *Service) printOrgNotSetMsg() {
	fmt.Printf("%s\n", bubbles.BaseStyle.Render("current org is not set, use orgs select to set one"))
}

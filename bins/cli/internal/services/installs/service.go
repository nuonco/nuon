package installs

import (
	"context"
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

func (s *Service) setInstallID(ctx context.Context, installID string) error {
	s.cfg.Set("install_id", installID)
	return s.cfg.WriteConfig()
}

func (s *Service) setAppID(ctx context.Context, appID string) error {
	s.cfg.Set("app_id", appID)
	return s.cfg.WriteConfig()
}

func (s *Service) GetInstallID() string {
	installID := s.cfg.GetString("install_id")
	if installID == "" {
		return ""
	}
	return installID
}

func (s *Service) unsetInstallID(ctx context.Context) error {
	s.cfg.Set("install_id", "")
	fmt.Printf("%s\n", bubbles.InfoStyle.Render("current install is now unset"))
	return s.cfg.WriteConfig()
}

func (s *Service) printAppSetMsg(id string) {
	fmt.Printf("%s\n", bubbles.InfoStyle.Render(fmt.Sprintf("current app is now %s", id)))
}

func (s *Service) printInstallSetMsg(name, id string) {
	fmt.Printf("%s\n", bubbles.InfoStyle.Render(fmt.Sprintf("current install is now %s: %s", name, id)))
}

func (s *Service) printNoInstallsMsg() {
	fmt.Printf("%s\n", bubbles.BaseStyle.Render("you don't have any installs, create one using installs create"))
}

func (s *Service) printInstallNotFoundMsg(id string) {
	fmt.Printf("%s\n", bubbles.BaseStyle.Render(fmt.Sprintf("can't find install %s, use installs list to view all installs or create one using installs create", id)))
}

func (s *Service) notFoundErr(id string) error {
	return fmt.Errorf("install %s was not found", id)
}

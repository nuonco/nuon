package installs

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

func (s *Service) setInstallID(ctx context.Context, installID string) error {
	s.cfg.Set("install_id", installID)
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
	pterm.Info.Printfln("current install is now %s", pterm.Green("unset"))
	return s.cfg.WriteConfig()
}

func (s *Service) printInstallSetMsg(name, id string) {
	pterm.Info.Printfln("current install is now %s: %s", pterm.Green(name), pterm.Green(id))
}

func (s *Service) printNoInstallsMsg() {
	pterm.DefaultBasicText.Printfln("you don't have any installs, create one using %s", pterm.LightMagenta("installs create"))
}

func (s *Service) printInstallNotFoundMsg(id string) {
	pterm.DefaultBasicText.Printfln("can't find install %s, use %s to view all installs or create one using %s", pterm.Green(id), pterm.LightMagenta("installs list"), pterm.LightMagenta("installs create"))
}

func (s *Service) notFoundErr(id string) error {
	return errors.New(fmt.Sprintf("install %s was not found", id))
}

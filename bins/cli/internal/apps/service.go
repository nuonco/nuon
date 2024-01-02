package apps

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

func (s *Service) setAppInConfig(ctx context.Context, appID string) error {
	s.cfg.Set("app_id", appID)
	return s.cfg.WriteConfig()
}

func (s *Service) printAppSetMsg(name, id string) {
	pterm.Info.Printfln("current app is now %s: %s", pterm.Green(name), pterm.Green(id))
}

func (s *Service) printNoAppsMsg() {
	pterm.DefaultBasicText.Printfln("you don't have any apps, create one using %s", pterm.LightMagenta("apps create"))
}

func (s *Service) printAppNotFoundMsg(id string) {
	pterm.DefaultBasicText.Printfln("can't find app %s, use %s to view all apps", pterm.Green(id), pterm.LightMagenta("apps list"))
}

func (s *Service) notFoundErr(id string) error {
	return errors.New(fmt.Sprintf("app %s was not found", id))
}

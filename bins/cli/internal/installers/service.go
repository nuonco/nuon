package installers

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

func (s *Service) setAppInConfig(ctx context.Context, appID string) error {
	s.cfg.Set("app_id", appID)
	return s.cfg.WriteConfig()
}

func (s *Service) printAppSetMsg(name, id string) {
	fmt.Printf("%s\n", bubbles.InfoStyle.Render(fmt.Sprintf("current app is now %s: %s", name, id)))
}

func (s *Service) printNoAppsMsg() {
	fmt.Printf("%s\n", bubbles.BaseStyle.Render("you don't have any apps, create one using apps create"))
}

func (s *Service) printAppNotFoundMsg(id string) {
	fmt.Printf("%s\n", bubbles.BaseStyle.Render(fmt.Sprintf("can't find app %s, use apps list to view all apps", id)))
}

func (s *Service) notFoundErr(id string) error {
	return errors.New(fmt.Sprintf("app %s was not found", id))
}

package apps

import (
	"context"
	"errors"
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

func (s *Service) getAppID() string {
	appID := s.cfg.GetString("app_id")
	if appID == "" {
		return ""
	}
	return appID
}

func (s *Service) setAppID(ctx context.Context, appID string) error {
	getCurrentAppID := s.cfg.GetString("app_id")
	if getCurrentAppID == appID {
		return nil
	}

	err := s.unsetAppID(ctx)
	if err != nil {
		return err
	}

	s.cfg.Set("app_id", appID)
	return s.cfg.WriteConfig()
}

func (s *Service) unsetAppID(ctx context.Context) error {
	// unset install_id
	s.cfg.Set("install_id", "")
	s.cfg.Set("app_id", "")
	fmt.Printf("%s\n", bubbles.InfoStyle.Render("current app is now unset"))
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

func (s *Service) printAppNotSetMsg() {
	fmt.Printf("%s\n", bubbles.BaseStyle.Render("current app is not set, use apps select to set one"))
}

func (s *Service) notFoundErr(id string) error {
	return errors.New(fmt.Sprintf("app %s was not found", id))
}

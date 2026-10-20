package actions

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"

	"github.com/nuonco/nuon/bins/cli/internal/config"
	"github.com/nuonco/nuon/bins/cli/internal/lookup"
	"github.com/nuonco/nuon/bins/cli/internal/ui/bubbles"
	"github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
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

func (s *Service) printAppNotSetMsg() {
	fmt.Printf("%s\n", bubbles.BaseStyle.Render("current app is not set, use apps select to set one"))
}

func (s *Service) installOrSelected(ctx context.Context, installIDOrName string) (*models.AppInstall, error) {
	if installIDOrName == "" {
		installIDOrName = s.cfg.GetString("install_id")
	}
	return lookup.Install(ctx, s.api, installIDOrName)
}

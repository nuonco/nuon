package runbooks

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nuonco/nuon/bins/cli/internal/config"
	runbooksui "github.com/nuonco/nuon/bins/cli/internal/ui/v3/runbooks"
	"github.com/nuonco/nuon/sdks/nuon-go"
)

type Service struct {
	api nuon.Client
	cfg *config.Config
}

func New(api nuon.Client, cfg *config.Config) *Service {
	return &Service{
		api: api,
		cfg: cfg,
	}
}

func (s *Service) RunbooksTUI(ctx context.Context, installID string, asJSON bool) error {
	if installID == "" {
		installID = s.cfg.GetString("install_id")
	}
	if installID == "" {
		return fmt.Errorf("install-id is required: use --install-id or select an install with 'nuon installs select'")
	}

	if asJSON {
		return s.runbooksJSON(ctx, installID)
	}

	runbooksui.App(ctx, s.cfg, s.api, installID)
	return nil
}

func (s *Service) runbooksJSON(ctx context.Context, installID string) error {
	runbooks, err := s.api.GetInstallRunbooks(ctx, installID)
	if err != nil {
		return fmt.Errorf("failed to get runbooks: %w", err)
	}

	data, err := json.MarshalIndent(runbooks, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal runbooks: %w", err)
	}

	fmt.Println(string(data))
	return nil
}

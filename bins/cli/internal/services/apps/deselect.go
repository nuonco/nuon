package apps

import (
	"context"

	"github.com/nuonco/nuon/bins/cli/internal/ui"
)

func (s *Service) Deselect(ctx context.Context, asJSON bool) error {
	hadInstall := s.cfg.GetString("install_id") != ""

	if err := s.unsetAppID(ctx); err != nil {
		return err
	}

	msg := "current app is now unset"
	if hadInstall {
		msg += "\nthe install has also been unset"
	}
	ui.PrintResult(asJSON, msg, map[string]string{
		"status":  "app_deselected",
		"message": "current app is now unset",
	})
	return nil
}

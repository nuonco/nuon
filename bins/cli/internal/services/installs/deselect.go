package installs

import (
	"context"

	"github.com/nuonco/nuon/bins/cli/internal/ui"
)

func (s *Service) Deselect(ctx context.Context, asJSON bool) error {

	if err := s.unsetInstallID(ctx); err != nil {
		return err
	}

	if asJSON {
		ui.PrintJSON(actionResult{Status: "install_deselected"})
		return nil
	}

	s.printInstallUnsetMsg()
	return nil
}

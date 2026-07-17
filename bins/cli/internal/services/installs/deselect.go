package installs

import (
	"context"

	"github.com/nuonco/nuon/bins/cli/internal/ui"
)

func (s *Service) Deselect(ctx context.Context, asJSON bool) error {
	view := ui.NewGetView()

	if err := s.unsetInstallID(ctx); err != nil {
		return view.Error(err)
	}

	if asJSON {
		ui.PrintJSON(actionResult{Status: "install_deselected"})
		return nil
	}

	s.printInstallUnsetMsg()
	return nil
}

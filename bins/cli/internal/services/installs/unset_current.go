package installs

import (
	"context"

	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) UnsetCurrent(ctx context.Context) error {
	view := ui.NewGetView()

	if err := s.unsetInstallID(ctx); err != nil {
		return view.Error(err)
	}

	return nil
}

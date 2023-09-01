package installs

import (
	"context"

	"github.com/powertoolsdev/mono/pkg/ui"
)

func (s *Service) Get(ctx context.Context, id string) error {
	install, err := s.api.GetInstall(ctx, id)
	if err != nil {
		return err
	}

	statusColor := ui.GetStatusColor(install.Status)
	ui.Line(ctx, "%s%s %s- %s - %s", statusColor, install.Status, ui.ColorReset, install.ID, install.Name)
	return nil
}

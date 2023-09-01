package orgs

import (
	"context"

	"github.com/powertoolsdev/mono/pkg/ui"
)

func (s *Service) Current(ctx context.Context) error {
	org, err := s.api.GetOrg(ctx)
	if err != nil {
		return err
	}

	statusColor := ui.GetStatusColor(org.Status)
	ui.Line(ctx, "%s%s %s- %s - %s", statusColor, org.Status, ui.ColorReset, org.ID, org.Name)
	return nil
}

package orgs

import (
	"context"

	"github.com/nuonco/nuon/bins/cli/internal/ui"
)

func (s *Service) Deselect(ctx context.Context, asJSON bool) error {
	if err := s.unsetOrgID(ctx); err != nil {
		return ui.PrintError(err)
	}

	ui.PrintResult(asJSON, "current org is now unset", map[string]string{"status": "org_deselected"})
	return nil
}

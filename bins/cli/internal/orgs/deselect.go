package orgs

import (
	"context"

	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) Deselect(ctx context.Context) error {
	view := ui.NewGetView()

	if err := s.unsetOrgID(ctx); err != nil {
		return view.Error(err)
	}

	return nil
}

package components

import (
	"context"

	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) Delete(ctx context.Context, compID string) {
	view := ui.NewDeleteView("component", compID)

	view.Start()
	_, err := s.api.DeleteComponent(ctx, compID)
	if err != nil {
		view.Fail(err)
		return
	}
	view.Success()
}

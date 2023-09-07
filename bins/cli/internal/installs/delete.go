package installs

import (
	"context"

	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) Delete(ctx context.Context, id string) {
	view := ui.NewDeleteView("install", id)

	view.Start()
	_, err := s.api.DeleteInstall(ctx, id)
	if err != nil {
		view.Fail(err)
		return
	}

	view.Success()
}

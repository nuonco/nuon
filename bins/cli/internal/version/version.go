package version

import (
	"context"

	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) Version(ctx context.Context) {
	view := ui.NewGetView()
	version := "development"
	view.Render([][]string{
		{"version", version},
	})
}

package version

import (
	"context"

	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

var Version string = "development"

func (s *Service) Version(ctx context.Context) {
	view := ui.NewGetView()
	view.Render([][]string{
		{"version", Version},
	})
}

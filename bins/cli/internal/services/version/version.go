package version

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

var Version string = "development"

func (s *Service) Version(ctx context.Context, asJSON bool) error {
	if asJSON {
		fmt.Printf("%s\n", Version)
		return nil
	}

	view := ui.NewGetView()
	view.Render([][]string{
		{"version", Version},
	})
	return nil
}

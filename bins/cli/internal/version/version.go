package version

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

var Version string = "development"

func (s *Service) Version(ctx context.Context, asJSON bool) {
	if asJSON {
		fmt.Printf("%s\n", Version)
		return
	}

	view := ui.NewGetView()
	view.Render([][]string{
		{"version", Version},
	})

}

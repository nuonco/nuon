package docs

import (
	"context"

	"github.com/pkg/browser"

	"github.com/powertoolsdev/mono/pkg/ui"
)

const (
	publicDocsSiteURL string = "https://docs.nuon.co"
)

func (s *Service) Browse(ctx context.Context, asJSON bool) error {
	ui.Line(ctx, "opening up docs")
	if asJSON {
		ui.Line(ctx, publicDocsSiteURL)
	} else {
		browser.OpenURL(publicDocsSiteURL)
	}

	return nil
}

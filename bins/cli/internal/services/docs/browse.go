package docs

import (
	"context"

	"github.com/pkg/browser"

	"github.com/nuonco/nuon/bins/cli/internal/ui"
)

const (
	publicDocsSiteURL string = "https://docs.nuon.co"
)

func (s *Service) Browse(ctx context.Context, asJSON bool) error {
	if asJSON {
		ui.PrintJSON(map[string]string{"url": publicDocsSiteURL})
		return nil
	}

	ui.PrintLn("opening up docs")
	browser.OpenURL(publicDocsSiteURL)
	return nil
}

package docs

import (
	"context"
	"fmt"
	"net/url"

	"github.com/pkg/browser"

	"github.com/nuonco/nuon/bins/cli/internal/ui"
)

func (s *Service) BrowseAPI(ctx context.Context, asJSON bool) error {
	params := url.Values{}
	params.Add("org_id", s.cfg.OrgID)
	params.Add("api_key", "Bearer "+s.cfg.APIToken)

	docsURL := fmt.Sprintf("%s/docs/index.html?%s", s.cfg.APIURL, params.Encode())

	if asJSON {
		ui.PrintJSON(map[string]string{"url": docsURL})
		return nil
	}

	ui.PrintLn("opening up api docs with local api-key and org-id preauthorized")
	browser.OpenURL(docsURL)
	return nil
}

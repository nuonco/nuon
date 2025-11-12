package docs

import (
	"context"
	"fmt"
	"net/url"

	"github.com/pkg/browser"

	"github.com/powertoolsdev/mono/pkg/ui"
)

func (s *Service) BrowseAPI(ctx context.Context, asJSON bool) error {
	ui.Line(ctx, "opening up api docs with local api-key and org-id preauthorized")

	params := url.Values{}
	params.Add("org_id", s.cfg.OrgID)
	params.Add("api_key", "Bearer "+s.cfg.APIToken)

	url := fmt.Sprintf("%s/docs/index.html?%s", s.cfg.APIURL, params.Encode())

	if asJSON {
		ui.Line(ctx, "%s", url)
	} else {
		browser.OpenURL(url)
	}

	return nil
}

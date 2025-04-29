package links

import (
	"net/url"

	"github.com/powertoolsdev/mono/services/ctl-api/internal"
)

func AppLinks(cfg *internal.Config, appID string) map[string]any {
	return map[string]any{
		"ui":               AppUILink(cfg, appID),
		"api":              AppAPILink(cfg, appID),
		"temporal_ui_link": AppAPILink(cfg, appID),
	}
}

func AppUILink(cfg *internal.Config, appID string) string {
	link, err := url.JoinPath(cfg.AppURL, "apps", appID)
	if err != nil {
		return handleErr(cfg, err)
	}

	return link
}

func AppTemporalUILink(cfg *internal.Config, appID string) string {
	return eventLoopLink(cfg, "apps", appID)
}

func AppAPILink(cfg *internal.Config, appID string) string {
	link, err := url.JoinPath(cfg.PublicAPIURL,
		"v1",
		"apps", appID)
	if err != nil {
		return handleErr(cfg, err)
	}

	return link
}

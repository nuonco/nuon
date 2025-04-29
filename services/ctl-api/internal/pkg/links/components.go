package links

import (
	"net/url"

	"github.com/powertoolsdev/mono/services/ctl-api/internal"
)

func ComponentLinks(cfg *internal.Config, componentID string) map[string]any {
	return map[string]any{
		"ui":               ComponentUILink(cfg, componentID),
		"api":              ComponentAPILink(cfg, componentID),
		"temporal_ui_link": ComponentAPILink(cfg, componentID),
	}
}

func ComponentUILink(cfg *internal.Config, componentID string) string {
	link, err := url.JoinPath(cfg.AppURL, "components", componentID)
	if err != nil {
		return handleErr(cfg, err)
	}

	return link
}

func ComponentTemporalUILink(cfg *internal.Config, componentID string) string {
	link, err := url.JoinPath(cfg.TemporalUIURL,
		"namespaces",
		"components",
		"workflows",
		"event-loop-"+componentID)
	if err != nil {
		return handleErr(cfg, err)
	}

	return link
}

func ComponentAPILink(cfg *internal.Config, componentID string) string {
	link, err := url.JoinPath(cfg.PublicAPIURL,
		"v1",
		"components", componentID)
	if err != nil {
		return handleErr(cfg, err)
	}

	return link
}

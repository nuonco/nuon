package links

import (
	"net/url"

	"github.com/powertoolsdev/mono/services/ctl-api/internal"
)

func ActionLinks(cfg *internal.Config, actionID string) map[string]any {
	return map[string]any{
		"ui":               ActionUILink(cfg, actionID),
		"api":              ActionAPILink(cfg, actionID),
		"temporal_ui_link": ActionAPILink(cfg, actionID),
	}
}

func ActionUILink(cfg *internal.Config, actionID string) string {
	link, err := url.JoinPath(cfg.AppURL, "actions", actionID)
	if err != nil {
		return handleErr(cfg, err)
	}

	return link
}

func ActionTemporalUILink(cfg *internal.Config, actionID string) string {
	return eventLoopLink(cfg, "actions", actionID)
}

func ActionAPILink(cfg *internal.Config, actionID string) string {
	link, err := url.JoinPath(cfg.PublicAPIURL,
		"v1",
		"actions", actionID)
	if err != nil {
		return handleErr(cfg, err)
	}

	return link
}

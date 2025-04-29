package links

import (
	"net/url"

	"github.com/powertoolsdev/mono/services/ctl-api/internal"
)

func InstallWorkflowLinks(cfg *internal.Config, iws string) map[string]any {
	return map[string]any{
		"ui":               InstallWorkflowUILink(cfg, iws),
		"api":              InstallWorkflowAPILink(cfg, iws),
		"temporal_ui_link": InstallWorkflowAPILink(cfg, iws),
	}
}

func InstallWorkflowUILink(cfg *internal.Config, appID string) string {
	link, err := url.JoinPath(cfg.TemporalUIURL, "apps", appID)
	if err != nil {
		return handleErr(cfg, err)
	}

	return link
}

func InstallWorkflowTemporalUILink(cfg *internal.Config, appID string) string {
	return eventLoopLink(cfg, "apps", appID)
}

func InstallWorkflowAPILink(cfg *internal.Config, appID string) string {
	link, err := url.JoinPath(cfg.PublicAPIURL,
		"v1",
		"apps", appID)
	if err != nil {
		return handleErr(cfg, err)
	}

	return link
}

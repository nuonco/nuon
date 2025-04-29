package links

import (
	"net/url"

	"github.com/powertoolsdev/mono/services/ctl-api/internal"
)

func InstallWorkflowStepLinks(cfg *internal.Config, iws string) map[string]any {
	return map[string]any{
		"ui":               InstallWorkflowStepUILink(cfg, iws),
		"api":              InstallWorkflowStepAPILink(cfg, iws),
		"temporal_ui_link": InstallWorkflowStepAPILink(cfg, iws),
	}
}

func InstallWorkflowStepUILink(cfg *internal.Config, appID string) string {
	link, err := url.JoinPath(cfg.TemporalUIURL, "apps", appID)
	if err != nil {
		return handleErr(cfg, err)
	}

	return link
}

func InstallWorkflowStepTemporalUILink(cfg *internal.Config, appID string) string {
	return eventLoopLink(cfg, "apps", appID)
}

func InstallWorkflowStepAPILink(cfg *internal.Config, appID string) string {
	link, err := url.JoinPath(cfg.PublicAPIURL,
		"v1",
		"apps", appID)
	if err != nil {
		return handleErr(cfg, err)
	}

	return link
}

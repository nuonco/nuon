package links

import (
	"net/url"

	"github.com/powertoolsdev/mono/services/ctl-api/internal"
)

func OrgLinks(cfg *internal.Config, orgID string) map[string]any {
	return map[string]any{
		"ui":               OrgUILink(cfg, orgID),
		"api":              OrgAPILink(cfg, orgID),
		"temporal_ui_link": OrgAPILink(cfg, orgID),
	}
}

func OrgUILink(cfg *internal.Config, orgID string) string {
	link, err := url.JoinPath(cfg.AppURL, "orgs", orgID)
	if err != nil {
		return handleErr(cfg, err)
	}

	return link
}

func OrgTemporalUILink(cfg *internal.Config, orgID string) string {
	link, err := url.JoinPath(cfg.TemporalUIURL,
		"namespaces",
		"orgs",
		"workflows",
		"event-loop-"+orgID)
	if err != nil {
		return handleErr(cfg, err)
	}

	return link
}

func OrgAPILink(cfg *internal.Config, orgID string) string {
	link, err := url.JoinPath(cfg.PublicAPIURL,
		"v1",
		"orgs", orgID)
	if err != nil {
		return handleErr(cfg, err)
	}

	return link
}

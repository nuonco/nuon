package links

import (
	"net/url"

	"github.com/powertoolsdev/mono/services/ctl-api/internal"
)

func InstallLinks(cfg *internal.Config, id string) map[string]any {
	return map[string]any{
		"ui":         InstallUILink(cfg, id),
		"api":        InstallAPILink(cfg, id),
		"event_loop": InstallEventLoopLink(cfg, id),
	}
}

func InstallLinksNested(cfg *internal.Config, installID string) map[string]any {
	return map[string]any{
		"ui":         InstallUILink(cfg, installID),
		"api":        InstallAPILink(cfg, installID),
		"event_loop": InstallEventLoopLink(cfg, installID),
	}
}

func InstallUILink(cfg *internal.Config, installID string) string {
	link, err := url.JoinPath(cfg.AppURL, "installs", installID)
	if err != nil {
		return handleErr(cfg, err)
	}

	return link
}

func InstallEventLoopLink(cfg *internal.Config, installID string) string {
	return eventLoopLink(cfg, "installs", installID)
}

func InstallAPILink(cfg *internal.Config, installID string) string {
	link, err := url.JoinPath(cfg.PublicAPIURL,
		"v1",
		"installs", installID)
	if err != nil {
		return handleErr(cfg, err)
	}

	return link
}

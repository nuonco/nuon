package links

import (
	"context"

	"github.com/powertoolsdev/mono/pkg/generics"
)

func InstallComponentLinks(ctx context.Context, installID string, compID string) map[string]any {
	links := map[string]any{
		"ui":  buildUILink(ctx, "v1", "installs", installID, "components", compID),
		"api": buildAPILink(ctx, "v1", "installs", installID, "components", compID),
	}
	if isEmployeeFromContext(ctx) {
		links = generics.MergeMap(links, InstallComponentEmployeeLinks(ctx, installID, compID))
	}

	return links
}

func InstallComponentEmployeeLinks(ctx context.Context, installID, componentID string) map[string]any {
	return map[string]any{
		"event_loop_ui": eventLoopLink(ctx, "installs", installID),
		"signal_ui":     eventLoopLink(ctx, "installs", installID),
		"admin_restart": buildAdminAPILink(ctx, "v1", "installs", installID, "components", componentID, "admin-restart"),
	}
}

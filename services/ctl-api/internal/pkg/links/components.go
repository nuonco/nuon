package links

import (
	"context"

	"github.com/powertoolsdev/mono/pkg/generics"
)

func ComponentLinks(ctx context.Context, componentID string) map[string]any {
	links := map[string]any{
		"ui":  buildUILink(ctx, "v1", "components", componentID),
		"api": buildAPILink(ctx, "v1", "components", componentID),
	}
	if isEmployeeFromContext(ctx) {
		links = generics.MergeMap(links, AppEmployeeLinks(ctx, componentID))
	}

	return links
}

func ComponentEmployeeLinks(ctx context.Context, componentID string) map[string]any {
	return map[string]any{
		"event_loop_ui": eventLoopLink(ctx, "components", componentID),
		"admin_restart": buildAdminAPILink(ctx, "v1", "components", componentID, "admin-restart"),
	}
}

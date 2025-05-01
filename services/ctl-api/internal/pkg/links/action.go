package links

import (
	"context"

	"github.com/powertoolsdev/mono/pkg/generics"
)

func ActionLinks(ctx context.Context, actionID string) map[string]any {
	links := map[string]any{
		"ui":  buildUILink(ctx, "v1", "actions", actionID),
		"api": buildAPILink(ctx, "v1", "actions", actionID),
	}
	if isEmployeeFromContext(ctx) {
		links = generics.MergeMap(links, AppEmployeeLinks(ctx, actionID))
	}

	return links
}

func ActionEmployeeLinks(ctx context.Context, actionID string) map[string]any {
	return map[string]any{
		"event_loop_ui": eventLoopLink(ctx, "actions", actionID),
		"admin_restart": buildAdminAPILink(ctx, "v1", "actions", actionID, "admin-restart"),
	}
}

package links

import (
	"context"

	"github.com/powertoolsdev/mono/pkg/generics"
)

func AppLinks(ctx context.Context, appID string) map[string]any {
	links := map[string]any{
		"ui":  buildUILink(ctx, orgIDFromContext(ctx), "apps", appID),
		"api": buildAPILink(ctx, "v1", "apps", appID),
	}
	if isEmployeeFromContext(ctx) {
		links = generics.MergeMap(links, AppEmployeeLinks(ctx, appID))
	}

	return links
}

func AppEmployeeLinks(ctx context.Context, appID string) map[string]any {
	return map[string]any{
		"event_loop_ui":     eventLoopLink(ctx, "apps", appID),
		"admin_restart":     buildAdminAPILink(ctx, "v1", "apps", appID, "admin-restart"),
		"admin_reprovision": buildAdminAPILink(ctx, "v1", "apps", appID, "admin-reprovision"),
	}
}

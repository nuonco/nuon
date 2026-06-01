package links

import (
	"context"

	"github.com/nuonco/nuon/pkg/generics"
)

func OrgLinks(ctx context.Context, orgID string) map[string]any {
	links := map[string]any{
		"ui":  buildUILink(ctx, orgIDFromContext(ctx)),
		"api": buildAPILink(ctx, "v1", "orgs", "current"),
	}
	if isEmployeeFromContext(ctx) {
		links = generics.MergeMap(links, AppEmployeeLinks(ctx, orgID))
	}

	return links
}

func OrgEmployeeLinks(ctx context.Context, orgID string) map[string]any {
	return map[string]any{
		"queue_ui":      queueLink(ctx, "orgs", orgID),
		"admin_restart": buildAdminAPILink(ctx, "v1", "orgs", orgID, "admin-restart"),
	}
}

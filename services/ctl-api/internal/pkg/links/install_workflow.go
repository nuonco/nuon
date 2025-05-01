package links

import (
	"context"

	"github.com/powertoolsdev/mono/pkg/generics"
)

func InstallWorkflowLinks(ctx context.Context, installDeployID string) map[string]any {
	links := map[string]any{
		"ui":  buildUILink(ctx, "v1", "installDeploys", installDeployID),
		"api": buildAPILink(ctx, "v1", "installDeploys", installDeployID),
	}
	if isEmployeeFromContext(ctx) {
		links = generics.MergeMap(links, AppEmployeeLinks(ctx, installDeployID))
	}

	return links
}

func InstallWorkflowEmployeeLinks(ctx context.Context, installDeployID string) map[string]any {
	return map[string]any{
		"event_loop_ui": eventLoopLink(ctx, "installDeploys", installDeployID),
		"admin_restart": buildAdminAPILink(ctx, "v1", "installDeploys", installDeployID, "admin-restart"),
	}
}

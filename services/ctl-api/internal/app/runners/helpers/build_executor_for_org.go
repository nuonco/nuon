package helpers

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type BuildExecutor = app.RunnerJobExecutor

const (
	BuildExecutorOrgRunner    = app.RunnerJobExecutorOrgRunner
	BuildExecutorControlPlane = app.RunnerJobExecutorControlPlane
)

func (h *Helpers) BuildExecutorForOrg(ctx context.Context, org *app.Org, typ app.RunnerJobType) (BuildExecutor, string, error) {
	if org == nil || org.ID == "" {
		return app.RunnerJobExecutorUnknown, "", fmt.Errorf("org is required")
	}

	resolvedOrg := org
	if resolvedOrg.Features[string(app.OrgFeatureControlPlaneBuilds)] {
		if typ == app.RunnerJobTypeDockerBuild {
			return app.RunnerJobExecutorUnknown, "", fmt.Errorf("docker_build components are not supported by control-plane builds; replace this component with a container_image component")
		}
		return app.RunnerJobExecutorControlPlane, "", nil
	}

	if (resolvedOrg.Features == nil || len(resolvedOrg.RunnerGroup.Runners) == 0) && h.db != nil {
		var fullOrg app.Org
		if res := h.db.WithContext(ctx).
			Preload("RunnerGroup").
			Preload("RunnerGroup.Runners").
			Where(app.Org{ID: org.ID}).
			First(&fullOrg); res.Error != nil {
			return app.RunnerJobExecutorUnknown, "", fmt.Errorf("unable to get org: %w", res.Error)
		}
		resolvedOrg = &fullOrg
	}

	if resolvedOrg.Features[string(app.OrgFeatureControlPlaneBuilds)] {
		if typ == app.RunnerJobTypeDockerBuild {
			return app.RunnerJobExecutorUnknown, "", fmt.Errorf("docker_build components are not supported by control-plane builds; replace this component with a container_image component")
		}
		return app.RunnerJobExecutorControlPlane, "", nil
	}

	if len(resolvedOrg.RunnerGroup.Runners) == 0 {
		return app.RunnerJobExecutorUnknown, "", fmt.Errorf("no runners available in runner group for org %s", resolvedOrg.ID)
	}

	return app.RunnerJobExecutorOrgRunner, resolvedOrg.RunnerGroup.Runners[0].ID, nil
}

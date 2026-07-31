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

	if typ == app.RunnerJobTypeDockerBuild {
		return app.RunnerJobExecutorUnknown, "", fmt.Errorf("docker_build components are not supported by control-plane builds; replace this component with a container_image component")
	}

	return app.RunnerJobExecutorControlPlane, "", nil
}

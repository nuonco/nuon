package plan

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (p *Planner) createDockerBuildPlan(_ workflow.Context, _ *app.ComponentBuild) (*plantypes.DockerBuildPlan, error) {
	// docker_build runs kaniko in-process inside the build runner pod,
	// which mutates the runner container's rootfs. Provider vendoring
	// needs git on PATH, so the two cannot share a pod.
	return nil, errors.Errorf(
		"docker_build components are not supported. " +
			"Kaniko corrupts the build runner's rootfs (removes /usr/bin/git etc.), " +
			"which breaks terraform provider vendoring. " +
			"Replace this docker_build component with a container_image component.",
	)
}

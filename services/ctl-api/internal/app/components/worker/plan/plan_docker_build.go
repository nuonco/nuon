package plan

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	componentsactivities "github.com/nuonco/nuon/services/ctl-api/internal/app/components/worker/activities"
)

func (p *Planner) createDockerBuildPlan(ctx workflow.Context, bld *app.ComponentBuild) (*plantypes.DockerBuildPlan, error) {
	// docker_build runs kaniko in-process inside the build runner pod,
	// which mutates the runner container's rootfs to perform the user's
	// docker build. After kaniko runs, /usr/bin/git (and other runner
	// tooling) is gone. The terraform-provider-mirror feature relies on
	// `terraform get` finding git on PATH at build time, so the two
	// features are mutually incompatible inside the same pod. Refuse to
	// plan a docker build when the org has the mirror flag enabled.
	mirrorEnabled, err := componentsactivities.AwaitOrgHasFeature(ctx, componentsactivities.OrgHasFeatureRequest{
		OrgID:   bld.OrgID,
		Feature: string(app.OrgFeatureTerraformProviderMirror),
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to check terraform provider mirror feature flag")
	}
	if mirrorEnabled {
		return nil, errors.Errorf(
			"docker_build components are not supported when the %q feature is enabled. "+
				"Kaniko corrupts the build runner's rootfs (removes /usr/bin/git etc.), "+
				"which breaks subsequent terraform provider vendoring. "+
				"Replace this docker_build component with a container_image component, "+
				"or disable the feature on the org.",
			app.OrgFeatureTerraformProviderMirror,
		)
	}

	return &plantypes.DockerBuildPlan{
		BuildArgs:  map[string]*string{},
		Target:     bld.ComponentConfigConnection.DockerBuildComponentConfig.Target,
		Dockerfile: bld.ComponentConfigConnection.DockerBuildComponentConfig.Dockerfile,
	}, nil
}

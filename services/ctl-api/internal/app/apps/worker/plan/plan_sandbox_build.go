package plan

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/buildsource"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
)

func (p *Planner) createSandboxBuildPlan(ctx workflow.Context, req *CreateSandboxBuildPlanRequest) (*plantypes.BuildPlan, error) {
	build, err := activities.AwaitGetAppSandboxBuildByIDByBuildID(ctx, req.AppSandboxBuildID)
	if err != nil {
		return nil, fmt.Errorf("unable to get sandbox build: %w", err)
	}

	source, err := activities.AwaitResolveSandboxBuildSource(ctx, &activities.ResolveSandboxBuildSourceInput{
		AppConfigID: build.AppConfigID,
		BuildID:     build.ID,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to get git source for sandbox build: %w", err)
	}
	gitSource, identity := source.GitSource, source.SourceIdentity

	dstCfg, err := sharedactivities.AwaitGetSandboxBuildOCIRegistry(ctx, sharedactivities.GetSandboxBuildOCIRegistryRequest{
		AppID: build.AppID,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to get OCI registry for sandbox build: %w", err)
	}

	tfPlan := buildsource.SandboxInputs(&build.AppSandboxConfig)
	tfPlan.Labels = map[string]string{
		"app_id":               build.AppID,
		"app_sandbox_build_id": build.ID,
	}

	plan := &plantypes.BuildPlan{
		Src:                gitSource,
		SourceIdentity:     identity,
		Dst:                dstCfg,
		DstTag:             build.ID,
		TerraformBuildPlan: tfPlan,
	}

	isSandboxOrg, err := activities.AwaitIsOrgSandboxMode(ctx, activities.IsOrgSandboxModeRequest{
		OrgID: build.OrgID,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to check org sandbox mode: %w", err)
	}
	if isSandboxOrg {
		plan.SandboxMode = &plantypes.SandboxMode{
			Enabled: true,
		}
	}

	return plan, nil
}

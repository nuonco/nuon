package plan

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/activities"
)

func (p *Planner) createSandboxBuildPlan(ctx workflow.Context, req *CreateSandboxBuildPlanRequest) (*plantypes.BuildPlan, error) {
	build, err := activities.AwaitGetAppSandboxBuildByIDByBuildID(ctx, req.AppSandboxBuildID)
	if err != nil {
		return nil, fmt.Errorf("unable to get sandbox build: %w", err)
	}

	gitSource, err := activities.AwaitGetSandboxBuildGitSource(ctx, activities.GetSandboxBuildGitSourceRequest{
		SandboxConfigID: build.AppSandboxConfigID,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to get git source for sandbox build: %w", err)
	}

	dstCfg, err := activities.AwaitGetSandboxBuildOCIRegistry(ctx, activities.GetSandboxBuildOCIRegistryRequest{
		AppID: build.AppID,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to get OCI registry for sandbox build: %w", err)
	}

	tfPlan := &plantypes.TerraformBuildPlan{
		Labels: map[string]string{
			"app_id":               build.AppID,
			"app_sandbox_build_id": build.ID,
		},
	}
	mirrorEnabled, err := activities.AwaitOrgHasFeature(ctx, activities.OrgHasFeatureRequest{
		OrgID:   build.OrgID,
		Feature: string(app.OrgFeatureTerraformProviderMirror),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to check terraform provider mirror feature flag: %w", err)
	}
	if mirrorEnabled {
		tfPlan.VendorProviders = true
		tfPlan.TerraformVersion = build.AppSandboxConfig.TerraformVersion
	}

	plan := &plantypes.BuildPlan{
		Src:                gitSource,
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

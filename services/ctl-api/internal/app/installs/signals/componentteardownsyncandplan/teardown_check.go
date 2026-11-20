package componentteardownsyncandplan

import (
	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// teardownPlanRetain returns the retained addresses a terraform or opentofu
// teardown plan must be checked against. The check is the backstop for runners
// that predate teardown.retain and ignore it, which would otherwise destroy
// everything.
func teardownPlanRetain(deployPlan *plantypes.DeployPlan, compType app.ComponentType, sandboxMode bool) []string {
	if sandboxMode {
		return nil
	}
	if deployPlan == nil || deployPlan.Teardown == nil {
		return nil
	}
	switch compType {
	case app.ComponentTypeTerraformModule, app.ComponentTypeOpentofuModule:
		return deployPlan.Teardown.Retain
	default:
		return nil
	}
}

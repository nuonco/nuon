package componentteardownsyncandplan

import (
	"encoding/json"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/pkg/errors"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/terraform/teardown"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// checkTeardownPlan refuses a terraform teardown plan that deletes a retained
// resource. This is the backstop for runners that predate teardown.retain and
// ignore it, which would otherwise destroy everything.
func checkTeardownPlan(deployPlan *plantypes.DeployPlan, compType app.ComponentType, sandboxMode bool, contentsDisplay string) error {
	if sandboxMode || compType != app.ComponentTypeTerraformModule {
		return nil
	}
	if deployPlan == nil || deployPlan.Teardown == nil || len(deployPlan.Teardown.Retain) == 0 {
		return nil
	}

	var plan tfjson.Plan
	if err := json.Unmarshal([]byte(contentsDisplay), &plan); err != nil {
		return errors.Wrap(err, "unable to parse teardown plan to verify retained resources")
	}

	if err := teardown.CheckPlan(&plan, deployPlan.Teardown.Retain); err != nil {
		return errors.Wrap(err, "teardown plan rejected (if this runner predates teardown.retain support, update it and retry)")
	}

	return nil
}

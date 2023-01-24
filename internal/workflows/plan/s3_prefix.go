package plan

import (
	"fmt"

	planv1 "github.com/powertoolsdev/protos/workflows/generated/types/executors/v1/plan/v1"
)

// getS3Prefix returns the prefix to be used for the plan and it's encompassed files
func getS3Prefix(req *planv1.CreatePlanRequest) string {
	return fmt.Sprintf("org=%s/app=%s/component=%s/deployment=%s",
		req.OrgId,
		req.AppId,
		req.Component.Name,
		req.DeploymentId)
}

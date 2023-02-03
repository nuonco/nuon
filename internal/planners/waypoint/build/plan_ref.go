package build

import (
	"path/filepath"

	planv1 "github.com/powertoolsdev/protos/workflows/generated/types/executors/v1/plan/v1"
)

const (
	planKey string = "build_plan.json"
)

func (p *planner) GetPlanRef() *planv1.PlanRef {
	return &planv1.PlanRef{
		Bucket:              p.OrgMetadata.Buckets.DeploymentsBucket,
		BucketKey:           filepath.Join(p.getPrefix(), planKey),
		BucketAssumeRoleArn: p.OrgMetadata.IamRoleArns.DeploymentsRoleArn,
		Type:                planv1.PlanType_PLAN_TYPE_WAYPOINT_BUILD,
	}
}

package build

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/go-generics"
	componentv1 "github.com/powertoolsdev/protos/components/generated/types/component/v1"
	planv1 "github.com/powertoolsdev/protos/workflows/generated/types/executors/v1/plan/v1"
	"github.com/stretchr/testify/assert"
)

func Test_planner_GetPlanRef(t *testing.T) {
	meta := generics.GetFakeObj[*planv1.Metadata]()
	orgMeta := generics.GetFakeObj[*planv1.OrgMetadata]()
	component := generics.GetFakeObj[*componentv1.Component]()

	pln, err := New(validator.New(), WithComponent(component), WithOrgMetadata(orgMeta), WithMetadata(meta))
	assert.NoError(t, err)

	planRef := pln.GetPlanRef()
	assert.Equal(t, orgMeta.Buckets.DeploymentsBucket, planRef.Bucket)
	assert.Equal(t, orgMeta.IamRoleArns.DeploymentsRoleArn, planRef.BucketAssumeRoleArn)
	assert.Equal(t, planv1.PlanType_PLAN_TYPE_WAYPOINT_BUILD, planRef.Type)
}

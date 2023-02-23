package build

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/go-generics"
	componentv1 "github.com/powertoolsdev/protos/components/generated/types/component/v1"
	planv1 "github.com/powertoolsdev/protos/workflows/generated/types/executors/v1/plan/v1"
	"github.com/powertoolsdev/workers-executors/internal/planners/waypoint"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func Test_planner_getBasePlan(t *testing.T) {
	v := validator.New()
	metadata := generics.GetFakeObj[*planv1.Metadata]()
	orgMetadata := generics.GetFakeObj[*planv1.OrgMetadata]()
	component := generics.GetFakeObj[*componentv1.Component]()

	planner, err := New(v,
		waypoint.WithComponent(component),
		waypoint.WithOrgMetadata(orgMetadata),
		waypoint.WithMetadata(metadata),
	)
	assert.NoError(t, err)

	p := planner.getBasePlan()

	// assert pass through configs
	assert.True(t, proto.Equal(metadata, p.Metadata))
	assert.True(t, proto.Equal(orgMetadata.WaypointServer, p.WaypointServer))

	// assert ecr
	assert.Equal(t, metadata.DeploymentShortId, p.EcrRepositoryRef.Tag)

	// assert waypoint
	assert.Equal(t, p.Metadata.AppShortId, p.WaypointRef.Project)
	assert.Equal(t, p.Metadata.AppShortId, p.WaypointRef.Workspace)
	assert.Equal(t, component.Id, p.WaypointRef.App)

	assert.Contains(t, p.WaypointRef.SingletonId, p.Component.Id)
	assert.Contains(t, p.WaypointRef.SingletonId, metadata.DeploymentShortId)
	assert.Equal(t, p.WaypointRef.Labels, waypoint.DefaultLabels(metadata, component.Name, "build"))
	assert.Equal(t, p.Metadata.OrgShortId, p.WaypointRef.RunnerId)
	assert.Equal(t, p.Metadata.OrgShortId, p.WaypointRef.OnDemandRunnerConfig)
	assert.Equal(t, defaultBuildTimeoutSeconds, p.WaypointRef.JobTimeoutSeconds)
	assert.Equal(t, planv1.WaypointJobType_WAYPOINT_JOB_TYPE_BUILD, p.WaypointRef.JobType)

	// assert outputs
	assert.Equal(t, orgMetadata.Buckets.DeploymentsBucket, p.Outputs.Bucket)
	assert.Equal(t, planner.Prefix(), p.Outputs.BucketPrefix)
	assert.Equal(t, orgMetadata.IamRoleArns.DeploymentsRoleArn, p.Outputs.BucketAssumeRoleArn)
}

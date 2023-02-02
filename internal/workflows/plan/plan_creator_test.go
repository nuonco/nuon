package plan

import (
	"context"
	"fmt"
	"testing"

	waypointv1 "github.com/hashicorp/waypoint/pkg/server/gen"
	"github.com/powertoolsdev/go-generics"
	componentv1 "github.com/powertoolsdev/protos/components/generated/types/component/v1"
	planv1 "github.com/powertoolsdev/protos/workflows/generated/types/executors/v1/plan/v1"
	planactivitiesv1 "github.com/powertoolsdev/protos/workflows/generated/types/executors/v1/plan/v1/activities/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockS3BlobUploader struct {
	mock.Mock
}

func (m *mockS3BlobUploader) UploadBlob(ctx context.Context, byts []byte, fileName string) error {
	args := m.Called(ctx, byts, fileName)
	return args.Error(0)
}

var _ s3BlobUploader = (*mockS3BlobUploader)(nil)

func Test_planCreatorImpl_uploadPlan(t *testing.T) {
	errUploadPlan := fmt.Errorf("error uploading plan")
	req := generics.GetFakeObj[*planactivitiesv1.CreatePlanRequest]()
	buildPlan := generics.GetFakeObj[*planv1.WaypointPlan]()

	tests := map[string]struct {
		clientFn    func(*testing.T) s3BlobUploader
		assertFn    func(*testing.T, s3BlobUploader, *planv1.PlanRef)
		errExpected error
	}{
		"happy path - correct upload": {
			clientFn: func(t *testing.T) s3BlobUploader {
				client := &mockS3BlobUploader{}
				client.On("UploadBlob", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				return client
			},
			assertFn: func(t *testing.T, client s3BlobUploader, planRef *planv1.PlanRef) {
				obj := client.(*mockS3BlobUploader)
				obj.AssertNumberOfCalls(t, "UploadBlob", 1)

				args := obj.Calls[0].Arguments
				assert.Equal(t, planFilename, args[2].(string))

				//expectedByts, err := proto.Marshal(buildPlan)
				//assert.NoError(t, err)
				//assert.Equal(t, expectedByts, args[1].([]byte))
			},
		},
		"happy path - correct plan ref response": {
			clientFn: func(t *testing.T) s3BlobUploader {
				client := &mockS3BlobUploader{}
				client.On("UploadBlob", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				return client
			},
			assertFn: func(t *testing.T, client s3BlobUploader, planRef *planv1.PlanRef) {
				obj := client.(*mockS3BlobUploader)
				obj.AssertNumberOfCalls(t, "UploadBlob", 1)

				//assert.Equal(t, req.Config.DeploymentsBucket, planRef.Bucket)
				//assert.Equal(t, filepath.Join(req.DeploymentsBucketPrefix, planFilename), planRef.BucketKey)
				//assert.Equal(t, req.DeploymentsBucketAssumeRoleARN, planRef.BucketAssumeRoleArn)
			},
		},
		"error uploading file": {
			clientFn: func(t *testing.T) s3BlobUploader {
				client := &mockS3BlobUploader{}
				client.On("UploadBlob", mock.Anything, mock.Anything, mock.Anything).Return(errUploadPlan).Once()
				return client
			},
			errExpected: errUploadPlan,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			pc := planCreatorImpl{}
			client := test.clientFn(t)

			planRef, err := pc.uploadPlan(context.Background(), client, req, buildPlan)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
			test.assertFn(t, client, planRef)
		})
	}
}

func Test_planCreatorImpl_getConfigBuilder(t *testing.T) {
	// TODO(jm): this is just a place holder until we build out actual mapping between builders
	comp := generics.GetFakeObj[*componentv1.Component]()

	planCreator := &planCreatorImpl{}
	builder, err := planCreator.getConfigBuilder(comp)
	assert.NoError(t, err)
	assert.NotNil(t, builder)
}

//nolint:all
type mockBuilder struct {
	mock.Mock
}

//nolint:all
func (m *mockBuilder) WithComponent(comp *componentv1.Component) {
	m.Called(comp)
}

//nolint:all
func (m *mockBuilder) WithMetadata(meta *planv1.Metadata) {
	m.Called(meta)
}

//nolint:all
func (m *mockBuilder) WithECRRef(ecr *planv1.ECRRepositoryRef) {
	m.Called(ecr)
}

//nolint:all
func (m *mockBuilder) Render() ([]byte, waypointv1.Hcl_Format, error) {
	args := m.Called()
	if args.Error(2) != nil {
		return nil, waypointv1.Hcl_HCL, args.Error(2)
	}

	return args.Get(0).([]byte), args.Get(1).(waypointv1.Hcl_Format), nil
}

var _ s3BlobUploader = (*mockS3BlobUploader)(nil)

//nolint:all
func newDefaultMockBuilder() *mockBuilder {
	obj := &mockBuilder{}
	obj.On("Render").Return([]byte("waypoint-hcl"), waypointv1.Hcl_HCL, nil)
	obj.On("WithComponent", mock.Anything).Return()
	obj.On("WithMetadata", mock.Anything).Return()
	obj.On("WithECRRef", mock.Anything).Return()
	return obj
}

func Test_planCreatorImpl_createPlan(t *testing.T) {
	//req := generics.GetFakeObj[*planactivitiesv1.CreatePlanRequest]()
	//longIDs := []string{uuid.NewString(), uuid.NewString(), uuid.NewString()}
	//shortIDs, err := shortid.ParseStrings(longIDs...)
	//assert.NoError(t, err)

	//req.OrgID = shortIDs[0]
	//req.AppID = shortIDs[1]
	//req.DeploymentID = shortIDs[2]
	//req.Config.WaypointTokenSecretTemplate = "token-%s"
	//assert.NoError(t, req.validate())

	//errCreatePlan := fmt.Errorf("err creating plan")
	//tests := map[string]struct {
	//builderFn   func() *mockBuilder
	//assertFn    func(*testing.T, *planv1.WaypointPlan)
	//errExpected error
	//}{
	//"happy path - metadata": {
	//builderFn: newDefaultMockBuilder,
	//assertFn: func(t *testing.T, plan *planv1.WaypointPlan) {
	//meta := plan.Metadata
	//assert.NotNil(t, meta)
	//assert.Equal(t, longIDs[0], meta.OrgId)
	//assert.Equal(t, shortIDs[0], meta.OrgShortId)

	//assert.Equal(t, longIDs[1], meta.AppId)
	//assert.Equal(t, shortIDs[1], meta.AppShortId)

	//assert.Equal(t, longIDs[2], meta.DeploymentId)
	//assert.Equal(t, shortIDs[2], meta.DeploymentShortId)
	//},
	//},
	//"happy path - waypoint server": {
	//builderFn: newDefaultMockBuilder,
	//assertFn: func(t *testing.T, plan *planv1.WaypointPlan) {
	//wpPlan := plan.WaypointServer
	//cfg := req.Config

	//expectedAddr := client.DefaultOrgServerAddress(cfg.WaypointServerRootDomain, req.OrgID)
	//assert.Equal(t, expectedAddr, wpPlan.Address)

	//expectedTokenSecretName := fmt.Sprintf(cfg.WaypointTokenSecretTemplate, req.OrgID)
	//assert.Equal(t, expectedTokenSecretName, wpPlan.TokenSecretName)
	//assert.Equal(t, cfg.WaypointTokenSecretNamespace, wpPlan.TokenSecretNamespace)
	//},
	//},
	//"happy path - ecr repository": {
	//builderFn: newDefaultMockBuilder,
	//assertFn: func(t *testing.T, plan *planv1.WaypointPlan) {
	//ecrPlan := plan.EcrRepositoryRef
	//cfg := req.Config

	//expectedRepoName := fmt.Sprintf("%s/%s", req.OrgID, req.AppID)
	//expectedRepoURI := fmt.Sprintf("%s.dkr.ecr.%s.amazonaws.com/%s", cfg.OrgsECRRegistryID,
	//cfg.OrgsECRRegion, expectedRepoName)
	//expectedRepoArn := fmt.Sprintf("%s/%s", cfg.OrgsECRRegistryARN, expectedRepoName)

	//assert.Equal(t, cfg.OrgsECRRegistryID, ecrPlan.RegistryId)
	//assert.Equal(t, expectedRepoName, ecrPlan.RepositoryName)
	//assert.Equal(t, expectedRepoArn, ecrPlan.RepositoryArn)
	//assert.Equal(t, expectedRepoURI, ecrPlan.RepositoryUri)
	//assert.Equal(t, req.DeploymentID, ecrPlan.Tag)
	//assert.Equal(t, cfg.OrgsECRRegion, ecrPlan.Region)
	//},
	//},
	//"happy path - waypoint ref": {
	//builderFn: newDefaultMockBuilder,
	//assertFn: func(t *testing.T, plan *planv1.WaypointPlan) {
	//wpPlan := plan.WaypointRef

	//assert.Equal(t, req.OrgID, wpPlan.Project)
	//assert.Equal(t, req.AppID, wpPlan.Workspace)
	//assert.Equal(t, req.Component.Name, wpPlan.App)
	//assert.Contains(t, wpPlan.SingletonId, req.Component.Name)
	//assert.Contains(t, wpPlan.SingletonId, req.DeploymentID)
	//assert.Equal(t, req.OrgID, wpPlan.RunnerId)
	//assert.Equal(t, req.OrgID, wpPlan.OnDemandRunnerConfig)
	//assert.Equal(t, defaultJobTimeoutSeconds, wpPlan.JobTimeoutSeconds)

	//assert.Equal(t, req.OrgID, wpPlan.Labels["org-id"])
	//assert.Equal(t, req.DeploymentID, wpPlan.Labels["deployment-id"])
	//assert.Equal(t, req.AppID, wpPlan.Labels["app-id"])
	//assert.Equal(t, req.Component.Name, wpPlan.Labels["component-name"])

	//assert.Equal(t, "waypoint-hcl", wpPlan.HclConfig)
	//assert.Equal(t, waypointv1.Hcl_HCL.String(), wpPlan.HclConfigFormat)
	//},
	//},
	//"happy path - outputs": {
	//builderFn: newDefaultMockBuilder,
	//assertFn: func(t *testing.T, plan *planv1.WaypointPlan) {
	//oPlan := plan.Outputs
	//cfg := req.Config

	//assert.Equal(t, cfg.DeploymentsBucket, oPlan.Bucket)
	//assert.Equal(t, req.DeploymentsBucketPrefix, oPlan.BucketPrefix)
	//assert.Equal(t, req.DeploymentsBucketAssumeRoleARN, oPlan.BucketAssumeRoleArn)

	//assert.Contains(t, oPlan.LogsKey, req.DeploymentsBucketPrefix)
	//assert.Contains(t, oPlan.LogsKey, "logs.txt")

	//assert.Contains(t, oPlan.EventsKey, req.DeploymentsBucketPrefix)
	//assert.Contains(t, oPlan.EventsKey, "events.json")

	//assert.Contains(t, oPlan.ArtifactKey, req.DeploymentsBucketPrefix)
	//assert.Contains(t, oPlan.ArtifactKey, "artifacts.json")
	//},
	//},
	//"happy path - component": {
	//builderFn: newDefaultMockBuilder,
	//assertFn: func(t *testing.T, plan *planv1.WaypointPlan) {
	//assert.True(t, proto.Equal(req.Component, plan.Component))
	//},
	//},
	//"error - builder": {
	//builderFn: func() *mockBuilder {
	//obj := &mockBuilder{}
	//obj.On("Render").Return(nil, waypointv1.Hcl_HCL, errCreatePlan)
	//obj.On("WithComponent", mock.Anything).Return()
	//obj.On("WithMetadata", mock.Anything).Return()
	//obj.On("WithECRRef", mock.Anything).Return()
	//return obj
	//},
	//errExpected: errCreatePlan,
	//},
	//}

	//for name, test := range tests {
	//t.Run(name, func(t *testing.T) {
	//pc := planCreatorImpl{}
	//builder := test.builderFn()

	//buildPlan, err := pc.createPlan(req, builder)
	//if test.errExpected != nil {
	//assert.ErrorContains(t, err, test.errExpected.Error())
	//return
	//}
	//assert.NoError(t, err)
	//test.assertFn(t, buildPlan)
	//})
	//}
}

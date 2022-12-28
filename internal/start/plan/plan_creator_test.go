package plan

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/powertoolsdev/go-generics"
	componentv1 "github.com/powertoolsdev/protos/components/generated/types/component/v1"
	planv1 "github.com/powertoolsdev/protos/deployments/generated/types/plan/v1"
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
	req := generics.GetFakeObj[CreatePlanRequest]()
	buildPlan := generics.GetFakeObj[*planv1.BuildPlan]()

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

				expectedByts, err := json.Marshal(buildPlan)
				assert.NoError(t, err)
				assert.Equal(t, expectedByts, args[1].([]byte))
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

				assert.Equal(t, req.Config.DeploymentsBucket, planRef.Bucket)
				assert.Equal(t, filepath.Join(req.DeploymentsBucketPrefix, planFilename), planRef.BucketKey)
				assert.Equal(t, req.DeploymentsBucketAssumeRoleARN, planRef.BucketAssumeRoleArn)
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

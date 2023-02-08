package plan

import (
	"context"
	"fmt"
	"testing"

	"github.com/powertoolsdev/go-generics"
	planv1 "github.com/powertoolsdev/protos/workflows/generated/types/executors/v1/plan/v1"
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

	plan := generics.GetFakeObj[*planv1.WaypointPlan]()
	planRef := generics.GetFakeObj[*planv1.PlanRef]()

	tests := map[string]struct {
		clientFn    func(*testing.T) s3BlobUploader
		assertFn    func(*testing.T, s3BlobUploader)
		errExpected error
	}{
		"happy path - correct upload": {
			clientFn: func(t *testing.T) s3BlobUploader {
				client := &mockS3BlobUploader{}
				client.On("UploadBlob", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				return client
			},
			assertFn: func(t *testing.T, client s3BlobUploader) {
				obj := client.(*mockS3BlobUploader)
				obj.AssertNumberOfCalls(t, "UploadBlob", 1)

				args := obj.Calls[0].Arguments

				assert.Equal(t, planRef.BucketKey, args[2].(string))

				//expectedByts, err := proto.Marshal(plan)
				//assert.NoError(t, err)
				//assert.Equal(t, expectedByts, args[1].([]byte))
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

			err := pc.uploadPlan(context.Background(), client, planRef, plan)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
			test.assertFn(t, client)
		})
	}
}

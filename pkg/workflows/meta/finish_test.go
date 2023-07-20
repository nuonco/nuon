package meta

import (
	"context"
	"fmt"
	"testing"

	"github.com/powertoolsdev/mono/pkg/generics"
	orgsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/orgs/v1"
	sharedv1 "github.com/powertoolsdev/mono/pkg/types/workflows/shared/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/proto"
)

type mockFinisherUploadClient struct {
	mock.Mock
}

func (m *mockFinisherUploadClient) UploadBlob(ctx context.Context, byts []byte, fileName string) error {
	args := m.Called(ctx, byts, fileName)
	return args.Error(0)
}

var _ finisherUploadClient = (*mockFinisherUploadClient)(nil)

func Test_finisherImpl_writeRequestFile(t *testing.T) {
	errStartUpload := fmt.Errorf("error uploading start request")
	req := generics.GetFakeObj[*sharedv1.Response]()
	req.Response = &sharedv1.ResponseRef{
		Response: &sharedv1.ResponseRef_OrgSignup{
			OrgSignup: generics.GetFakeObj[*orgsv1.ProvisionResponse](),
		},
	}

	tests := map[string]struct {
		clientFn    func(*testing.T) finisherUploadClient
		assertFn    func(*testing.T, finisherUploadClient)
		errExpected error
	}{
		"happy path": {
			clientFn: func(t *testing.T) finisherUploadClient {
				client := &mockFinisherUploadClient{}
				client.On("UploadBlob", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				return client
			},
			assertFn: func(t *testing.T, client finisherUploadClient) {
				obj := client.(*mockFinisherUploadClient)
				obj.AssertNumberOfCalls(t, "UploadBlob", 1)

				args := obj.Calls[0].Arguments
				assert.Equal(t, finishRequestFilename, args[2].(string))

				expectedByts, err := proto.Marshal(req)
				assert.NoError(t, err)
				assert.Equal(t, expectedByts, args[1].([]byte))
			},
		},
		"can be unmarshaled": {
			clientFn: func(t *testing.T) finisherUploadClient {
				client := &mockFinisherUploadClient{}
				client.On("UploadBlob", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				return client
			},
			assertFn: func(t *testing.T, client finisherUploadClient) {
				obj := client.(*mockFinisherUploadClient)
				obj.AssertNumberOfCalls(t, "UploadBlob", 1)
				args := obj.Calls[0].Arguments

				byts := args[1].([]byte)
				responseObj := &sharedv1.Response{}
				err := proto.Unmarshal(byts, responseObj)
				assert.NoError(t, err)
			},
		},
		"error uploading file": {
			clientFn: func(t *testing.T) finisherUploadClient {
				client := &mockFinisherUploadClient{}
				client.On("UploadBlob", mock.Anything, mock.Anything, mock.Anything).Return(errStartUpload).Once()
				return client
			},
			errExpected: errStartUpload,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			s := finisherImpl{}
			client := test.clientFn(t)
			err := s.writeRequestFile(context.Background(), client, req)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.Nil(t, err)
			test.assertFn(t, client)
		})
	}
}

func Test_finisherImpl_getRequest(t *testing.T) {
	inputResp := generics.GetFakeObj[*sharedv1.FinishActivityRequest]()
	inputResp.ResponseRef = &sharedv1.ResponseRef{
		Response: &sharedv1.ResponseRef_OrgSignup{
			OrgSignup: generics.GetFakeObj[*orgsv1.ProvisionResponse](),
		},
	}

	finisher := &finisherImpl{}
	resp := finisher.getResponse(inputResp)

	assert.Equal(t, inputResp.Status, resp.Status)
	assert.True(t, proto.Equal(inputResp.ResponseRef, inputResp.ResponseRef))
}

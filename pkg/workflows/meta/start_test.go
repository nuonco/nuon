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

type mockStarterUploadClient struct {
	mock.Mock
}

func (m *mockStarterUploadClient) UploadBlob(ctx context.Context, byts []byte, fileName string) error {
	args := m.Called(ctx, byts, fileName)
	return args.Error(0)
}

var _ starterUploadClient = (*mockStarterUploadClient)(nil)

func Test_starterImpl_writeRequestFile(t *testing.T) {
	errStartUpload := fmt.Errorf("error uploading start request")
	req := generics.GetFakeObj[*sharedv1.Request]()
	req.Request = &sharedv1.RequestRef{
		Request: &sharedv1.RequestRef_OrgSignup{
			OrgSignup: generics.GetFakeObj[*orgsv1.ProvisionRequest](),
		},
	}

	tests := map[string]struct {
		clientFn    func(*testing.T) starterUploadClient
		assertFn    func(*testing.T, starterUploadClient)
		errExpected error
	}{
		"happy path": {
			clientFn: func(t *testing.T) starterUploadClient {
				client := &mockStarterUploadClient{}
				client.On("UploadBlob", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				return client
			},
			assertFn: func(t *testing.T, client starterUploadClient) {
				obj := client.(*mockStarterUploadClient)
				obj.AssertNumberOfCalls(t, "UploadBlob", 1)

				args := obj.Calls[0].Arguments
				assert.Equal(t, startRequestFilename, args[2].(string))

				expectedByts, err := proto.Marshal(req)
				assert.NoError(t, err)
				assert.Equal(t, expectedByts, args[1].([]byte))
			},
		},
		"can be unmarshaled": {
			clientFn: func(t *testing.T) starterUploadClient {
				client := &mockStarterUploadClient{}
				client.On("UploadBlob", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				return client
			},
			assertFn: func(t *testing.T, client starterUploadClient) {
				obj := client.(*mockStarterUploadClient)
				obj.AssertNumberOfCalls(t, "UploadBlob", 1)
				args := obj.Calls[0].Arguments

				byts := args[1].([]byte)
				requestObj := &sharedv1.Request{}
				err := proto.Unmarshal(byts, requestObj)
				assert.NoError(t, err)
			},
		},
		"error uploading file": {
			clientFn: func(t *testing.T) starterUploadClient {
				client := &mockStarterUploadClient{}
				client.On("UploadBlob", mock.Anything, mock.Anything, mock.Anything).Return(errStartUpload).Once()
				return client
			},
			errExpected: errStartUpload,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			s := starterImpl{}
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

func Test_starterImpl_getRequest(t *testing.T) {
	startReq := generics.GetFakeObj[*sharedv1.StartActivityRequest]()
	startReq.RequestRef = &sharedv1.RequestRef{
		Request: &sharedv1.RequestRef_OrgSignup{
			OrgSignup: generics.GetFakeObj[*orgsv1.ProvisionRequest](),
		},
	}

	starter := &starterImpl{}
	req := starter.getRequest(startReq)

	assert.Equal(t, startReq.WorkflowInfo.Id, req.WorkflowId)
	assert.True(t, proto.Equal(startReq.RequestRef, req.Request))
}

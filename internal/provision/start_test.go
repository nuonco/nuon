package provision

import (
	"context"
	"fmt"
	"testing"

	"github.com/powertoolsdev/go-generics"
	"github.com/powertoolsdev/go-sender"
	installsv1 "github.com/powertoolsdev/protos/workflows/generated/types/installs/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/exp/slices"
)

// mock sender
type mockSender struct {
	mock.Mock
}

func (m *mockSender) Send(ctx context.Context, msg string) error {
	args := m.Called(ctx, msg)
	return args.Error(0)
}

var _ sender.NotificationSender = (*mockSender)(nil)

func Test_sendStartNotification(t *testing.T) {
	tests := map[string]struct {
		fn          func(*testing.T, func(string) bool) sender.NotificationSender
		req         StartWorkflowRequest
		errExpected error
	}{
		"happy path": {
			req: generics.GetFakeObj[StartWorkflowRequest](),
			fn: func(t *testing.T, matcher func(string) bool) sender.NotificationSender {
				ms := &mockSender{}
				ms.On("Send", mock.Anything, mock.MatchedBy(matcher)).Return(nil).Once()

				return ms
			},
		},

		"error on send": {
			req:         generics.GetFakeObj[StartWorkflowRequest](),
			errExpected: fmt.Errorf("send error"),
			fn: func(t *testing.T, matcher func(string) bool) sender.NotificationSender {
				ms := &mockSender{}
				ms.On("Send", mock.Anything, mock.MatchedBy(matcher)).Return(fmt.Errorf("send error")).Once()

				return ms
			},
		},

		"error without sender": {
			req:         generics.GetFakeObj[StartWorkflowRequest](),
			errExpected: errNoValidSender,
			fn: func(t *testing.T, matcher func(string) bool) sender.NotificationSender {
				return nil
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			matcher := func(s string) bool {
				var accum []bool
				for _, v := range []string{test.req.AppID, test.req.InstallID, test.req.OrgID} {
					accum = append(accum, assert.Contains(t, s, v))
				}
				accum = append(accum, assert.Contains(t, s, "started provisioning sandbox"))
				return !slices.Contains(accum, false)
			}

			s := test.fn(t, matcher)
			n := &starterImpl{sender: s}

			err := n.sendStartNotification(context.Background(), test.req)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)

			if s, ok := s.(*mockSender); ok {
				s.AssertExpectations(t)
			}
		})
	}
}

// mock uploader
type tests3BlobUploader struct {
	mock.Mock
}

func (m *tests3BlobUploader) UploadBlob(ctx context.Context, byts []byte, outputName string) error {
	args := m.Called(ctx, byts, outputName)
	return args.Error(0)
}

func getStatusFileContents() StatusFileContents {
	return StatusFileContents{
		Status:       "Started",
		ErrorStep:    "",
		ErrorMessage: "",
	}
}

func Test_writeStatusFile(t *testing.T) {
	uploadErr := fmt.Errorf("unable to upload status file to s3: upload error")

	tests := map[string]struct {
		clientFn           func(*testing.T) s3BlobUploader
		assertFn           func(*testing.T, s3BlobUploader)
		errExpected        error
		statusFileContents func() StatusFileContents
	}{
		"happy path": {
			clientFn: func(t *testing.T) s3BlobUploader {
				client := &tests3BlobUploader{}
				client.On("UploadBlob", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				return client
			},
			assertFn: func(t *testing.T, client s3BlobUploader) {
				obj := client.(*tests3BlobUploader)
				obj.AssertNumberOfCalls(t, "UploadBlob", 1)
				assertFilename := obj.Calls[0].Arguments[2].(string)
				assert.Equal(t, assertFilename, "status.json")
			},
			statusFileContents: func() StatusFileContents {
				return getStatusFileContents()
			},
		},
		"update status file with error": {
			clientFn: func(t *testing.T) s3BlobUploader {
				client := &tests3BlobUploader{}
				client.On("UploadBlob", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				return client
			},
			assertFn: func(t *testing.T, client s3BlobUploader) {
				obj := client.(*tests3BlobUploader)
				obj.AssertNumberOfCalls(t, "UploadBlob", 1)
				assertFilename := obj.Calls[0].Arguments[2].(string)
				assert.Equal(t, assertFilename, "status.json")
			},
			statusFileContents: func() StatusFileContents {
				req := getStatusFileContents()
				req.Status = "Finished"
				req.ErrorStep = "provision_sandbox"
				req.ErrorMessage = "error provisioning sandbox"
				return req
			},
		},
		"error uploading file": {
			clientFn: func(t *testing.T) s3BlobUploader {
				client := &tests3BlobUploader{}
				client.On("UploadBlob", mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("upload error")).Once()
				return client
			},
			statusFileContents: func() StatusFileContents {
				return getStatusFileContents()
			},
			errExpected: uploadErr,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			n := starterImpl{}
			client := test.clientFn(t)
			err := n.writeStatusFile(context.Background(), client, test.statusFileContents())
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.Nil(t, err)
		})
	}
}

func Test_writeRequestFile(t *testing.T) {
	uploadErr := fmt.Errorf("unable to upload request file to s3: upload error")
	req := generics.GetFakeObj[StartWorkflowRequest]()

	tests := map[string]struct {
		clientFn    func(*testing.T) s3BlobUploader
		assertFn    func(*testing.T, s3BlobUploader)
		errExpected error
	}{
		"happy path": {
			clientFn: func(t *testing.T) s3BlobUploader {
				client := &tests3BlobUploader{}
				client.On("UploadBlob", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				return client
			},
			assertFn: func(t *testing.T, client s3BlobUploader) {
				obj := client.(*tests3BlobUploader)
				obj.AssertNumberOfCalls(t, "UploadBlob", 1)
				assertReq := obj.Calls[0].Arguments[1].(*installsv1.ProvisionRequest)
				assertFilename := obj.Calls[0].Arguments[2].(string)
				assert.Equal(t, assertReq, req)
				assert.Equal(t, assertFilename, "request.json")
			},
		},
		"error uploading file": {
			clientFn: func(t *testing.T) s3BlobUploader {
				client := &tests3BlobUploader{}
				client.On("UploadBlob", mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("upload error")).Once()
				return client
			},
			errExpected: uploadErr,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			n := starterImpl{}
			client := test.clientFn(t)
			err := n.writeRequestFile(context.Background(), client, req.ProvisionRequest)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.Nil(t, err)
		})
	}
}

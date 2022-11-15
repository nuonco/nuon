package build

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type tests3BlobUploader struct {
	mock.Mock
}

func (m *tests3BlobUploader) UploadBlob(ctx context.Context, byts []byte, outputName string) error {
	args := m.Called(ctx, byts, outputName)
	return args.Error(0)
}

func Test_uploadEventFile(t *testing.T) {
	uploadErr := fmt.Errorf("unable to upload events file to s3: upload error")

	tests := map[string]struct {
		osReadFileFn   func(string) ([]byte, error)
		osRemoveFileFn func(string) error
		clientFn       func(*testing.T) s3BlobUploader
		assertFn       func(*testing.T, s3BlobUploader)
		errExpected    error
	}{
		"happy path": {
			osReadFileFn: func(string) ([]byte, error) {
				return []byte(""), nil
			},
			clientFn: func(t *testing.T) s3BlobUploader {
				client := &tests3BlobUploader{}
				client.On("UploadBlob", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				return client
			},
			assertFn: func(t *testing.T, client s3BlobUploader) {
				obj := client.(*tests3BlobUploader)
				obj.AssertNumberOfCalls(t, "UploadBlob", 1)
				assertFilename := obj.Calls[0].Arguments[2].(string)
				assert.Equal(t, assertFilename, eventFilename)
			},
			osRemoveFileFn: func(string) error {
				return nil
			},
		},
		"error uploading file": {
			osReadFileFn: func(string) ([]byte, error) {
				return []byte(""), nil
			},
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
			n := waypointDeploymentJobPollerImpl{}
			osReadFile = test.osReadFileFn
			osRemoveFile = test.osRemoveFileFn
			client := test.clientFn(t)
			fileWriter := fileEventWriter{}
			err := n.uploadEventFile(context.Background(), client, &fileWriter)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.Nil(t, err)
		})
	}
}

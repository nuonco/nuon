package uploader

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strconv"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockUploader struct {
	mock.Mock
}

func (m *mockUploader) SetUploadPrefix(prefix string) {
	m.Called(prefix)
}

func (m *mockUploader) UploadFile(ctx context.Context, tmpDir, name, outputName string) error {
	args := m.Called(tmpDir, name, outputName)
	return args.Error(0)
}

func (m *mockUploader) UploadBlob(ctx context.Context, byts []byte, outputName string) error {
	args := m.Called(byts, outputName)
	return args.Error(0)
}

var _ Uploader = (*mockUploader)(nil)

func TestUploader_setPrefix(t *testing.T) {
	up := s3Uploader{}

	prefix := "/installID=abc"
	up.SetUploadPrefix(prefix)
	assert.NotNil(t, up.prefix)

	// assert that the output key is of format prefix/log_epoch
	expectedPrefix := prefix + "/runs/ts="
	assert.True(t, strings.HasPrefix(up.prefix, expectedPrefix))

	epochStr := strings.Split(up.prefix, expectedPrefix)[1]
	_, err := strconv.ParseInt(epochStr, 10, 64)
	assert.Nil(t, err)
}

type mockUploaderClient func(
	ctx context.Context,
	in *s3.PutObjectInput,
	fns ...func(*manager.Uploader),
) (*manager.UploadOutput, error)

func (m mockUploaderClient) Upload(
	ctx context.Context,
	in *s3.PutObjectInput,
	fns ...func(*manager.Uploader),
) (*manager.UploadOutput, error) {
	return m(ctx, in, fns...)
}

func Test_upload(t *testing.T) {
	clientErr := errors.New("some client error")
	testOutputKey := "output-key"
	testContent := []byte("test")
	bucketName := "bucketName"

	tests := map[string]struct {
		client      func(*testing.T) mockUploaderClient
		errExpected error
	}{
		"error returned": {
			errExpected: clientErr,
			client: func(t *testing.T) mockUploaderClient {
				return func(ctx context.Context, in *s3.PutObjectInput, fns ...func(*manager.Uploader)) (*manager.UploadOutput, error) {
					return nil, clientErr
				}
			},
		},
		"passes proper key for file": {
			client: func(t *testing.T) mockUploaderClient {
				return func(ctx context.Context, in *s3.PutObjectInput, fns ...func(*manager.Uploader)) (*manager.UploadOutput, error) {
					assert.Equal(t, testOutputKey, *in.Key)
					return nil, nil
				}
			},
		},
		"passes proper bucket for installations": {
			client: func(t *testing.T) mockUploaderClient {
				return func(ctx context.Context, in *s3.PutObjectInput, fns ...func(*manager.Uploader)) (*manager.UploadOutput, error) {
					assert.Equal(t, bucketName, *in.Bucket)
					return nil, nil
				}
			},
		},
		"passes a file handle": {
			client: func(t *testing.T) mockUploaderClient {
				return func(ctx context.Context, in *s3.PutObjectInput, fns ...func(*manager.Uploader)) (*manager.UploadOutput, error) {
					byts, err := io.ReadAll(in.Body)
					assert.Nil(t, err)
					assert.Equal(t, byts, testContent)
					return nil, nil
				}
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			client := test.client(t)

			up := s3Uploader{
				installBucket: bucketName,
			}
			f := bytes.NewBuffer(testContent)
			err := up.upload(context.Background(), client, f, testOutputKey)
			if test.errExpected != nil {
				assert.ErrorIs(t, err, test.errExpected)
			}
		})
	}
}

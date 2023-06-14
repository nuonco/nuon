package s3uploader

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/aws/credentials"
	"github.com/powertoolsdev/mono/pkg/generics"

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
				Bucket: bucketName,
			}
			f := bytes.NewBuffer(testContent)
			err := up.upload(context.Background(), client, f, testOutputKey)
			if test.errExpected != nil {
				assert.ErrorIs(t, err, test.errExpected)
			}
		})
	}
}

func TestNewS3Uploader(t *testing.T) {
	bucketName := "test-nuon-uploads"
	bucketPrefix := "test-prefix"
	creds := generics.GetFakeObj[*credentials.Config]()

	tests := map[string]struct {
		initFn      func() (*s3Uploader, error)
		assertFn    func(*testing.T, Uploader)
		errExpected error
	}{
		"errors with no options": {
			initFn: func() (*s3Uploader, error) {
				return NewS3Uploader(validator.New())
			},
			assertFn: func(t *testing.T, obj Uploader) {
				s3Obj, ok := obj.(*s3Uploader)
				assert.True(t, ok)
				assert.Equal(t, bucketName, s3Obj.Bucket)
			},
			errExpected: fmt.Errorf("Bucket"),
		},
		"initializes with bucket": {
			initFn: func() (*s3Uploader, error) {
				return NewS3Uploader(validator.New(),
					WithBucketName(bucketName))
			},
			assertFn: func(t *testing.T, obj Uploader) {
				s3Obj, ok := obj.(*s3Uploader)
				assert.True(t, ok)
				assert.Equal(t, bucketName, s3Obj.Bucket)
			},
		},
		"sets prefix correctly": {
			initFn: func() (*s3Uploader, error) {
				return NewS3Uploader(validator.New(),
					WithBucketName(bucketName),
					WithPrefix(bucketPrefix))
			},
			assertFn: func(t *testing.T, obj Uploader) {
				s3Obj, ok := obj.(*s3Uploader)
				assert.True(t, ok)
				assert.Equal(t, bucketPrefix, s3Obj.prefix)
			},
		},
		"sets assume role correctly": {
			initFn: func() (*s3Uploader, error) {
				return NewS3Uploader(validator.New(),
					WithBucketName(bucketName),
					WithAssumeRoleARN("test-role-arn"))
			},
			assertFn: func(t *testing.T, obj Uploader) {
				s3Obj, ok := obj.(*s3Uploader)
				assert.True(t, ok)
				assert.Equal(t, bucketName, s3Obj.Bucket)
				assert.Equal(t, "test-role-arn", s3Obj.assumeRoleARN)
			},
		},
		"sets credentials": {
			initFn: func() (*s3Uploader, error) {
				return NewS3Uploader(validator.New(),
					WithBucketName(bucketName),
					WithCredentials(creds))
			},
			assertFn: func(t *testing.T, obj Uploader) {
				s3Obj, ok := obj.(*s3Uploader)
				assert.True(t, ok)
				assert.Equal(t, creds, s3Obj.creds)
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			uploader, err := test.initFn()
			if test.errExpected != nil {
				assert.NotNil(t, err)
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}

			test.assertFn(t, uploader)
		})
	}
}

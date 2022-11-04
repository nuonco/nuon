package terraform

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/mholt/archiver/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockModuleFetcher struct {
	mock.Mock
}

func (m *mockModuleFetcher) createTmpDir(installID string) (string, error) {
	args := m.Called(installID)
	return args.String(0), args.Error(1)
}

func (m *mockModuleFetcher) cleanupTmpDir(path string) error {
	args := m.Called(path)
	return args.Error(0)
}

func (m *mockModuleFetcher) fetchModule(ctx context.Context, module Module, tmpDir string) error {
	args := m.Called(ctx, module, tmpDir)
	return args.Error(0)
}

var _ moduleFetcher = (*mockModuleFetcher)(nil)

func TestModuleFetcherCreateTmpDir(t *testing.T) {
	tests := map[string]struct {
		fn func(*s3ModuleFetcher)
	}{
		"returns a temporary directory": {
			fn: func(s *s3ModuleFetcher) {
				tmpDir, err := s.createTmpDir("install-id")
				assert.Nil(t, err)
				assert.NotNil(t, tmpDir)

				stat, err := os.Stat(tmpDir)
				assert.Nil(t, err)
				assert.True(t, stat.IsDir())
			},
		},

		"does not create a duplicate directory": {
			fn: func(s *s3ModuleFetcher) {
				tmpDir, err := s.createTmpDir("install-id")
				assert.Nil(t, err)
				assert.NotNil(t, tmpDir)

				secondTmpDir, err := s.createTmpDir("install-id")
				assert.Nil(t, err)
				assert.NotNil(t, tmpDir)

				assert.NotEqual(t, tmpDir, secondTmpDir)
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			obj := &s3ModuleFetcher{}
			test.fn(obj)
		})
	}
}

func TestModuleFetcherCleanupDir(t *testing.T) {
	tests := map[string]struct {
		fn func(*s3ModuleFetcher)
	}{
		"cleans up a temporary directory": {
			fn: func(s *s3ModuleFetcher) {
				tmpDir, err := s.createTmpDir("install-id")
				assert.Nil(t, err)
				assert.NotNil(t, tmpDir)

				err = s.cleanupTmpDir(tmpDir)
				assert.Nil(t, err)

				stat, err := os.Stat(tmpDir)
				assert.Nil(t, stat)
				assert.NotNil(t, err)
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			obj := &s3ModuleFetcher{}
			test.fn(obj)
		})
	}
}

func TestModule_getS3Key(t *testing.T) {
	obj := &s3ModuleFetcher{}
	s3Key := obj.getS3Key("foo", "v_0.0.1")
	assert.Equal(t, s3Key, "modulees/foo_v_0.0.1.tar.gz")
}

type mockDownloaderClient func(
	ctx context.Context,
	writer io.WriterAt,
	in *s3.GetObjectInput,
	fns ...func(*manager.Downloader),
) (int64, error)

func (m mockDownloaderClient) Download(
	ctx context.Context,
	writer io.WriterAt,
	in *s3.GetObjectInput,
	fns ...func(*manager.Downloader),
) (int64, error) {
	return m(ctx, writer, in, fns...)
}

var errNotFound = errors.New("not found")

func TestModule_downloadModule(t *testing.T) {
	module := Module{
		BucketName: "nuon-test-modules",
		BucketKey:  "sandboxes/foobar/v0.0.0.1",
	}
	b := &s3ModuleFetcher{}

	tests := map[string]struct {
		api         func(*testing.T) mockDownloaderClient
		cb          func(*testing.T, []byte)
		errExpected error
	}{
		"object not found": {
			errExpected: errNotFound,
			api: func(t *testing.T) mockDownloaderClient {
				return func(ctx context.Context, writer io.WriterAt, in *s3.GetObjectInput, fns ...func(*manager.Downloader)) (int64, error) {
					return 0, errNotFound
				}
			},
		},

		"creates proper key for the module": {
			api: func(t *testing.T) mockDownloaderClient {
				return func(ctx context.Context, writer io.WriterAt, in *s3.GetObjectInput, fns ...func(*manager.Downloader)) (int64, error) {
					assert.Equal(t, module.BucketKey, *in.Key)
					return 0, nil
				}
			},
		},
		"passes proper bucket for module": {
			api: func(t *testing.T) mockDownloaderClient {
				return func(ctx context.Context, writer io.WriterAt, in *s3.GetObjectInput, fns ...func(*manager.Downloader)) (int64, error) {
					assert.Equal(t, module.BucketName, *in.Bucket)
					return 0, nil
				}
			},
		},

		"successfully returns a valid module archive": {
			api: func(t *testing.T) mockDownloaderClient {
				return func(ctx context.Context, writer io.WriterAt, in *s3.GetObjectInput, fns ...func(*manager.Downloader)) (int64, error) {
					writer.WriteAt([]byte("abc"), 0) //nolint:errcheck
					return 0, nil
				}
			},
			cb: func(t *testing.T, resp []byte) {
				assert.Equal(t, []byte("abc"), resp)
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			api := test.api(t)
			resp, err := b.downloadModule(context.Background(), module, api)
			if test.errExpected != nil {
				assert.ErrorIs(t, err, test.errExpected)
				return
			}

			if test.cb != nil {
				test.cb(t, resp)
			}
		})
	}
}

func newTmpFile(t *testing.T, content string) string {
	tmpFile, err := os.CreateTemp(t.TempDir(), "test-*.tf")
	assert.Nil(t, err)
	defer tmpFile.Close()

	_, err = tmpFile.Write([]byte(content))
	assert.Nil(t, err)
	return tmpFile.Name()
}

func TestModule_extractModule(t *testing.T) {
	b := &s3ModuleFetcher{}

	tests := map[string]struct {
		archive     func(*testing.T) []byte
		cb          func(*testing.T, string)
		errExpected error
	}{
		"properly writes files from archive into temp dir": {
			archive: func(t *testing.T) []byte {
				ctx := context.Background()
				tmpFile := newTmpFile(t, "test\n")
				files, err := archiver.FilesFromDisk(nil, map[string]string{
					tmpFile: "test.tf",
				})
				assert.Nil(t, err)
				assert.NotEmpty(t, files)

				var b bytes.Buffer
				compressedWriter, err := archiver.Gz{}.OpenWriter(&b)
				assert.Nil(t, err)

				err = archiver.Tar{}.Archive(ctx, compressedWriter, files)
				assert.Nil(t, err)

				err = os.Remove(tmpFile)
				assert.Nil(t, err)

				compressedWriter.Close()

				return b.Bytes()
			},
			cb: func(t *testing.T, tmpDir string) {
				path := filepath.Join(tmpDir, "test.tf")
				byts, err := os.ReadFile(path)
				assert.Nil(t, err)
				assert.NotNil(t, byts)
				assert.Equal(t, []byte("test\n"), byts)
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			tmpDir := t.TempDir()
			archive := test.archive(t)
			err := b.extractModule(context.Background(), tmpDir, archive)
			if test.errExpected != nil {
				assert.ErrorIs(t, err, test.errExpected)
				return
			}

			if test.cb != nil {
				test.cb(t, tmpDir)
			}
		})
	}
}

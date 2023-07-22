package dir

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/powertoolsdev/mono/pkg/terraform/archive"
	"github.com/stretchr/testify/assert"
)

func Test_oci_unpackDir(t *testing.T) {
	errUnpackDir := fmt.Errorf("error unpacking directory")

	tests := map[string]struct {
		dirFn       func(t *testing.T) string
		callbackFn  func(mockCtl *gomock.Controller) archive.Callback
		errExpected error
	}{
		"happy path": {
			dirFn: func(t *testing.T) string {
				tmpDir := t.TempDir()
				fp := filepath.Join(tmpDir, "test.txt")
				err := os.WriteFile(fp, []byte("hello world"), 0600)
				assert.NoError(t, err)
				return tmpDir
			},
			callbackFn: func(mockCtl *gomock.Controller) archive.Callback {
				mock := archive.NewMockCallbacker(mockCtl)
				mock.EXPECT().Callback(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, path string, rc io.ReadCloser) error {
					assert.Equal(t, "test.txt", path)

					byts, err := io.ReadAll(rc)
					assert.NoError(t, err)
					assert.Equal(t, byts, []byte("hello world"))
					return nil
				})
				return mock.Callback
			},
		},
		"happy path - dir": {
			dirFn: func(t *testing.T) string {
				tmpDir := t.TempDir()
				fp := filepath.Join(tmpDir, "data/test.txt")
				err := os.MkdirAll(filepath.Dir(fp), 0744)
				assert.NoError(t, err)

				err = os.WriteFile(fp, []byte("hello world"), 0600)
				assert.NoError(t, err)

				return tmpDir
			},
			callbackFn: func(mockCtl *gomock.Controller) archive.Callback {
				mock := archive.NewMockCallbacker(mockCtl)
				mock.EXPECT().Callback(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, path string, rc io.ReadCloser) error {
					assert.Equal(t, "data/test.txt", path)

					byts, err := io.ReadAll(rc)
					assert.NoError(t, err)
					assert.Equal(t, byts, []byte("hello world"))
					return nil
				})
				return mock.Callback
			},
		},
		"error": {
			dirFn: func(t *testing.T) string {
				tmpDir := t.TempDir()
				fp := filepath.Join(tmpDir, "test.txt")
				err := os.WriteFile(fp, []byte("hello world"), 0600)
				assert.NoError(t, err)
				return tmpDir
			},
			callbackFn: func(mockCtl *gomock.Controller) archive.Callback {
				mock := archive.NewMockCallbacker(mockCtl)
				mock.EXPECT().Callback(gomock.Any(), gomock.Any(), gomock.Any()).Return(errUnpackDir)
				return mock.Callback
			},
			errExpected: errUnpackDir,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			mockCtl, ctx := gomock.WithContext(ctx, t)

			tmpDir := test.dirFn(t)
			obj := &dir{
				Path: tmpDir,
			}

			cb := test.callbackFn(mockCtl)

			err := obj.Unpack(ctx, cb)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}

			assert.NoError(t, err)
		})
	}
}

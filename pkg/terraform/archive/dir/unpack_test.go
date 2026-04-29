package dir

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/nuonco/nuon/pkg/terraform/archive"
	"github.com/stretchr/testify/assert"
)

func Test_oci_unpackDir(t *testing.T) {
	errUnpackDir := fmt.Errorf("error unpacking directory")

	tests := map[string]struct {
		dirFn                 func(t *testing.T) string
		ignoreDotTerraformDir bool
		callbackFn            func(mockCtl *gomock.Controller) archive.Callback
		errExpected           error
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
		"ignore .terraform but keep .terraform/modules": {
			dirFn: func(t *testing.T) string {
				tmpDir := t.TempDir()

				// Files we expect to survive the filter:
				//   - main.tf (root source)
				//   - .terraform/modules/foo/main.tf (vendored module)
				// Files we expect to be filtered out:
				//   - .terraform/providers/registry.terraform.io/.../terraform-provider-aws
				mustWrite := func(rel, body string) {
					fp := filepath.Join(tmpDir, rel)
					assert.NoError(t, os.MkdirAll(filepath.Dir(fp), 0o755))
					assert.NoError(t, os.WriteFile(fp, []byte(body), 0o600))
				}
				mustWrite("main.tf", "main")
				mustWrite(".terraform/modules/foo/main.tf", "module")
				mustWrite(".terraform/providers/registry.terraform.io/hashicorp/aws/x", "binary")

				return tmpDir
			},
			ignoreDotTerraformDir: true,
			callbackFn: func(mockCtl *gomock.Controller) archive.Callback {
				mock := archive.NewMockCallbacker(mockCtl)
				seen := map[string]string{}
				mock.EXPECT().Callback(gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, path string, rc io.ReadCloser) error {
						byts, err := io.ReadAll(rc)
						assert.NoError(t, err)
						seen[path] = string(byts)
						return nil
					}).AnyTimes()
				// Validation runs after Unpack returns; assert via t.Cleanup.
				t.Cleanup(func() {
					assert.Equal(t, "main", seen["main.tf"])
					assert.Equal(t, "module", seen[".terraform/modules/foo/main.tf"])
					_, providerSeen := seen[".terraform/providers/registry.terraform.io/hashicorp/aws/x"]
					assert.False(t, providerSeen, "expected .terraform/providers to be filtered")
				})
				return mock.Callback
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			mockCtl, ctx := gomock.WithContext(ctx, t)

			tmpDir := test.dirFn(t)
			obj := &dir{
				Path:                  tmpDir,
				IgnoreDotTerraformDir: test.ignoreDotTerraformDir,
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

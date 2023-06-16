package workspace

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"
	gomock "github.com/golang/mock/gomock"
	"github.com/powertoolsdev/mono/pkg/terraform/archive"
	"github.com/powertoolsdev/mono/pkg/terraform/backend"
	"github.com/powertoolsdev/mono/pkg/terraform/binary"
	"github.com/powertoolsdev/mono/pkg/terraform/variables"
	"github.com/stretchr/testify/assert"
)

func Test_LoadArchive(t *testing.T) {
	back := backend.NewMockBackend(nil)
	vars := variables.NewMockVariables(nil)
	bin := binary.NewMockBinary(nil)

	v := validator.New()
	errLoadArchive := fmt.Errorf("error")

	tests := map[string]struct {
		archiveFn   func(*gomock.Controller) archive.Archive
		assertFn    func(*testing.T, *workspace)
		errExpected error
	}{
		"happy path": {
			archiveFn: func(mockCtl *gomock.Controller) archive.Archive {
				mock := archive.NewMockArchive(mockCtl)

				mock.EXPECT().Init(gomock.Any()).Return(nil)
				mock.EXPECT().Unpack(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, cb archive.Callback) error {
					r := io.NopCloser(strings.NewReader("hello world"))
					err := cb(ctx, "test.txt", r)
					assert.NoError(t, err)
					return nil
				})
				return mock
			},
			assertFn: func(t *testing.T, w *workspace) {
				fp := filepath.Join(w.root, "test.txt")
				stat, err := os.Stat(fp)
				assert.NoError(t, err)
				assert.Equal(t, stat.Mode(), defaultFilePermissions)
			},
			errExpected: nil,
		},
		"happy path - creates directory": {
			archiveFn: func(mockCtl *gomock.Controller) archive.Archive {
				mock := archive.NewMockArchive(mockCtl)

				mock.EXPECT().Init(gomock.Any()).Return(nil)
				mock.EXPECT().Unpack(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, cb archive.Callback) error {
					r := io.NopCloser(strings.NewReader("hello world"))
					err := cb(ctx, "modules/test.txt", r)
					assert.NoError(t, err)
					return nil
				})
				return mock
			},
			assertFn: func(t *testing.T, w *workspace) {
				fp := filepath.Join(w.root, "modules/test.txt")
				stat, err := os.Stat(fp)
				assert.NoError(t, err)
				assert.Equal(t, stat.Mode(), defaultFilePermissions)
			},
			errExpected: nil,
		},
		"error on init": {
			archiveFn: func(mockCtl *gomock.Controller) archive.Archive {
				mock := archive.NewMockArchive(mockCtl)

				mock.EXPECT().Init(gomock.Any()).Return(errLoadArchive)
				return mock
			},
			errExpected: fmt.Errorf("unable to initialize archive"),
		},
		"error on unpack": {
			archiveFn: func(mockCtl *gomock.Controller) archive.Archive {
				mock := archive.NewMockArchive(mockCtl)

				mock.EXPECT().Init(gomock.Any()).Return(nil)
				mock.EXPECT().Unpack(gomock.Any(), gomock.Any()).Return(errLoadArchive)
				return mock
			},
			errExpected: fmt.Errorf("unable to unpack archive"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			mockCtl := gomock.NewController(t)
			arch := test.archiveFn(mockCtl)

			wkspace, err := New(v,
				WithArchive(arch),
				WithBackend(back),
				WithBinary(bin),
				WithVariables(vars),
			)
			assert.NoError(t, err)

			err = wkspace.InitRoot(ctx)
			assert.NoError(t, err)

			err = wkspace.LoadArchive(ctx)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}

			assert.NoError(t, err)
			test.assertFn(t, wkspace)
		})
	}
}

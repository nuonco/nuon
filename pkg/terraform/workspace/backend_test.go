package workspace

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-playground/validator/v10"
	gomock "github.com/golang/mock/gomock"
	"github.com/powertoolsdev/mono/pkg/terraform/archive"
	"github.com/powertoolsdev/mono/pkg/terraform/backend"
	"github.com/powertoolsdev/mono/pkg/terraform/binary"
	"github.com/powertoolsdev/mono/pkg/terraform/hooks"
	"github.com/powertoolsdev/mono/pkg/terraform/variables"
	"github.com/stretchr/testify/assert"
)

func Test_LoadBackend(t *testing.T) {
	arch := archive.NewMockArchive(nil)
	vars := variables.NewMockVariables(nil)
	bin := binary.NewMockBinary(nil)
	hooks := hooks.NewMockHooks(nil)

	v := validator.New()
	errLoadBackend := fmt.Errorf("error")

	tests := map[string]struct {
		backendFn   func(*gomock.Controller) backend.Backend
		assertFn    func(*testing.T, *workspace)
		errExpected error
	}{
		"happy path": {
			backendFn: func(mockCtl *gomock.Controller) backend.Backend {
				mock := backend.NewMockBackend(mockCtl)

				mock.EXPECT().Init(gomock.Any()).Return(nil)
				mock.EXPECT().ConfigFile(gomock.Any()).DoAndReturn(func(ctx context.Context) ([]byte, error) {
					return []byte("{}"), nil
				})
				return mock
			},
			assertFn: func(t *testing.T, w *workspace) {
				fp := filepath.Join(w.root, defaultBackendConfigFilename)
				stat, err := os.Stat(fp)
				assert.NoError(t, err)
				assert.Equal(t, stat.Mode(), defaultFilePermissions)

				byts, err := os.ReadFile(fp)
				assert.NoError(t, err)
				assert.Equal(t, byts, []byte("{}"))
			},
			errExpected: nil,
		},
		"error on init": {
			backendFn: func(mockCtl *gomock.Controller) backend.Backend {
				mock := backend.NewMockBackend(mockCtl)

				mock.EXPECT().Init(gomock.Any()).Return(errLoadBackend)
				return mock
			},
			errExpected: fmt.Errorf("unable to initialize backend"),
		},
		"error on config file": {
			backendFn: func(mockCtl *gomock.Controller) backend.Backend {
				mock := backend.NewMockBackend(mockCtl)

				mock.EXPECT().Init(gomock.Any()).Return(nil)
				mock.EXPECT().ConfigFile(gomock.Any()).Return(nil, errLoadBackend)
				return mock
			},
			errExpected: fmt.Errorf("unable to get config file"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			mockCtl, ctx := gomock.WithContext(ctx, t)
			defer mockCtl.Finish()
			back := test.backendFn(mockCtl)

			wkspace, err := New(v,
				WithArchive(arch),
				WithHooks(hooks),
				WithBackend(back),
				WithBinary(bin),
				WithVariables(vars),
			)
			assert.NoError(t, err)

			err = wkspace.InitRoot(ctx)
			assert.NoError(t, err)

			err = wkspace.LoadBackend(ctx)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}

			assert.NoError(t, err)
			test.assertFn(t, wkspace)
		})
	}
}

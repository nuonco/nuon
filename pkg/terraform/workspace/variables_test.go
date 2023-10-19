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

func Test_LoadVariables(t *testing.T) {
	arch := archive.NewMockArchive(nil)
	back := backend.NewMockBackend(nil)
	bin := binary.NewMockBinary(nil)
	hooks := hooks.NewMockHooks(nil)

	v := validator.New()
	errLoadVariables := fmt.Errorf("error")

	tests := map[string]struct {
		variablesFn func(*gomock.Controller) variables.Variables
		assertFn    func(*testing.T, *workspace)
		errExpected error
	}{
		"happy path": {
			variablesFn: func(mockCtl *gomock.Controller) variables.Variables {
				mock := variables.NewMockVariables(mockCtl)

				mock.EXPECT().Init(gomock.Any()).Return(nil)
				mock.EXPECT().GetEnv(gomock.Any()).Return(map[string]string{"key": "value"}, nil)
				mock.EXPECT().GetFile(gomock.Any()).Return([]byte("{}"), nil)
				return mock
			},
			assertFn: func(t *testing.T, w *workspace) {
				assert.Equal(t, w.envVars["key"], "value")

				fp := filepath.Join(w.root, defaultVariablesFilename)
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
			variablesFn: func(mockCtl *gomock.Controller) variables.Variables {
				mock := variables.NewMockVariables(mockCtl)

				mock.EXPECT().Init(gomock.Any()).Return(errLoadVariables)
				return mock
			},
			errExpected: fmt.Errorf("unable to init variables"),
		},
		"error on get env": {
			variablesFn: func(mockCtl *gomock.Controller) variables.Variables {
				mock := variables.NewMockVariables(mockCtl)

				mock.EXPECT().Init(gomock.Any()).Return(nil)
				mock.EXPECT().GetEnv(gomock.Any()).Return(nil, errLoadVariables)
				return mock
			},
			errExpected: fmt.Errorf("unable to get env variables"),
		},
		"error on get file": {
			variablesFn: func(mockCtl *gomock.Controller) variables.Variables {
				mock := variables.NewMockVariables(mockCtl)

				mock.EXPECT().Init(gomock.Any()).Return(nil)
				mock.EXPECT().GetEnv(gomock.Any()).Return(map[string]string{}, nil)
				mock.EXPECT().GetFile(gomock.Any()).Return(nil, errLoadVariables)
				return mock
			},
			errExpected: fmt.Errorf("unable to get file variables"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			mockCtl, ctx := gomock.WithContext(ctx, t)
			defer mockCtl.Finish()
			vars := test.variablesFn(mockCtl)

			wkspace, err := New(v,
				WithArchive(arch),
				WithHooks(hooks),
				WithVariables(vars),
				WithBinary(bin),
				WithBackend(back),
			)
			assert.NoError(t, err)

			err = wkspace.InitRoot(ctx)
			assert.NoError(t, err)

			err = wkspace.LoadVariables(ctx)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}

			assert.NoError(t, err)
			test.assertFn(t, wkspace)
		})
	}
}

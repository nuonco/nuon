package workspace

import (
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/go-playground/validator/v10"
	gomock "github.com/golang/mock/gomock"
	"github.com/hashicorp/go-hclog"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/terraform/archive"
	"github.com/powertoolsdev/mono/pkg/terraform/backend"
	"github.com/powertoolsdev/mono/pkg/terraform/binary"
	"github.com/powertoolsdev/mono/pkg/terraform/hooks"
	"github.com/powertoolsdev/mono/pkg/terraform/variables"
	"github.com/stretchr/testify/assert"
)

func Test_LoadBinary(t *testing.T) {
	v := validator.New()

	arch := archive.NewMockArchive(nil)
	vars := variables.NewMockVariables(nil)
	back := backend.NewMockBackend(nil)
	hooks := hooks.NewMockHooks(nil)

	errLoadBinary := fmt.Errorf("error")
	execPath := generics.GetFakeObj[string]()

	lg := hclog.New(&hclog.LoggerOptions{
		Output: io.Discard,
	})

	tests := map[string]struct {
		binaryFn    func(*gomock.Controller) binary.Binary
		assertFn    func(*testing.T, *workspace)
		errExpected error
	}{
		"happy path": {
			binaryFn: func(mockCtl *gomock.Controller) binary.Binary {
				mock := binary.NewMockBinary(mockCtl)

				mock.EXPECT().Init(gomock.Any()).Return(nil)
				mock.EXPECT().Install(
					gomock.Any(),
					gomock.Any(),
					gomock.Any()).DoAndReturn(
					func(ctx context.Context, lg hclog.Logger, path string) (string, error) {
						return execPath, nil
					})
				return mock
			},
			assertFn: func(t *testing.T, w *workspace) {
				assert.Equal(t, w.execPath, execPath)
			},
			errExpected: nil,
		},
		"error on init": {
			binaryFn: func(mockCtl *gomock.Controller) binary.Binary {
				mock := binary.NewMockBinary(mockCtl)

				mock.EXPECT().Init(gomock.Any()).Return(errLoadBinary)
				return mock
			},
			errExpected: fmt.Errorf("unable to initialize binary"),
		},
		"error on install": {
			binaryFn: func(mockCtl *gomock.Controller) binary.Binary {
				mock := binary.NewMockBinary(mockCtl)

				mock.EXPECT().Init(gomock.Any()).Return(nil)
				mock.EXPECT().Install(gomock.Any(), gomock.Any(), gomock.Any()).Return("", errLoadBinary)
				return mock
			},
			errExpected: fmt.Errorf("unable to install binary"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			mockCtl, ctx := gomock.WithContext(ctx, t)
			defer mockCtl.Finish()
			bin := test.binaryFn(mockCtl)

			wkspace, err := New(v,
				WithHooks(hooks),
				WithArchive(arch),
				WithBinary(bin),
				WithBackend(back),
				WithVariables(vars),
			)
			assert.NoError(t, err)

			err = wkspace.InitRoot(ctx)
			assert.NoError(t, err)

			err = wkspace.LoadBinary(ctx, lg)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}

			assert.NoError(t, err)
			test.assertFn(t, wkspace)
		})
	}
}

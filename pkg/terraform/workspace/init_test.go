package workspace

import (
	"context"
	"os"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/terraform/archive"
	"github.com/powertoolsdev/mono/pkg/terraform/backend"
	"github.com/powertoolsdev/mono/pkg/terraform/binary"
	"github.com/powertoolsdev/mono/pkg/terraform/variables"
	"github.com/stretchr/testify/assert"
)

func Test_workspace_Init(t *testing.T) {
	v := validator.New()

	back := backend.NewMockBackend(nil)
	vars := variables.NewMockVariables(nil)
	bin := binary.NewMockBinary(nil)
	arch := archive.NewMockArchive(nil)

	tests := map[string]struct {
		workspaceFn func(*testing.T) *workspace
		assertFn    func(*testing.T, *workspace)
		errExpected error
	}{
		"happy path": {
			workspaceFn: func(t *testing.T) *workspace {
				wkspace, err := New(v, WithArchive(arch),
					WithBackend(back),
					WithVariables(vars),
					WithBinary(bin),
				)
				assert.NoError(t, err)

				return wkspace
			},
			assertFn: func(t *testing.T, w *workspace) {
				assert.NotEmpty(t, w.root)
				stat, err := os.Stat(w.root)
				assert.NoError(t, err)
				assert.True(t, stat.IsDir())
			},
			errExpected: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()

			wkspace := test.workspaceFn(t)
			err := wkspace.InitRoot(ctx)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}

			assert.NoError(t, err)
			test.assertFn(t, wkspace)
		})
	}
}

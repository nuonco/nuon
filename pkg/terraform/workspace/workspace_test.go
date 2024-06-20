package workspace

import (
	"fmt"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/terraform/archive"
	"github.com/powertoolsdev/mono/pkg/terraform/backend"
	"github.com/powertoolsdev/mono/pkg/terraform/binary"
	"github.com/powertoolsdev/mono/pkg/terraform/hooks"
	"github.com/powertoolsdev/mono/pkg/terraform/variables"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	return
	v := validator.New()

	arch := archive.NewMockArchive(nil)
	back := backend.NewMockBackend(nil)
	vars := variables.NewMockVariables(nil)
	bin := binary.NewMockBinary(nil)
	hooks := hooks.NewMockHooks(nil)

	tests := map[string]struct {
		errExpected error
		optsFn      func() []workspaceOption
		assertFn    func(*testing.T, *workspace)
	}{
		"happy path": {
			optsFn: func() []workspaceOption {
				return []workspaceOption{
					WithBackend(back),
					WithArchive(arch),
					WithVariables(vars),
					WithHooks(hooks),
					WithBinary(bin),
				}
			},
			assertFn: func(t *testing.T, s *workspace) {
				assert.Equal(t, back, s.Backend)
				assert.Equal(t, vars, s.Variables)
				assert.Equal(t, arch, s.Archive)
				assert.Equal(t, bin, s.Binary)
				assert.False(t, s.DisableCleanup)
			},
		},
		"disable cleanup": {
			optsFn: func() []workspaceOption {
				return []workspaceOption{
					WithBackend(back),
					WithArchive(arch),
					WithVariables(vars),
					WithHooks(hooks),
					WithBinary(bin),
					WithDisableCleanup(true),
				}
			},
			assertFn: func(t *testing.T, s *workspace) {
				assert.Equal(t, back, s.Backend)
				assert.Equal(t, vars, s.Variables)
				assert.Equal(t, arch, s.Archive)
				assert.Equal(t, bin, s.Binary)
				assert.True(t, s.DisableCleanup)
			},
		},
		"missing archive": {
			optsFn: func() []workspaceOption {
				return []workspaceOption{
					WithBackend(back),
					WithVariables(vars),
					WithHooks(hooks),
					WithBinary(bin),
				}
			},
			errExpected: fmt.Errorf("Archive"),
		},
		"missing backend": {
			optsFn: func() []workspaceOption {
				return []workspaceOption{
					WithArchive(arch),
					WithVariables(vars),
					WithHooks(hooks),
					WithBinary(bin),
				}
			},
			errExpected: fmt.Errorf("Backend"),
		},
		"missing binary": {
			optsFn: func() []workspaceOption {
				return []workspaceOption{
					WithBackend(back),
					WithArchive(arch),
					WithHooks(hooks),
					WithVariables(vars),
				}
			},
			errExpected: fmt.Errorf("Binary"),
		},
		"missing hooks": {
			optsFn: func() []workspaceOption {
				return []workspaceOption{
					WithBackend(back),
					WithArchive(arch),
					WithBinary(bin),
					WithVariables(vars),
				}
			},
			errExpected: fmt.Errorf("Hooks"),
		},
		"missing variables": {
			optsFn: func() []workspaceOption {
				return []workspaceOption{
					WithBackend(back),
					WithArchive(arch),
					WithHooks(hooks),
					WithBinary(bin),
				}
			},
			errExpected: fmt.Errorf("Variables"),
		},
	}

	for name, test := range tests {
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			e, err := New(v, test.optsFn()...)
			if test.errExpected != nil {
				assert.Error(t, err)
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
			test.assertFn(t, e)
		})
	}
}

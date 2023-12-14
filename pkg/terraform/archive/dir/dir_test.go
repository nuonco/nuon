package dir

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	expected := generics.GetFakeObj[dir]()
	v := validator.New()

	tests := map[string]struct {
		errExpected error
		optsFn      func() []dirOption
		assertFn    func(*testing.T, *dir)
	}{
		"happy path": {
			optsFn: func() []dirOption {
				return []dirOption{
					WithPath(expected.Path),
				}
			},
			assertFn: func(t *testing.T, s *dir) {
				absPath, err := filepath.Abs(expected.Path)
				require.NoError(t, err)
				assert.Equal(t, absPath, s.Path)
			},
		},
		"missing path": {
			optsFn: func() []dirOption {
				return []dirOption{}
			},
			errExpected: fmt.Errorf("Path"),
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

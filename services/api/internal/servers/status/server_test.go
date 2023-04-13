package status

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	gitRef := generics.GetFakeObj[string]()

	tests := map[string]struct {
		optFns      func() []serverOption
		errExpected error
		assertFn    func(*testing.T, *server)
	}{
		"happy path": {
			optFns: func() []serverOption {
				return []serverOption{
					WithGitRef(gitRef),
				}
			},
			assertFn: func(t *testing.T, mgr *server) {
				assert.Equal(t, gitRef, mgr.GitRef)
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			v := validator.New()
			opts := test.optFns()
			commands, err := New(v, opts...)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
			test.assertFn(t, commands)
		})
	}
}

package static

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	v := validator.New()

	tests := map[string]struct {
		errExpected error
		optsFn      func() []varsOption
		assertFn    func(*testing.T, *vars)
	}{
		"happy path": {
			optsFn: func() []varsOption {
				return []varsOption{}
			},
			assertFn: func(t *testing.T, s *vars) {
			},
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

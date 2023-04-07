package activities

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	tests := map[string]struct {
		optFns      func() []activitiesOption
		assertFn    func(*testing.T, *Activities)
		errExpected error
	}{
		"happy path": {
			optFns: func() []activitiesOption {
				return []activitiesOption{}
			},
			assertFn: func(t *testing.T, a *Activities) {
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			v := validator.New()
			opts := test.optFns()
			r, err := New(v, opts...)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
			test.assertFn(t, r)
		})
	}
}

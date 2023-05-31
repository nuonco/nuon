package oci

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	expected := generics.GetFakeObj[oci]()
	v := validator.New()

	tests := map[string]struct {
		errExpected error
		optsFn      func() []ociOption
		assertFn    func(*testing.T, *oci)
	}{
		"happy path": {
			optsFn: func() []ociOption {
				return []ociOption{
					WithRoleSessionName(expected.RoleSessionName),
					WithRoleARN(expected.RoleARN),
				}
			},
			assertFn: func(t *testing.T, s *oci) {
				assert.Equal(t, expected.RoleARN, s.RoleARN)
				assert.Equal(t, expected.RoleSessionName, s.RoleSessionName)
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

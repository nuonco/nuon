package remote

import (
	"fmt"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/opentofu/tofudl"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	v := validator.New()

	tests := map[string]struct {
		errExpected error
		optsFn      func() []remoteOption
		assertFn    func(*testing.T, *remote)
	}{
		"happy path": {
			optsFn: func() []remoteOption {
				return []remoteOption{
					WithVersion("1.13.9"),
				}
			},
			assertFn: func(t *testing.T, r *remote) {
				assert.Equal(t, tofudl.Version("1.13.9"), r.Version)
			},
		},
		"v-prefixed version": {
			optsFn: func() []remoteOption {
				return []remoteOption{
					WithVersion("v1.13.9"),
				}
			},
			assertFn: func(t *testing.T, r *remote) {
				assert.Equal(t, tofudl.Version("1.13.9"), r.Version)
			},
		},
		"missing version": {
			optsFn: func() []remoteOption {
				return []remoteOption{}
			},
			errExpected: fmt.Errorf("Version"),
		},
		"invalid version": {
			optsFn: func() []remoteOption {
				return []remoteOption{
					WithVersion("abc"),
				}
			},
			errExpected: fmt.Errorf("invalid version"),
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

func TestSource(t *testing.T) {
	r, err := New(validator.New(), WithVersion("1.13.9"))
	assert.NoError(t, err)
	assert.Equal(t, "remote", r.Source())
}

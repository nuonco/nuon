package remote

import (
	"fmt"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	v := validator.New()
	version := "v0.0.1"

	tests := map[string]struct {
		errExpected error
		optsFn      func() []remoteOption
		assertFn    func(*testing.T, *remote)
	}{
		"happy path": {
			optsFn: func() []remoteOption {
				return []remoteOption{
					WithVersion(version),
				}
			},
			assertFn: func(t *testing.T, r *remote) {
				assert.Equal(t, r.Version.Original(), version)
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

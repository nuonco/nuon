package catalog

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/aws/credentials"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	v := validator.New()
	creds := generics.GetFakeObj[*credentials.Config]()

	tests := map[string]struct {
		errExpected error
		optsFn      func() []catalogOption
		assertFn    func(*testing.T, *catalog)
	}{
		"happy path": {
			optsFn: func() []catalogOption {
				return []catalogOption{
					WithCredentials(creds),
				}
			},
			assertFn: func(t *testing.T, e *catalog) {
				assert.Equal(t, creds, e.Credentials)
			},
		},
		"happy path - override": {
			optsFn: func() []catalogOption {
				return []catalogOption{
					WithCredentials(creds),
					WithDevOverride(true),
				}
			},
			assertFn: func(t *testing.T, e *catalog) {
				assert.True(t, e.DevOverride)
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

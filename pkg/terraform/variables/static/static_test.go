package static

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	v := validator.New()

	envVars := generics.GetFakeObj[map[string]string]()
	fileVars := map[string]interface{}{
		"key": "value",
		"map": map[string]interface{}{
			"key": "value",
		},
	}

	tests := map[string]struct {
		errExpected error
		optsFn      func() []varsOption
		assertFn    func(*testing.T, *vars)
	}{
		"happy path": {
			optsFn: func() []varsOption {
				return []varsOption{
					WithEnvVars(envVars),
					WithFileVars(fileVars),
				}
			},
			assertFn: func(t *testing.T, v *vars) {
				assert.Equal(t, envVars, v.EnvVars)
				assert.Equal(t, fileVars, v.FileVars)
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

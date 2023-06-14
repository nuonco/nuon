package static

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/stretchr/testify/assert"
)

func Test_vars_GetEnv(t *testing.T) {
	v := validator.New()
	envVars := generics.GetFakeObj[map[string]string]()

	tests := map[string]struct {
		varsFn   func(*testing.T) *vars
		assertFn func(*testing.T, map[string]string)
	}{
		"happy path": {
			varsFn: func(t *testing.T) *vars {
				vs, err := New(v, WithEnvVars(envVars))
				assert.NoError(t, err)
				return vs
			},
			assertFn: func(t *testing.T, res map[string]string) {
				assert.Equal(t, envVars, res)
			},
		},
		"no env vars found": {
			varsFn: func(t *testing.T) *vars {
				vs, err := New(v)
				assert.NoError(t, err)
				return vs
			},
			assertFn: func(t *testing.T, res map[string]string) {
				assert.Nil(t, res)
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			ctx, cancel := context.WithCancel(ctx)
			defer cancel()

			vs := test.varsFn(t)
			res, err := vs.GetEnv(ctx)
			assert.NoError(t, err)
			test.assertFn(t, res)
		})
	}
}

func Test_vars_FileVars(t *testing.T) {
	v := validator.New()
	fileVars := map[string]interface{}{
		"key": "value",
		"map": map[string]interface{}{
			"key": "value",
		},
	}

	tests := map[string]struct {
		varsFn   func(*testing.T) *vars
		assertFn func(*testing.T, []byte)
	}{
		"happy path": {
			varsFn: func(t *testing.T) *vars {
				vs, err := New(v, WithFileVars(fileVars))
				assert.NoError(t, err)
				return vs
			},
			assertFn: func(t *testing.T, res []byte) {
				expectedByts, err := json.Marshal(fileVars)
				assert.NoError(t, err)
				assert.Equal(t, expectedByts, res)
			},
		},
		"no file vars found": {
			varsFn: func(t *testing.T) *vars {
				vs, err := New(v)
				assert.NoError(t, err)
				return vs
			},
			assertFn: func(t *testing.T, res []byte) {
				assert.Nil(t, res)
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			ctx, cancel := context.WithCancel(ctx)
			defer cancel()

			vs := test.varsFn(t)
			res, err := vs.GetFile(ctx)
			assert.NoError(t, err)
			test.assertFn(t, res)
		})
	}
}

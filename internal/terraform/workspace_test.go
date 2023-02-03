package terraform

import (
	"fmt"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewWorkspace(t *testing.T) {
	t.Parallel()
	v := validator.New()

	tests := map[string]struct {
		v           *validator.Validate
		id          string
		backend     *Object
		sandbox     *Object
		vars        map[string]interface{}
		expected    *workspace
		errExpected error
	}{
		"valid": {
			v:       v,
			id:      "valid",
			backend: &Object{BucketName: "valid", Key: "valid", Region: "us-east-2"},
			sandbox: &Object{BucketName: "sandbox", Key: "sandbox", Region: "us-east-1"},
			vars:    map[string]interface{}{"test": "vars"},
		},
		"empty vars is fine": {
			v:       v,
			id:      "valid",
			backend: &Object{BucketName: "valid", Key: "valid", Region: "us-east-2"},
			sandbox: &Object{BucketName: "sandbox", Key: "sandbox", Region: "us-east-1"},
			vars:    map[string]interface{}{},
		},
		"missing id": {
			v:           v,
			id:          "",
			backend:     &Object{BucketName: "valid", Key: "valid", Region: "us-east-2"},
			sandbox:     &Object{BucketName: "sandbox", Key: "sandbox", Region: "us-east-1"},
			vars:        map[string]interface{}{"test": "vars"},
			errExpected: fmt.Errorf("Field validation for 'ID' failed on the 'required' tag"),
		},
		"missing validator": {
			v:           nil,
			id:          "valid",
			backend:     &Object{BucketName: "valid", Key: "valid", Region: "us-east-2"},
			sandbox:     &Object{BucketName: "sandbox", Key: "sandbox", Region: "us-east-1"},
			vars:        map[string]interface{}{"test": "vars"},
			errExpected: fmt.Errorf("validator is nil"),
		},
		"empty backend bucket": {
			v:           v,
			id:          "valid",
			backend:     &Object{},
			sandbox:     &Object{BucketName: "sandbox", Key: "sandbox", Region: "us-east-1"},
			vars:        map[string]interface{}{"test": "vars"},
			errExpected: fmt.Errorf("Key: 'workspace.Backend.BucketName' Error"),
		},
		"missing backend bucket": {
			v:           v,
			id:          "valid",
			backend:     nil,
			sandbox:     &Object{BucketName: "sandbox", Key: "sandbox", Region: "us-east-1"},
			vars:        map[string]interface{}{"test": "vars"},
			errExpected: fmt.Errorf("Field validation for 'Backend' failed on the 'required' tag"),
		},
		"empty sandbox bucket": {
			v:           v,
			id:          "valid",
			backend:     &Object{BucketName: "sandbox", Key: "sandbox", Region: "us-east-1"},
			sandbox:     &Object{},
			vars:        map[string]interface{}{"test": "vars"},
			errExpected: fmt.Errorf("Key: 'workspace.Sandbox.BucketName' Error"),
		},
		"missing sandbox bucket": {
			v:           v,
			id:          "valid",
			backend:     &Object{BucketName: "sandbox", Key: "sandbox", Region: "us-east-1"},
			sandbox:     nil,
			vars:        map[string]interface{}{"test": "vars"},
			errExpected: fmt.Errorf("Field validation for 'Sandbox' failed on the 'required' tag"),
		},
		"missing vars": {
			v:           v,
			id:          "valid",
			backend:     &Object{BucketName: "valid", Key: "valid", Region: "us-east-2"},
			sandbox:     &Object{BucketName: "sandbox", Key: "sandbox", Region: "us-east-1"},
			vars:        nil,
			errExpected: fmt.Errorf("Field validation for 'Vars' failed on the 'required' tag"),
		},
	}

	for name, test := range tests {
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			w, err := NewWorkspace(
				test.v,
				WithID(test.id),
				WithSandboxBucket(test.sandbox),
				WithBackendBucket(test.backend),
				WithVars(test.vars),
			)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, w)
			assert.Equal(t, test.id, w.ID)
			assert.Equal(t, test.sandbox, w.Sandbox)
			assert.Equal(t, test.backend, w.Backend)
			assert.Equal(t, test.vars, w.Vars)
			assert.NoError(t, w.Cleanup())
		})
	}
}

// TODO(jdt): inline all mock.mock
type mockCleanup struct{ mock.Mock }

func (m *mockCleanup) cleanup() error {
	args := m.Called()
	return args.Error(0)
}

func TestWorkspace_Cleanup(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		fns         func(*testing.T) []func() error
		errExpected error
	}{
		"no fns registered": {fns: func(t *testing.T) []func() error { return nil }},
		"single fn w/o error": {
			fns: func(t *testing.T) []func() error {
				m := &mockCleanup{}
				m.On("cleanup").Return(nil)
				return []func() error{m.cleanup}
			},
		},
		"single fn w error": {
			fns: func(t *testing.T) []func() error {
				m := &mockCleanup{}
				m.On("cleanup").Return(fmt.Errorf("oops"))
				return []func() error{m.cleanup}
			},
			errExpected: fmt.Errorf("oops"),
		},
		"multiple fns w/o error": {
			fns: func(t *testing.T) []func() error {
				m := &mockCleanup{}
				m.On("cleanup").Return(nil).Times(3)
				return []func() error{m.cleanup, m.cleanup, m.cleanup}
			},
		},
		"multiple fns w error": {
			fns: func(t *testing.T) []func() error {
				m := &mockCleanup{}
				m.On("cleanup").Return(fmt.Errorf("oops")).Times(3)
				return []func() error{m.cleanup, m.cleanup, m.cleanup}
			},
			errExpected: fmt.Errorf("3 errors occurred"),
		},
	}

	for name, test := range tests {
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			w := &workspace{cleanupFns: test.fns(t)}
			err := w.Cleanup()
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
		})
	}
}

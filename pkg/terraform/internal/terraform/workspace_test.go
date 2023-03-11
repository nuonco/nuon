package terraform

import (
	"fmt"
	"testing"

	"github.com/go-playground/validator/v10"
	planv1 "github.com/powertoolsdev/mono/pkg/protos/workflows/generated/types/executors/v1/plan/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewWorkspace(t *testing.T) {
	t.Parallel()
	v := validator.New()

	tests := map[string]struct {
		v           *validator.Validate
		id          string
		backend     *planv1.Object
		sandbox     *planv1.Object
		vars        map[string]interface{}
		version     string
		expected    *workspace
		errExpected error
	}{
		"valid": {
			v:       v,
			id:      "valid",
			backend: &planv1.Object{Bucket: "valid", Key: "valid", Region: "us-east-2"},
			sandbox: &planv1.Object{Bucket: "sandbox", Key: "sandbox", Region: "us-east-1"},
			vars:    map[string]interface{}{"test": "vars"},
			version: "v1.3.9",
		},
		"empty vars is fine": {
			v:       v,
			id:      "valid",
			backend: &planv1.Object{Bucket: "valid", Key: "valid", Region: "us-east-2"},
			sandbox: &planv1.Object{Bucket: "sandbox", Key: "sandbox", Region: "us-east-1"},
			vars:    map[string]interface{}{},
			version: "v1.3.9",
		},
		"empty version": {
			v:           v,
			id:          "valid",
			backend:     &planv1.Object{Bucket: "valid", Key: "valid", Region: "us-east-2"},
			sandbox:     &planv1.Object{Bucket: "sandbox", Key: "sandbox", Region: "us-east-1"},
			vars:        map[string]interface{}{"test": "vars"},
			errExpected: fmt.Errorf("Field validation for 'Version' failed on the 'required' tag"),
		},
		"missing id": {
			v:           v,
			id:          "",
			backend:     &planv1.Object{Bucket: "valid", Key: "valid", Region: "us-east-2"},
			sandbox:     &planv1.Object{Bucket: "sandbox", Key: "sandbox", Region: "us-east-1"},
			vars:        map[string]interface{}{"test": "vars"},
			version:     "v1.3.9",
			errExpected: fmt.Errorf("Field validation for 'ID' failed on the 'required' tag"),
		},
		"missing validator": {
			v:           nil,
			id:          "valid",
			backend:     &planv1.Object{Bucket: "valid", Key: "valid", Region: "us-east-2"},
			sandbox:     &planv1.Object{Bucket: "sandbox", Key: "sandbox", Region: "us-east-1"},
			vars:        map[string]interface{}{"test": "vars"},
			version:     "v1.3.9",
			errExpected: fmt.Errorf("validator is nil"),
		},
		// "empty backend bucket": {
		// 	v:           v,
		// 	id:          "valid",
		// 	backend:     &planv1.Object{},
		// 	sandbox:     &planv1.Object{Bucket: "sandbox", Key: "sandbox", Region: "us-east-1"},
		// 	vars:        map[string]interface{}{"test": "vars"},
		// version: "v1.3.9",
		// 	errExpected: fmt.Errorf("Key: 'workspace.Backend.BucketName' Error"),
		// },
		"missing backend bucket": {
			v:           v,
			id:          "valid",
			backend:     nil,
			sandbox:     &planv1.Object{Bucket: "sandbox", Key: "sandbox", Region: "us-east-1"},
			vars:        map[string]interface{}{"test": "vars"},
			version:     "v1.3.9",
			errExpected: fmt.Errorf("Field validation for 'Backend' failed on the 'required' tag"),
		},
		// "empty sandbox bucket": {
		// 	v:           v,
		// 	id:          "valid",
		// 	backend:     &planv1.Object{Bucket: "sandbox", Key: "sandbox", Region: "us-east-1"},
		// 	sandbox:     &planv1.Object{},
		// 	vars:        map[string]interface{}{"test": "vars"},
		// version: "v1.3.9",
		// 	errExpected: fmt.Errorf("Key: 'workspace.Module.Bucket' Error"),
		// },
		"missing sandbox bucket": {
			v:           v,
			id:          "valid",
			backend:     &planv1.Object{Bucket: "sandbox", Key: "sandbox", Region: "us-east-1"},
			sandbox:     nil,
			vars:        map[string]interface{}{"test": "vars"},
			version:     "v1.3.9",
			errExpected: fmt.Errorf("Field validation for 'Module' failed on the 'required' tag"),
		},
		"missing vars": {
			v:           v,
			id:          "valid",
			backend:     &planv1.Object{Bucket: "valid", Key: "valid", Region: "us-east-2"},
			sandbox:     &planv1.Object{Bucket: "sandbox", Key: "sandbox", Region: "us-east-1"},
			vars:        nil,
			version:     "v1.3.9",
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
				WithModuleBucket(test.sandbox),
				WithBackendBucket(test.backend),
				WithVars(test.vars),
				WithVersion(test.version),
			)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, w)
			assert.Equal(t, test.id, w.ID)
			assert.Equal(t, test.sandbox, w.Module)
			assert.Equal(t, test.backend, w.Backend)
			assert.Equal(t, test.vars, w.Vars)
			assert.Equal(t, test.version, w.Version)
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

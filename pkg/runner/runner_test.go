package runner

import (
	"fmt"
	"testing"

	"github.com/go-playground/validator/v10"
	planv1 "github.com/powertoolsdev/protos/workflows/generated/types/executors/v1/plan/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNew(t *testing.T) {
	t.Parallel()
	v := validator.New()
	tests := map[string]struct {
		v           *validator.Validate
		opts        func(*testing.T) []runnerOption
		expected    *runner
		errExpected error
	}{
		"happy path": {
			v: v,
			opts: func(t *testing.T) []runnerOption {
				return []runnerOption{
					WithPlan(&planv1.TerraformPlan{}),
				}
			},
			expected: &runner{Plan: &planv1.TerraformPlan{}, validator: v},
		},
		"missing validator": {
			v: nil,
			opts: func(t *testing.T) []runnerOption {
				return []runnerOption{
					WithPlan(&planv1.TerraformPlan{}),
				}
			},
			expected:    &runner{Plan: &planv1.TerraformPlan{}},
			errExpected: fmt.Errorf("validator is nil"),
		},
		"no plan given": {
			v: v,
			opts: func(t *testing.T) []runnerOption {
				return []runnerOption{}
			},
			errExpected: fmt.Errorf("Field validation for 'Plan' failed on the 'required' tag"),
		},
	}

	for name, test := range tests {
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			opts := test.opts(t)
			got, err := New(test.v, opts...)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)

			got.cleanupFns = nil
			got.workspaceSetuper = nil
			assert.Equal(t, test.expected, got)
		})
	}
}

type mockCleanup struct{ mock.Mock }

func (m *mockCleanup) cleanup() error {
	args := m.Called()
	return args.Error(0)
}

func TestRunner_cleanup(t *testing.T) {
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
			r := &runner{cleanupFns: test.fns(t)}
			err := r.cleanup()
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
		})
	}
}

package terraform

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockTerraformClient struct {
	mock.Mock
}

var _ tfExecutor = (*mockTerraformClient)(nil)

func (m *mockTerraformClient) Init(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestWorkspace_Init(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		tfExecutor  func() tfExecutor
		errExpected error
	}{
		"valid": {
			tfExecutor: func() tfExecutor {
				m := &mockTerraformClient{}
				m.On("Init", mock.Anything).Return(nil)
				return m
			},
		},
		"error": {
			tfExecutor: func() tfExecutor {
				m := &mockTerraformClient{}
				m.On("Init", mock.Anything).Return(fmt.Errorf("oops"))
				return m
			},
			errExpected: fmt.Errorf("oops"),
		},
	}

	for name, test := range tests {
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			w := &workspace{tfExecutor: test.tfExecutor()}
			err := w.Init(context.Background())
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
		})
	}
}

func (m *mockTerraformClient) Plan(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestWorkspace_Plan(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		tfExecutor  func() tfExecutor
		errExpected error
	}{
		"valid": {
			tfExecutor: func() tfExecutor {
				m := &mockTerraformClient{}
				m.On("Plan", mock.Anything).Return(nil)
				return m
			},
		},
		"error": {
			tfExecutor: func() tfExecutor {
				m := &mockTerraformClient{}
				m.On("Plan", mock.Anything).Return(fmt.Errorf("oops"))
				return m
			},
			errExpected: fmt.Errorf("oops"),
		},
	}

	for name, test := range tests {
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			w := &workspace{tfExecutor: test.tfExecutor()}
			err := w.Plan(context.Background())
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
		})
	}
}

func (m *mockTerraformClient) Apply(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestWorkspace_Apply(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		tfExecutor  func() tfExecutor
		errExpected error
	}{
		"valid": {
			tfExecutor: func() tfExecutor {
				m := &mockTerraformClient{}
				m.On("Apply", mock.Anything).Return(nil)
				return m
			},
		},
		"error": {
			tfExecutor: func() tfExecutor {
				m := &mockTerraformClient{}
				m.On("Apply", mock.Anything).Return(fmt.Errorf("oops"))
				return m
			},
			errExpected: fmt.Errorf("oops"),
		},
	}

	for name, test := range tests {
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			w := &workspace{tfExecutor: test.tfExecutor()}
			err := w.Apply(context.Background())
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
		})
	}
}

func (m *mockTerraformClient) Destroy(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestWorkspace_Destroy(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		tfExecutor  func() tfExecutor
		errExpected error
	}{
		"valid": {
			tfExecutor: func() tfExecutor {
				m := &mockTerraformClient{}
				m.On("Destroy", mock.Anything).Return(nil)
				return m
			},
		},
		"error": {
			tfExecutor: func() tfExecutor {
				m := &mockTerraformClient{}
				m.On("Destroy", mock.Anything).Return(fmt.Errorf("oops"))
				return m
			},
			errExpected: fmt.Errorf("oops"),
		},
	}

	for name, test := range tests {
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			w := &workspace{tfExecutor: test.tfExecutor()}
			err := w.Destroy(context.Background())
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
		})
	}
}

func (m *mockTerraformClient) Output(ctx context.Context) (map[string]interface{}, error) {
	args := m.Called(ctx)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func TestWorkspace_Output(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		tfExecutor  func() tfExecutor
		expected    map[string]interface{}
		errExpected error
	}{
		"valid": {
			tfExecutor: func() tfExecutor {
				m := &mockTerraformClient{}
				m.On("Output", mock.Anything).Return(map[string]interface{}{}, nil)
				return m
			},
			expected: map[string]interface{}{},
		},
		"outputs returned unchanged": {
			tfExecutor: func() tfExecutor {
				m := &mockTerraformClient{}
				m.On("Output", mock.Anything).Return(map[string]interface{}{"output": "unchanged"}, nil)
				return m
			},
			expected: map[string]interface{}{"output": "unchanged"},
		},

		"error": {
			tfExecutor: func() tfExecutor {
				m := &mockTerraformClient{}
				m.On("Output", mock.Anything).Return(map[string]interface{}{}, fmt.Errorf("oops"))
				return m
			},
			errExpected: fmt.Errorf("oops"),
		},
	}

	for name, test := range tests {
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			w := &workspace{tfExecutor: test.tfExecutor()}
			m, err := w.Output(context.Background())
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, test.expected, m)
		})
	}
}

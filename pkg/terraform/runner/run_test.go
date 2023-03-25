package runner

import (
	"context"
	"fmt"
	"testing"

	planv1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/plan/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockWorkspaceSetuper struct{ mock.Mock }

var _ workspaceSetuper = (*mockWorkspaceSetuper)(nil)

func (m *mockWorkspaceSetuper) setupWorkspace(ctx context.Context, req *planv1.TerraformPlan) (executor, error) {
	args := m.Called(ctx, req)
	err := args.Error(1)
	if args.Get(0) != nil {
		return args.Get(0).(executor), err
	}
	return nil, err
}

func TestRunner_Run(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	tests := map[string]struct {
		ws          func(*testing.T) *mockWorkspaceSetuper
		expected    map[string]interface{}
		errExpected error
	}{
		"happiest path": {
			ws: func(t *testing.T) *mockWorkspaceSetuper {
				me := &mockExecutor{}
				me.On("Init", ctx).Return(nil)
				me.On("Apply", ctx).Return(nil)
				me.On("Output", ctx).Return(map[string]interface{}{"got": "outputs"}, nil)
				m := &mockWorkspaceSetuper{}
				m.On("setupWorkspace", ctx, &planv1.TerraformPlan{RunType: planv1.TerraformRunType_TERRAFORM_RUN_TYPE_APPLY}).Return(me, nil)
				return m
			},
			expected: map[string]interface{}{"got": "outputs"},
		},
		"failed setting up workspace": {
			ws: func(t *testing.T) *mockWorkspaceSetuper {
				m := &mockWorkspaceSetuper{}
				m.On("setupWorkspace",
					ctx,
					&planv1.TerraformPlan{RunType: planv1.TerraformRunType_TERRAFORM_RUN_TYPE_APPLY},
				).Return(nil, fmt.Errorf("failed setting up workspace"))
				return m
			},
			errExpected: fmt.Errorf("failed setting up workspace"),
		},
	}

	for name, test := range tests {
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ws := test.ws(t)
			r := &runner{
				Plan:             &planv1.TerraformPlan{RunType: planv1.TerraformRunType_TERRAFORM_RUN_TYPE_APPLY},
				workspaceSetuper: ws,
			}

			got, err := r.Run(ctx)
			ws.AssertExpectations(t)

			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, test.expected, got)
		})
	}
}

type mockExecutor struct{ mock.Mock }

func (m *mockExecutor) Init(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}
func (m *mockExecutor) Plan(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}
func (m *mockExecutor) Apply(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}
func (m *mockExecutor) Destroy(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}
func (m *mockExecutor) Output(ctx context.Context) (map[string]interface{}, error) {
	args := m.Called(ctx)
	err := args.Error(1)
	if m, ok := args.Get(0).(map[string]interface{}); ok {
		return m, err
	}
	return nil, err
}

func Test_run(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		executor    func(*testing.T) *mockExecutor
		runType     planv1.TerraformRunType
		expected    map[string]interface{}
		errExpected error
	}{
		"happy path - plan": {
			executor: func(t *testing.T) *mockExecutor {
				m := &mockExecutor{}
				m.On("Init", mock.Anything).Return(nil)
				m.On("Plan", mock.Anything).Return(nil)
				return m
			},
			runType: planv1.TerraformRunType_TERRAFORM_RUN_TYPE_PLAN,
		},
		"happy path - apply": {
			executor: func(t *testing.T) *mockExecutor {
				m := &mockExecutor{}
				m.On("Init", mock.Anything).Return(nil)
				m.On("Apply", mock.Anything).Return(nil)
				m.On("Output", mock.Anything).Return(map[string]interface{}{}, nil)
				return m
			},
			runType:  planv1.TerraformRunType_TERRAFORM_RUN_TYPE_APPLY,
			expected: map[string]interface{}{},
		},
		"happy path - destroy": {
			executor: func(t *testing.T) *mockExecutor {
				m := &mockExecutor{}
				m.On("Init", mock.Anything).Return(nil)
				m.On("Destroy", mock.Anything).Return(nil)
				return m
			},
			runType: planv1.TerraformRunType_TERRAFORM_RUN_TYPE_DESTROY,
		},
		"returns outputs": {
			executor: func(t *testing.T) *mockExecutor {
				m := &mockExecutor{}
				m.On("Init", mock.Anything).Return(nil)
				m.On("Apply", mock.Anything).Return(nil)
				m.On("Output", mock.Anything).Return(map[string]interface{}{
					"a":    "bunch",
					"of":   "really",
					"cool": "outputs",
				}, nil)
				return m
			},
			runType:  planv1.TerraformRunType_TERRAFORM_RUN_TYPE_APPLY,
			expected: map[string]interface{}{"a": "bunch", "of": "really", "cool": "outputs"},
		},
		"invalid runtype": {
			executor: func(t *testing.T) *mockExecutor {
				m := &mockExecutor{}
				m.On("Init", mock.Anything).Return(nil)
				return m
			},
			runType:     planv1.TerraformRunType(555),
			errExpected: fmt.Errorf("invalid run type"),
		},
		"init error": {
			executor: func(t *testing.T) *mockExecutor {
				m := &mockExecutor{}
				m.On("Init", mock.Anything).Return(fmt.Errorf("init error"))
				return m
			},
			runType:     planv1.TerraformRunType_TERRAFORM_RUN_TYPE_DESTROY,
			errExpected: fmt.Errorf("init error"),
		},
		"plan error": {
			executor: func(t *testing.T) *mockExecutor {
				m := &mockExecutor{}
				m.On("Init", mock.Anything).Return(nil)
				m.On("Plan", mock.Anything).Return(fmt.Errorf("plan error"))
				return m
			},
			runType:     planv1.TerraformRunType_TERRAFORM_RUN_TYPE_PLAN,
			errExpected: fmt.Errorf("plan error"),
		},
		"destroy error": {
			executor: func(t *testing.T) *mockExecutor {
				m := &mockExecutor{}
				m.On("Init", mock.Anything).Return(nil)
				m.On("Destroy", mock.Anything).Return(fmt.Errorf("destroy error"))
				return m
			},
			runType:     planv1.TerraformRunType_TERRAFORM_RUN_TYPE_DESTROY,
			errExpected: fmt.Errorf("destroy error"),
		},
		"apply error": {
			executor: func(t *testing.T) *mockExecutor {
				m := &mockExecutor{}
				m.On("Init", mock.Anything).Return(nil)
				m.On("Apply", mock.Anything).Return(fmt.Errorf("apply error"))
				return m
			},
			runType:     planv1.TerraformRunType_TERRAFORM_RUN_TYPE_APPLY,
			errExpected: fmt.Errorf("apply error"),
		},
		"output error": {
			executor: func(t *testing.T) *mockExecutor {
				m := &mockExecutor{}
				m.On("Init", mock.Anything).Return(nil)
				m.On("Apply", mock.Anything).Return(nil)
				m.On("Output", mock.Anything).Return(nil, fmt.Errorf("output error"))
				return m
			},
			runType:     planv1.TerraformRunType_TERRAFORM_RUN_TYPE_APPLY,
			errExpected: fmt.Errorf("output error"),
		},
	}

	for name, test := range tests {
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			e := test.executor(t)

			got, err := run(context.Background(), e, test.runType)
			assert.Equal(t, test.expected, got)

			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
			e.AssertExpectations(t)
		})
	}
}

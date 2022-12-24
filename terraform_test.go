package terraform

import (
	"context"
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockTerraformExecutor struct {
	mock.Mock
}

func (t *mockTerraformExecutor) initClient(execPath, tmpDir string) error {
	args := t.Called(execPath, tmpDir)
	return args.Error(0)
}

func (t *mockTerraformExecutor) setLogger(printfer) {
	t.Called()
}

func (t *mockTerraformExecutor) setStderr(io.Writer) {
	t.Called()
}

func (t *mockTerraformExecutor) setStdout(io.Writer) {
	t.Called()
}

func (t *mockTerraformExecutor) setEnvVars(vars map[string]string) error {
	args := t.Called(vars)
	return args.Error(0)
}

func (t *mockTerraformExecutor) initModule(context.Context) error {
	args := t.Called()
	return args.Error(0)
}

func (t *mockTerraformExecutor) planModule(context.Context) error {
	args := t.Called()
	return args.Error(0)
}

func (t *mockTerraformExecutor) applyModule(context.Context) error {
	args := t.Called()
	return args.Error(0)
}

func (t *mockTerraformExecutor) destroyModule(context.Context) error {
	args := t.Called()
	return args.Error(0)
}

func (t *mockTerraformExecutor) outputs(context.Context) (map[string]interface{}, error) {
	args := t.Called()
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

var _ terraformExecutor = (*mockTerraformExecutor)(nil)

type mockTerraformClient struct {
	mock.Mock
}

func (m *mockTerraformClient) Init(ctx context.Context, opts ...tfexec.InitOption) error {
	args := m.Called(opts)
	return args.Error(0)
}

func (m *mockTerraformClient) Apply(ctx context.Context, opts ...tfexec.ApplyOption) error {
	args := m.Called(opts)
	return args.Error(0)
}

func (m *mockTerraformClient) Destroy(ctx context.Context, opts ...tfexec.DestroyOption) error {
	args := m.Called(opts)
	return args.Error(0)
}

func (m *mockTerraformClient) Plan(ctx context.Context, opts ...tfexec.PlanOption) (bool, error) {
	args := m.Called(opts)
	return args.Bool(0), args.Error(1)
}

func (m *mockTerraformClient) SetStderr(f io.Writer) {
	m.Called(f)
}

func (m *mockTerraformClient) SetStdout(f io.Writer) {
	m.Called(f)
}

func (m *mockTerraformClient) SetEnv(e map[string]string) error {
	args := m.Called(e)
	return args.Error(0)
}

var _ terraformClient = (*mockTerraformClient)(nil)

type mockOutputter struct {
	mock.Mock
}

func (m *mockOutputter) Output(ctx context.Context, opts ...tfexec.OutputOption) (map[string]tfexec.OutputMeta, error) {
	args := m.Called(ctx, opts[:])
	return args.Get(0).(map[string]tfexec.OutputMeta), args.Error(1)
}

var _ outputter = (*mockOutputter)(nil)

func Test_init(t *testing.T) {
	testErr := errors.New("test")

	mockClient := new(mockTerraformClient)
	mockClient.On("Init", mock.Anything, mock.Anything).Return(testErr)

	tfExecutor := &localTerraformExecutor{tfClient: mockClient}
	err := tfExecutor.initModule(context.Background())
	assert.Equal(t, err, testErr)
}

func Test_apply(t *testing.T) {
	testErr := errors.New("test")

	mockClient := new(mockTerraformClient)
	mockClient.On("Apply", mock.Anything).Return(testErr)

	tfExecutor := &localTerraformExecutor{tfClient: mockClient}
	err := tfExecutor.applyModule(context.Background())
	assert.Equal(t, err, testErr)
}

func Test_destroy(t *testing.T) {
	testErr := errors.New("test")

	mockClient := new(mockTerraformClient)
	mockClient.On("Destroy", mock.Anything).Return(testErr)

	tfExecutor := &localTerraformExecutor{tfClient: mockClient}
	err := tfExecutor.destroyModule(context.Background())
	assert.Equal(t, err, testErr)
}

func Test_setStderr(t *testing.T) {
	mockClient := new(mockTerraformClient)
	mockClient.On("SetStderr", mock.Anything).Return()

	tfExecutor := &localTerraformExecutor{tfClient: mockClient}
	tfExecutor.setStderr(io.Discard)
	mockClient.AssertNumberOfCalls(t, "SetStderr", 1)
}

func Test_setStdout(t *testing.T) {
	mockClient := new(mockTerraformClient)
	mockClient.On("SetStdout", mock.Anything).Return()

	tfExecutor := &localTerraformExecutor{tfClient: mockClient}
	tfExecutor.setStdout(io.Discard)
	mockClient.AssertNumberOfCalls(t, "SetStdout", 1)
}

func Test_terraformSetEnvVars(t *testing.T) {
	mockClient := new(mockTerraformClient)
	envVars := map[string]string{
		"FOO": "BAR",
	}

	expectedEnv := getEnv()
	expectedEnv["FOO"] = "BAR"

	mockClient.On("SetEnv", expectedEnv).Return(nil)

	tfExecutor := &localTerraformExecutor{tfClient: mockClient}
	err := tfExecutor.setEnvVars(envVars)
	assert.NoError(t, err)
	mockClient.AssertNumberOfCalls(t, "SetEnv", 1)
}

func Test_plan(t *testing.T) {
	testErr := errors.New("test")

	mockClient := new(mockTerraformClient)
	mockClient.On("Plan", mock.Anything).Return(false, testErr)

	tfExecutor := &localTerraformExecutor{tfClient: mockClient}
	err := tfExecutor.planModule(context.Background())
	assert.Equal(t, err, testErr)
}

const outputObjectType string = `
      "object",
      {
	"number": "number",
	"string": "string"
      }
`

const outputObjectValue string = `
{
  "number": 1,
  "string": "a"
}
`

func Test_outputs(t *testing.T) {
	tests := map[string]struct {
		fn          func(t *testing.T) outputter
		expected    map[string]interface{}
		errExpected error
	}{
		"happy path - no outputs": {
			fn: func(t *testing.T) outputter {
				mo := &mockOutputter{}
				mo.On("Output", mock.Anything, []tfexec.OutputOption(nil)).
					Return(map[string]tfexec.OutputMeta{}, nil).Once()

				return mo
			},
			expected: map[string]interface{}{},
		},

		"happy path - with outputs": {
			fn: func(t *testing.T) outputter {
				mo := &mockOutputter{}
				mo.On("Output", mock.Anything, []tfexec.OutputOption(nil)).
					Return(map[string]tfexec.OutputMeta{
						"myoutput": {
							Sensitive: false,
							Type:      []byte(`"string"`),
							Value:     []byte(`"string value"`),
						},
					}, nil).Once()

				return mo
			},

			expected: map[string]interface{}{"myoutput": "string value"},
		},

		"happy path - map outputs": {
			fn: func(t *testing.T) outputter {
				mo := &mockOutputter{}
				mo.On("Output", mock.Anything, []tfexec.OutputOption(nil)).
					Return(map[string]tfexec.OutputMeta{
						"myoutput": {
							Sensitive: false,
							Type:      []byte(outputObjectType),
							Value:     []byte(outputObjectValue),
						},
					}, nil).Once()

				return mo
			},

			expected: map[string]interface{}{"myoutput": map[string]interface{}{"number": float64(1), "string": "a"}},
		},

		"errors on invalid json output": {
			fn: func(t *testing.T) outputter {
				mo := &mockOutputter{}
				mo.On("Output", mock.Anything, []tfexec.OutputOption(nil)).
					Return(map[string]tfexec.OutputMeta{
						"invalid": {
							Sensitive: false,
							Type:      []byte(`"string"`),
							Value:     []byte(`unquoted / invalid string value`),
						},
					}, nil).Once()

				return mo
			},
			errExpected: fmt.Errorf("invalid character"),
		},

		"errors on outputter error": {
			fn: func(t *testing.T) outputter {
				mo := &mockOutputter{}
				mo.On("Output", mock.Anything, []tfexec.OutputOption(nil)).
					Return(map[string]tfexec.OutputMeta{}, errors.New("oops")).Once()

				return mo
			},
			errExpected: errors.New("oops"),
		},

		"errors without outputter": {
			fn: func(t *testing.T) outputter {
				return nil
			},
			errExpected: fmt.Errorf("missing outputter"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			o := test.fn(t)
			tfExecutor := &localTerraformExecutor{outputter: o}
			m, err := tfExecutor.outputs(context.Background())
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, test.expected, m)

			if o, ok := o.(*mockOutputter); ok {
				o.AssertExpectations(t)
			}
		})
	}
}

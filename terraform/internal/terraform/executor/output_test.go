package executor

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func (m *mockTerraformClient) Output(ctx context.Context, opts ...tfexec.OutputOption) (map[string]tfexec.OutputMeta, error) {
	args := m.Called(ctx, opts[:])
	return args.Get(0).(map[string]tfexec.OutputMeta), args.Error(1)
}

func TestTfExecutor_Output(t *testing.T) {
	tests := map[string]struct {
		setupFn     func(t *testing.T) outputter
		expected    map[string]interface{}
		errExpected error
	}{
		"happy path - no outputs": {
			setupFn: func(t *testing.T) outputter {
				m := &mockTerraformClient{}
				m.On("Output", mock.Anything, []tfexec.OutputOption(nil)).
					Return(map[string]tfexec.OutputMeta{}, nil).Once()

				return m
			},
			expected: map[string]interface{}{},
		},

		"happy path - with outputs": {
			setupFn: func(t *testing.T) outputter {
				m := &mockTerraformClient{}
				m.On("Output", mock.Anything, []tfexec.OutputOption(nil)).
					Return(map[string]tfexec.OutputMeta{
						"mystring": {
							Sensitive: false,
							Type:      []byte(`"string"`),
							Value:     []byte(`"string value"`),
						},
						"mynumber": {
							Sensitive: false,
							Type:      []byte(`"number"`),
							Value:     []byte(`5`),
						},
					}, nil).Once()

				return m
			},

			expected: map[string]interface{}{"mynumber": float64(5), "mystring": "string value"},
		},
		"nested values are skipped": {
			setupFn: func(t *testing.T) outputter {
				m := &mockTerraformClient{}
				m.
					On("Output", mock.Anything, []tfexec.OutputOption(nil)).
					Return(map[string]tfexec.OutputMeta{
						"nested": {
							Sensitive: false,
							Type:      []byte(`"object", {"number": "number","string": "string"}`),
							Value:     []byte(`{"number": 1, "string": "a" }`),
						},
						"mystring": {
							Sensitive: false,
							Type:      []byte(`"string"`),
							Value:     []byte(`"string value"`),
						},
					}, nil).Once()

				return m
			},
			expected: map[string]interface{}{
				"mystring": "string value",
				"nested":   map[string]interface{}{"number": float64(1), "string": "a"},
			},
		},

		"errors on invalid json output": {
			setupFn: func(t *testing.T) outputter {
				m := &mockTerraformClient{}
				m.On("Output", mock.Anything, []tfexec.OutputOption(nil)).
					Return(map[string]tfexec.OutputMeta{
						"invalid": {
							Sensitive: false,
							Type:      []byte(`"string"`),
							Value:     []byte(`unquoted / invalid string value`),
						},
					}, nil).Once()

				return m
			},
			errExpected: fmt.Errorf("invalid character"),
		},

		"errors on outputter error": {
			setupFn: func(t *testing.T) outputter {
				m := &mockTerraformClient{}
				m.On("Output", mock.Anything, []tfexec.OutputOption(nil)).
					Return(map[string]tfexec.OutputMeta{}, errors.New("oops")).Once()

				return m
			},
			errExpected: errors.New("oops"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			o := test.setupFn(t)
			tfExecutor := &tfExecutor{outputter: o}
			m, err := tfExecutor.Output(context.Background())
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, test.expected, m)

			o.(*mockTerraformClient).AssertExpectations(t)
		})
	}
}

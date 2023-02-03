package executor

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func (m *mockTerraformClient) Plan(ctx context.Context, opts ...tfexec.PlanOption) (bool, error) {
	args := m.Called(ctx, opts)
	return args.Bool(0), args.Error(1)
}

func TestTfExecutor_Plan(t *testing.T) {
	tests := map[string]struct {
		setupFn     func(t *testing.T) planner
		errExpected error
	}{
		"happy path": {
			setupFn: func(t *testing.T) planner {
				m := &mockTerraformClient{}
				m.
					On("Plan", mock.Anything, []tfexec.PlanOption{tfexec.Refresh(true), tfexec.VarFile(t.Name())}).
					Return(true, nil).
					Once()

				return m
			},
		},
		"errors on error": {
			setupFn: func(t *testing.T) planner {
				m := &mockTerraformClient{}
				m.
					On("Plan", mock.Anything, []tfexec.PlanOption{tfexec.Refresh(true), tfexec.VarFile(t.Name())}).
					Return(true, fmt.Errorf("oops")).
					Once()

				return m
			},
			errExpected: fmt.Errorf("oops"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			m := test.setupFn(t)
			tfExecutor := &tfExecutor{planner: m, VarFile: t.Name()}
			err := tfExecutor.Plan(context.Background())
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)

			m.(*mockTerraformClient).AssertExpectations(t)
		})
	}
}

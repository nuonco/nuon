package executor

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func (m *mockTerraformClient) Apply(ctx context.Context, opts ...tfexec.ApplyOption) error {
	args := m.Called(ctx, opts)
	return args.Error(0)
}

func TestTfExecutor_Apply(t *testing.T) {
	tests := map[string]struct {
		setupFn     func(t *testing.T) applier
		errExpected error
	}{
		"happy path": {
			setupFn: func(t *testing.T) applier {
				m := &mockTerraformClient{}
				m.
					On("Apply", mock.Anything, []tfexec.ApplyOption{tfexec.Refresh(true), tfexec.VarFile(t.Name())}).
					Return(nil).
					Once()

				return m
			},
		},
		"errors on error": {
			setupFn: func(t *testing.T) applier {
				m := &mockTerraformClient{}
				m.
					On("Apply", mock.Anything, []tfexec.ApplyOption{tfexec.Refresh(true), tfexec.VarFile(t.Name())}).
					Return(fmt.Errorf("oops")).
					Once()

				return m
			},
			errExpected: fmt.Errorf("oops"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			m := test.setupFn(t)
			tfExecutor := &tfExecutor{applier: m, VarFile: t.Name()}
			err := tfExecutor.Apply(context.Background())
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)

			m.(*mockTerraformClient).AssertExpectations(t)
		})
	}
}

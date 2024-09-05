package pipeline

import (
	"context"
	"fmt"
	io "io"
	"testing"

	"github.com/go-playground/validator/v10"
	gomock "github.com/golang/mock/gomock"
	"github.com/hashicorp/go-hclog"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/stretchr/testify/assert"
)

//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=exec_step_mock_test.go -source=exec_step_test.go -package=pipeline
type testStepFunctions interface {
	ValidExecFn(context.Context, hclog.Logger) ([]byte, error)
	ValidCallbackFn(context.Context, hclog.Logger, []byte) error
}

func TestPipeline_execStep(t *testing.T) {
	errPipelineRun := fmt.Errorf("error running pipeline")
	execResp := generics.GetFakeObj[[]byte]()
	stepName := generics.GetFakeObj[string]()

	v := validator.New()
	l := hclog.New(&hclog.LoggerOptions{
		Output: io.Discard,
		Name:   "exp",
		Level:  hclog.LevelFromString("DEBUG"),
	})

	tests := map[string]struct {
		stepFn      func(*gomock.Controller, context.Context) *Step
		errExpected error
		assertFn    func(t *testing.T)
	}{
		"happy path": {
			stepFn: func(mockCtl *gomock.Controller, ctx context.Context) *Step {
				mock := NewMocktestStepFunctions(mockCtl)
				mock.EXPECT().ValidExecFn(ctx, l).Return(execResp, nil)
				mock.EXPECT().ValidCallbackFn(ctx, l, execResp).Return(nil)

				return &Step{
					Name:       stepName,
					ExecFn:     mock.ValidExecFn,
					CallbackFn: mock.ValidCallbackFn,
				}
			},
		},
		"invalid step missing name": {
			stepFn: func(mockCtl *gomock.Controller, _ context.Context) *Step {
				mock := NewMocktestStepFunctions(mockCtl)
				return &Step{
					ExecFn:     mock.ValidExecFn,
					CallbackFn: mock.ValidCallbackFn,
				}
			},
			errExpected: fmt.Errorf("Name"),
		},
		"invalid step missing exec": {
			stepFn: func(mockCtl *gomock.Controller, _ context.Context) *Step {
				mock := NewMocktestStepFunctions(mockCtl)
				return &Step{
					Name:       stepName,
					CallbackFn: mock.ValidCallbackFn,
				}
			},
			errExpected: fmt.Errorf("ExecFn"),
		},
		"invalid step missing callback": {
			stepFn: func(mockCtl *gomock.Controller, _ context.Context) *Step {
				mock := NewMocktestStepFunctions(mockCtl)
				return &Step{
					Name:   stepName,
					ExecFn: mock.ValidExecFn,
				}
			},
			errExpected: fmt.Errorf("CallbackFn"),
		},
		"error on exec fn": {
			stepFn: func(mockCtl *gomock.Controller, ctx context.Context) *Step {
				mock := NewMocktestStepFunctions(mockCtl)
				mock.EXPECT().ValidExecFn(ctx, l).Return(nil, errPipelineRun)

				return &Step{
					Name:       stepName,
					ExecFn:     mock.ValidExecFn,
					CallbackFn: mock.ValidCallbackFn,
				}
			},
			errExpected: errPipelineRun,
		},
		"error on callback fn": {
			stepFn: func(mockCtl *gomock.Controller, ctx context.Context) *Step {
				mock := NewMocktestStepFunctions(mockCtl)
				mock.EXPECT().ValidExecFn(ctx, l).Return(execResp, nil)
				mock.EXPECT().ValidCallbackFn(ctx, l, execResp).Return(errPipelineRun)

				return &Step{
					Name:       stepName,
					ExecFn:     mock.ValidExecFn,
					CallbackFn: mock.ValidCallbackFn,
				}
			},
			errExpected: errPipelineRun,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			mockCtl, ctx := gomock.WithContext(ctx, t)

			pipe, err := New(v,
				WithLogger(l),
			)
			assert.NoError(t, err)

			step := test.stepFn(mockCtl, ctx)
			err = pipe.execStep(ctx, step)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}

			assert.NoError(t, err)
		})
	}
}

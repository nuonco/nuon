package exec

import (
	context "context"
	"encoding/json"
	"fmt"
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/hashicorp/go-hclog"
	tfexec "github.com/hashicorp/terraform-exec/tfexec"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/powertoolsdev/mono/pkg/pipeline"
	"github.com/stretchr/testify/assert"
)

func TestTerraform(t *testing.T) {
	l := NewMockhcLog(nil)
	ui := NewMockui(nil)
	terraformErr := fmt.Errorf("error terraform mapper")

	tests := map[string]struct {
		execFn      func(*gomock.Controller) pipeline.ExecFn
		assertFn    func(*testing.T, []byte)
		errExpected error
	}{
		"output - happy path": {
			execFn: func(mockCtl *gomock.Controller) pipeline.ExecFn {
				mock := NewMocktestExecFns(mockCtl)
				mock.EXPECT().TerraformOutput(gomock.Any(), l).DoAndReturn(func(ctx context.Context, log hclog.Logger) (map[string]tfexec.OutputMeta, error) {
					return map[string]tfexec.OutputMeta{
						"key": {},
					}, nil
				})
				return MapTerraformOutput(mock.TerraformOutput)
			},
			assertFn: func(t *testing.T, byts []byte) {
				assert.NotEmpty(t, byts)
			},
		},
		"output - error": {
			execFn: func(mockCtl *gomock.Controller) pipeline.ExecFn {
				mock := NewMocktestExecFns(mockCtl)
				mock.EXPECT().TerraformOutput(gomock.Any(), l).Return(nil, terraformErr)
				return MapTerraformOutput(mock.TerraformOutput)
			},
			errExpected: terraformErr,
		},
		"state - happy path": {
			execFn: func(mockCtl *gomock.Controller) pipeline.ExecFn {
				mock := NewMocktestExecFns(mockCtl)
				mock.EXPECT().TerraformState(gomock.Any(), l).DoAndReturn(func(ctx context.Context, log hclog.Logger) (*tfjson.State, error) {
					return &tfjson.State{
						FormatVersion:    "v0.1.0",
						TerraformVersion: "abc",
					}, nil
				})
				return MapTerraformState(mock.TerraformState)
			},
			assertFn: func(t *testing.T, byts []byte) {
				var state tfjson.State
				err := json.Unmarshal(byts, &state)
				assert.NoError(t, err)
				assert.Equal(t, "abc", state.TerraformVersion)
			},
		},
		"state - error": {
			execFn: func(mockCtl *gomock.Controller) pipeline.ExecFn {
				mock := NewMocktestExecFns(mockCtl)
				mock.EXPECT().TerraformState(gomock.Any(), l).Return(nil, terraformErr)
				return MapTerraformState(mock.TerraformState)
			},
			errExpected: terraformErr,
		},
		"plan - happy path": {
			execFn: func(mockCtl *gomock.Controller) pipeline.ExecFn {
				mock := NewMocktestExecFns(mockCtl)
				mock.EXPECT().TerraformPlan(gomock.Any(), l).DoAndReturn(func(ctx context.Context, lg hclog.Logger) (*tfjson.Plan, error) {
					return &tfjson.Plan{
						FormatVersion:    "v0.1.0",
						TerraformVersion: "abc",
					}, nil
				})
				return MapTerraformPlan(mock.TerraformPlan)
			},
			assertFn: func(t *testing.T, byts []byte) {
				var state tfjson.Plan
				err := json.Unmarshal(byts, &state)
				assert.NoError(t, err)
				assert.Equal(t, "abc", state.TerraformVersion)
			},
		},
		"plan - error": {
			execFn: func(mockCtl *gomock.Controller) pipeline.ExecFn {
				mock := NewMocktestExecFns(mockCtl)
				mock.EXPECT().TerraformPlan(gomock.Any(), l).Return(nil, terraformErr)
				return MapTerraformPlan(mock.TerraformPlan)
			},
			errExpected: terraformErr,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			mockCtl, ctx := gomock.WithContext(ctx, t)

			execFn := test.execFn(mockCtl)

			byts, err := execFn(ctx, l, ui)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}

			assert.NoError(t, err)
			test.assertFn(t, byts)
		})
	}
}

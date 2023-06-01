package exec

import (
	context "context"
	"fmt"
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/powertoolsdev/mono/pkg/pipeline"
	"github.com/stretchr/testify/assert"
)

func Test_execInitFn_exec(t *testing.T) {
	l := NewMockhcLog(nil)
	ui := NewMockui(nil)
	errInit := fmt.Errorf("error init")

	tests := map[string]struct {
		execFn      func(*gomock.Controller) pipeline.ExecFn
		errExpected error
	}{
		"happy path": {
			execFn: func(mockCtl *gomock.Controller) pipeline.ExecFn {
				mock := NewMocktestExecFns(mockCtl)
				mock.EXPECT().Init(gomock.Any()).Return(nil)
				return MapInit(mock.Init)
			},
		},
		"error": {
			execFn: func(mockCtl *gomock.Controller) pipeline.ExecFn {
				mock := NewMocktestExecFns(mockCtl)
				mock.EXPECT().Init(gomock.Any()).Return(errInit)
				return MapInit(mock.Init)
			},
			errExpected: errInit,
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
			assert.Nil(t, byts)
		})
	}

}

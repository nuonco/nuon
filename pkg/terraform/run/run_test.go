package run

import (
	"fmt"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/terraform/workspace"
)

//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=run_mock_test.go -source=run_test.go -package=run
type hcLog interface {
	hclog.Logger
}

func TestNew(t *testing.T) {
	v := validator.New()

	wkspace := workspace.NewMockWorkspace(nil)
	log := NewMockhcLog(nil)
	outputSettings := generics.GetFakeObj[*OutputSettings]()

	tests := map[string]struct {
		errExpected error
		optsFn      func() []runOption
		assertFn    func(*testing.T, *run)
	}{
		"happy path": {
			optsFn: func() []runOption {
				return []runOption{
					WithWorkspace(wkspace),
					WithLogger(log),
					WithOutputSettings(outputSettings),
				}
			},
			assertFn: func(t *testing.T, e *run) {
				assert.Equal(t, wkspace, e.Workspace)
			},
		},
		"missing workspace": {
			optsFn: func() []runOption {
				return []runOption{
					WithLogger(log),
					WithOutputSettings(outputSettings),
				}
			},
			errExpected: fmt.Errorf("Workspace"),
		},
		"missing log": {
			optsFn: func() []runOption {
				return []runOption{
					WithWorkspace(wkspace),
					WithOutputSettings(outputSettings),
				}
			},
			errExpected: fmt.Errorf("Log"),
		},
		"missing output settings": {
			optsFn: func() []runOption {
				return []runOption{
					WithWorkspace(wkspace),
					WithLogger(log),
				}
			},
			errExpected: fmt.Errorf("OutputSettings"),
		},
	}

	for name, test := range tests {
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			e, err := New(v, test.optsFn()...)
			if test.errExpected != nil {
				assert.Error(t, err)
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
			test.assertFn(t, e)
		})
	}
}

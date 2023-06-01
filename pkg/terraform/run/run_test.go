package run

import (
	"fmt"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/waypoint-plugin-sdk/terminal"
	"github.com/powertoolsdev/mono/pkg/terraform/workspace"
	"github.com/stretchr/testify/assert"
)

//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=run_mock_test.go -source=run_test.go -package=run
type ui interface {
	terminal.UI
}

type hcLog interface {
	hclog.Logger
}

func TestNew(t *testing.T) {
	v := validator.New()

	wkspace := workspace.NewMockWorkspace(nil)
	ui := NewMockui(nil)
	log := NewMockhcLog(nil)

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
					WithUI(ui),
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
					WithUI(ui),
				}
			},
			errExpected: fmt.Errorf("Workspace"),
		},
		"missing ui": {
			optsFn: func() []runOption {
				return []runOption{
					WithWorkspace(wkspace),
					WithLogger(log),
				}
			},
			errExpected: fmt.Errorf("UI"),
		},
		"missing log": {
			optsFn: func() []runOption {
				return []runOption{
					WithWorkspace(wkspace),
					WithUI(ui),
				}
			},
			errExpected: fmt.Errorf("Log"),
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

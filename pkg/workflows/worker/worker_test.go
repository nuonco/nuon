package worker

import (
	"fmt"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	v := validator.New()

	cfg := generics.GetFakeObj[*Config]()
	act := &struct{}{}
	workflow := &struct{}{}

	tests := map[string]struct {
		optFns      func() []workerOption
		assertFn    func(*testing.T, *worker)
		errExpected error
	}{
		"happy path": {
			optFns: func() []workerOption {
				return []workerOption{
					WithConfig(cfg),
					WithActivity(act),
					WithWorkflow(workflow),
				}
			},
			assertFn: func(t *testing.T, wrk *worker) {
				assert.Equal(t, cfg, wrk.Config)
				assert.Equal(t, act, wrk.Activities[0])
				assert.Equal(t, workflow, wrk.Workflows[0])
			},
		},
		"invalid config": {
			optFns: func() []workerOption {
				badCfg := generics.GetFakeObj[*Config]()
				badCfg.Env = ""
				return []workerOption{
					WithConfig(badCfg),
					WithActivity(act),
					WithWorkflow(workflow),
				}
			},
			errExpected: fmt.Errorf("Env"),
		},
		"missing config": {
			optFns: func() []workerOption {
				return []workerOption{
					WithActivity(act),
					WithWorkflow(workflow),
				}
			},
			errExpected: fmt.Errorf("Config"),
		},
		"missing workflow": {
			optFns: func() []workerOption {
				return []workerOption{
					WithConfig(cfg),
					WithActivity(act),
				}
			},
			errExpected: fmt.Errorf("Workflow"),
		},
		"missing activity": {
			optFns: func() []workerOption {
				return []workerOption{
					WithConfig(cfg),
					WithWorkflow(workflow),
				}
			},
			errExpected: fmt.Errorf("Activities"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			opts := test.optFns()
			wrk, err := New(v, opts...)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}

			assert.NoError(t, err)
			test.assertFn(t, wrk)
		})
	}
}

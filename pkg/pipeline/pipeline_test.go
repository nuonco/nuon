package pipeline

import (
	"fmt"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/waypoint-plugin-sdk/terminal"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=pipeline_mock_test.go -source=pipeline_test.go -package=pipeline
type ui interface {
	terminal.UI
}

func TestNew(t *testing.T) {
	v := validator.New()
	l := zaptest.NewLogger(t)
	ui := NewMockui(nil)

	tests := map[string]struct {
		errExpected error
		optsFn      func() []pipelineOption
		assertFn    func(*testing.T, *Pipeline)
	}{
		"happy path": {
			optsFn: func() []pipelineOption {
				return []pipelineOption{
					WithLogger(l),
					WithUI(ui),
				}
			},
			assertFn: func(t *testing.T, p *Pipeline) {
				assert.Equal(t, l, p.Log)
				assert.Equal(t, ui, p.UI)
			},
		},
		"missing log": {
			optsFn: func() []pipelineOption {
				return []pipelineOption{
					WithUI(ui),
				}
			},
			errExpected: fmt.Errorf("Log"),
		},
		"missing ui": {
			optsFn: func() []pipelineOption {
				return []pipelineOption{
					WithLogger(l),
				}
			},
			errExpected: fmt.Errorf("UI"),
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

package pipeline

import (
	"fmt"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
)

//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=pipeline_mock_test.go -source=pipeline_test.go -package=pipeline
type hcLog interface {
	hclog.Logger
}

func TestNew(t *testing.T) {
	v := validator.New()
	l := NewMockhcLog(nil)

	tests := map[string]struct {
		errExpected error
		optsFn      func() []pipelineOption
		assertFn    func(*testing.T, *Pipeline)
	}{
		"happy path": {
			optsFn: func() []pipelineOption {
				return []pipelineOption{
					WithLogger(l),
				}
			},
			assertFn: func(t *testing.T, p *Pipeline) {
				assert.Equal(t, l, p.Log)
			},
		},
		"missing log": {
			optsFn: func() []pipelineOption {
				return []pipelineOption{}
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

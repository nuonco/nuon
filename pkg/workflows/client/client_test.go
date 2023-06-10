package client

import (
	"fmt"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/clients/temporal"
	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	v := validator.New()
	client := temporal.NewMockClient(nil)

	tests := map[string]struct {
		errExpected error
		optsFn      func() []workflowsClientOption
		assertFn    func(*testing.T, *workflowsClient)
	}{
		"happy path": {
			optsFn: func() []workflowsClientOption {
				return []workflowsClientOption{
					WithClient(client),
				}
			},
			assertFn: func(t *testing.T, w *workflowsClient) {
				assert.Equal(t, client, w.TemporalClient)
				assert.Equal(t, defaultAgent, w.Agent)
			},
		},
		"happy path - custom agent": {
			optsFn: func() []workflowsClientOption {
				return []workflowsClientOption{
					WithClient(client),
					WithAgent("abc"),
				}
			},
			assertFn: func(t *testing.T, w *workflowsClient) {
				assert.Equal(t, client, w.TemporalClient)
				assert.Equal(t, "abc", w.Agent)
			},
		},
		"missing client": {
			optsFn: func() []workflowsClientOption {
				return []workflowsClientOption{}
			},
			errExpected: fmt.Errorf("Client"),
		},
	}

	for name, test := range tests {
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			e, err := NewClient(v, test.optsFn()...)
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

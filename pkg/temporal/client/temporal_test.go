package temporal

import (
	"fmt"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestNew(t *testing.T) {
	namespace := generics.GetFakeObj[string]()
	logger := zaptest.NewLogger(t)
	address := generics.GetFakeObj[string]()

	tests := map[string]struct {
		optFns      func() []temporalOption
		errExpected error
		assertFn    func(*testing.T, Client)
	}{
		// NOTE(jm): we can only load this with lazy load on and assert, otherwise the connection will be
		// attempted and fail during testing
		"happy path": {
			optFns: func() []temporalOption {
				return []temporalOption{
					WithNamespace(namespace),
					WithLogger(logger),
					WithLazyLoad(true),
					WithAddr(address),
				}
			},
			assertFn: func(t *testing.T, client Client) {
				tClient, ok := client.(*temporal)
				assert.True(t, ok)
				assert.True(t, tClient.LazyLoad)
				assert.Equal(t, tClient.Addr, address)
				assert.Equal(t, tClient.Namespace, namespace)
				assert.Equal(t, tClient.Logger, logger)
			},
		},
		"sets lazy load": {
			optFns: func() []temporalOption {
				return []temporalOption{
					WithNamespace(namespace),
					WithLogger(logger),
					WithAddr(address),
					WithLazyLoad(true),
				}
			},
			assertFn: func(t *testing.T, client Client) {
				tClient, ok := client.(*temporal)
				assert.True(t, ok)
				assert.True(t, tClient.LazyLoad)
			},
		},
		"missing address": {
			optFns: func() []temporalOption {
				return []temporalOption{
					WithNamespace(namespace),
					WithLogger(logger),
				}
			},
			errExpected: fmt.Errorf("Addr"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			v := validator.New()
			opts := test.optFns()
			_, err := New(v, opts...)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
		})
	}
}

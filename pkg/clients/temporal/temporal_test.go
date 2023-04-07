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
	temporalHost := generics.GetFakeObj[string]()
	namespace := generics.GetFakeObj[string]()
	logger := zaptest.NewLogger(t)

	tests := map[string]struct {
		optFns      func() []temporalOption
		errExpected error
	}{
		"missing address": {
			optFns: func() []temporalOption {
				return []temporalOption{
					WithNamespace(namespace),
					WithLogger(logger),
				}
			},
			errExpected: fmt.Errorf("Addr"),
		},
		"missing namespace": {
			optFns: func() []temporalOption {
				return []temporalOption{
					WithLogger(logger),
					WithAddr(temporalHost),
				}
			},
			errExpected: fmt.Errorf("Namespace"),
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

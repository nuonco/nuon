package jobs

import (
	"fmt"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/temporal/client"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	mockClient := temporal.NewMockClient(nil)

	tests := map[string]struct {
		optFns      func() []managerOption
		errExpected error
		assertFn    func(*testing.T, *manager)
	}{
		"happy path": {
			optFns: func() []managerOption {
				return []managerOption{
					WithClient(mockClient),
				}
			},
			assertFn: func(t *testing.T, mgr *manager) {
				assert.NotNil(t, mgr.Client)
			},
		},
		"missing client": {
			optFns: func() []managerOption {
				return []managerOption{}
			},
			errExpected: fmt.Errorf("Client"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			v := validator.New()
			opts := test.optFns()
			commands, err := New(v, opts...)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
			test.assertFn(t, commands)
		})
	}
}

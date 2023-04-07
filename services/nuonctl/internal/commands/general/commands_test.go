package general

import (
	"fmt"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/services/nuonctl/internal/repos/temporal"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	temp := temporal.NewMockRepo(nil)

	tests := map[string]struct {
		optFns      func() []commandsOption
		errExpected error
		assertFn    func(*testing.T, *commands)
	}{
		"happy path": {
			optFns: func() []commandsOption {
				return []commandsOption{
					WithTemporalRepo(temp),
				}
			},
			assertFn: func(t *testing.T, cmd *commands) {
				assert.NotNil(t, cmd.Temporal)
			},
		},
		"missing temporal": {
			optFns: func() []commandsOption {
				return []commandsOption{}
			},
			errExpected: fmt.Errorf("Temporal"),
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

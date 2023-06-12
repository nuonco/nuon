package instances

import (
	"fmt"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/services/api/internal/services"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	mockSvc := services.NewMockInstanceService(nil)

	tests := map[string]struct {
		optFns      func() []serverOption
		errExpected error
		assertFn    func(*testing.T, *server)
	}{
		"happy path": {
			optFns: func() []serverOption {
				return []serverOption{
					WithService(mockSvc),
				}
			},
			assertFn: func(t *testing.T, mgr *server) {
				assert.Equal(t, mockSvc, mgr.Svc)
			},
		},
		"missing service": {
			optFns: func() []serverOption {
				return []serverOption{}
			},
			errExpected: fmt.Errorf("Svc"),
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

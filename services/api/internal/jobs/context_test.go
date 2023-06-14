package jobs

import (
	"context"
	"fmt"
	"testing"

	"github.com/powertoolsdev/mono/pkg/temporal/client"
	"github.com/stretchr/testify/assert"
)

func TestFromContext(t *testing.T) {
	tClient := temporal.NewMockClient(nil)

	tests := map[string]struct {
		errExpected error
		ctxFn       func() context.Context
		assertFn    func(*testing.T, *manager)
	}{
		"happy path": {
			ctxFn: func() context.Context {
				ctx := context.Background()
				return context.WithValue(ctx, temporal.ContextKey{}, tClient)
			},
			assertFn: func(t *testing.T, m *manager) {
				assert.Equal(t, tClient, m.Client)
			},
		},
		"no context found": {
			ctxFn: func() context.Context {
				return context.Background()
			},
			errExpected: fmt.Errorf("no temporal client"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := test.ctxFn()
			mgr, err := FromContext(ctx)
			if test.errExpected != nil {
				assert.Error(t, err)
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}

			assert.NoError(t, err)
			test.assertFn(t, mgr)
		})
	}
}

package orgcontext

import (
	"context"
	"testing"

	"github.com/powertoolsdev/go-generics"
	"github.com/stretchr/testify/assert"
)

func TestGetContext(t *testing.T) {
	orgCtx := generics.GetFakeObj[*Context]()

	tests := map[string]struct {
		ctxFn       func() context.Context
		errExpected error
		assertFn    func(*testing.T, *Context)
	}{
		"happy path": {
			ctxFn: func() context.Context {
				ctx := context.Background()
				return context.WithValue(ctx, orgContextKey{}, orgCtx)
			},
			assertFn: func(t *testing.T, res *Context) {
				assert.Equal(t, orgCtx, res)
			},
		},
		"no context value set": {
			ctxFn: func() context.Context {
				return context.Background()
			},
			errExpected: errNotFound,
		},
		"invalid context set": {
			ctxFn: func() context.Context {
				ctx := context.Background()
				return context.WithValue(ctx, orgContextKey{}, "abc")
			},
			errExpected: errInvalid,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := test.ctxFn()

			orgCtx, err := Get(ctx)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
			test.assertFn(t, orgCtx)
		})
	}
}

package context

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

func TestParseMetadata(t *testing.T) {
	tests := map[string]struct {
		ctxFn       func() context.Context
		assertFn    func(*testing.T, context.Context)
		errExpected error
	}{
		"happy path": {
			ctxFn: func() context.Context {
				ctx := context.Background()
				meta := metadata.New(map[string]string{userIDHeaderKey: "nuon-user-id"})
				return metadata.NewIncomingContext(ctx, meta)
			},
			assertFn: func(t *testing.T, ctx context.Context) {
				obj := ctx.Value(UserIDContext{})
				assert.NotNil(t, obj)
				userID, ok := obj.(string)
				assert.True(t, ok)
				assert.Equal(t, "nuon-user-id", userID)
			},
		},
		"no metadata": {
			ctxFn: func() context.Context {
				return context.Background()
			},
			errExpected: fmt.Errorf(codes.DataLoss.String()),
		},
		"no header key set": {
			ctxFn: func() context.Context {
				ctx := context.Background()
				meta := metadata.New(map[string]string{})
				return metadata.NewIncomingContext(ctx, meta)
			},
			errExpected: fmt.Errorf(codes.InvalidArgument.String()),
		},
		"empty header set": {
			ctxFn: func() context.Context {
				ctx := context.Background()
				meta := metadata.New(map[string]string{userIDHeaderKey: ""})
				return metadata.NewIncomingContext(ctx, meta)
			},
			errExpected: fmt.Errorf(codes.InvalidArgument.String()),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			reqCtx := test.ctxFn()

			respCtx, err := ParseMetadata(reqCtx)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.Nil(t, err)
			test.assertFn(t, respCtx)
		})
	}
}

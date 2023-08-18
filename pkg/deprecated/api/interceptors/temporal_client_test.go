package interceptors

import (
	"context"
	"testing"

	"github.com/bufbuild/connect-go"
	"github.com/powertoolsdev/mono/pkg/temporal/client"
	"github.com/stretchr/testify/assert"
)

func TestNewTemporalClientInterceptor(t *testing.T) {
	client := temporal.NewMockClient(nil)

	t.Run("sets client", func(t *testing.T) {
		fn := func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			val := ctx.Value(temporal.ContextKey{})
			assert.NotNil(t, val)
			assert.Equal(t, client, val)
			return nil, nil
		}

		int := NewTemporalClientInterceptor(client)
		int(fn)
	})
}

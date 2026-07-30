package waitforevent

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

func TestTimeoutDerivation(t *testing.T) {
	t.Run("absent is unbounded", func(t *testing.T) {
		require.Equal(t, signal.UnboundedTimeout, signal.DeriveTimeout(&Signal{}))
	})

	t.Run("explicit timeout propagates", func(t *testing.T) {
		const timeout = 3 * time.Hour
		require.Equal(t, timeout, signal.DeriveTimeout(&Signal{WaitTimeout: timeout}))
	})
}

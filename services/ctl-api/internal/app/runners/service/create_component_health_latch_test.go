package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// The latch exists so a wedged rollout that Argo reports as benign
// "progressing" keeps its degraded verdict until something genuinely reports
// healthy. It must not do that to an unknown report: unknown means the runner
// could not assess the resource this cycle, and replacing it with a remembered
// verdict republishes a stale diagnosis under a fresh timestamp.
func TestLatchHealth(t *testing.T) {
	t.Parallel()

	degraded := priorHealth{health: "degraded", message: "CrashLoopBackOff", nativeStatus: "Degraded"}

	t.Run("progressing latches to the worse prior", func(t *testing.T) {
		health, message, _ := latchHealth("progressing", "rolling out", "Progressing", degraded, true)
		assert.Equal(t, "degraded", health)
		assert.Equal(t, "CrashLoopBackOff", message)
	})

	t.Run("unknown is not latched", func(t *testing.T) {
		health, message, native := latchHealth("unknown", "probe could not run", "", degraded, true)
		assert.Equal(t, "unknown", health, "a resource nobody could assess is unknown, not degraded")
		assert.Equal(t, "probe could not run", message, "the stale diagnosis must not be re-attributed")
		assert.Empty(t, native)
	})

	t.Run("healthy clears the latch", func(t *testing.T) {
		health, _, _ := latchHealth("healthy", "all replicas ready", "Healthy", degraded, true)
		assert.Equal(t, "healthy", health)
	})

	t.Run("a worse incoming report wins", func(t *testing.T) {
		health, message, _ := latchHealth("unhealthy", "image pull failed", "Degraded", degraded, true)
		assert.Equal(t, "unhealthy", health)
		assert.Equal(t, "image pull failed", message)
	})

	t.Run("no prior passes through", func(t *testing.T) {
		health, _, _ := latchHealth("progressing", "rolling out", "Progressing", priorHealth{}, false)
		assert.Equal(t, "progressing", health)
	})
}

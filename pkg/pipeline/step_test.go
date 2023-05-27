package pipeline

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestPipeline_AddStep(t *testing.T) {
	v := validator.New()
	l := zaptest.NewLogger(t)
	ui := NewMockui(nil)

	t.Run("successfully adds test to internal", func(t *testing.T) {
		pipe, err := New(v, WithLogger(l),
			WithUI(ui),
		)
		assert.NoError(t, err)

		step := generics.GetFakeObj[*Step]()
		pipe.AddStep(step)
		assert.Equal(t, 1, len(pipe.Steps))
		assert.Equal(t, step, pipe.Steps[0])
	})
}

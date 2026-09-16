package pipeline

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"

	"github.com/nuonco/nuon/pkg/generics"
)

func TestPipeline_AddStep(t *testing.T) {
	v := validator.New()
	l := NewMockhcLog(nil)

	t.Run("successfully adds step", func(t *testing.T) {
		pipe, err := New(v, WithLogger(l))
		assert.NoError(t, err)

		step := generics.GetFakeObj[*Step]()
		pipe.AddStep(step)
		assert.Equal(t, 1, len(pipe.Steps))
		assert.Equal(t, step, pipe.Steps[0])
	})
}

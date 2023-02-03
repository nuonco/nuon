//go:build integrationlocal

package runner

import (
	"context"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

func TestRunner_Run_Int(t *testing.T) {
	t.Parallel()
	r, err := New(
		validator.New(),
		WithBucket("jdt-test"),
		WithKey("request.json"),
		WithRegion("us-west-2"),
	)
	assert.NoError(t, err)

	m, err := r.Run(context.Background())
	assert.NoError(t, err)

	assert.Equal(t, map[string]string{"test_number": "1", "test_string": "test_string"}, m)
}

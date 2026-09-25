package hooks

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestRetryDBRead(t *testing.T) {
	t.Run("succeeds first attempt without retrying", func(t *testing.T) {
		calls := 0
		err := retryDBRead(context.Background(), func() error {
			calls++
			return nil
		})
		require.NoError(t, err)
		assert.Equal(t, 1, calls)
	})

	t.Run("recovers from transient failure", func(t *testing.T) {
		calls := 0
		err := retryDBRead(context.Background(), func() error {
			calls++
			if calls < 2 {
				return errors.New("transient connection failure")
			}
			return nil
		})
		require.NoError(t, err)
		assert.Equal(t, 2, calls)
	})

	t.Run("gives up after bounded attempts", func(t *testing.T) {
		calls := 0
		persistent := errors.New("db down")
		err := retryDBRead(context.Background(), func() error {
			calls++
			return persistent
		})
		require.ErrorIs(t, err, persistent)
		assert.Equal(t, dbReadRetryAttempts, calls)
	})

	t.Run("record not found is a domain outcome, never retried", func(t *testing.T) {
		calls := 0
		err := retryDBRead(context.Background(), func() error {
			calls++
			return gorm.ErrRecordNotFound
		})
		require.ErrorIs(t, err, gorm.ErrRecordNotFound)
		assert.Equal(t, 1, calls)
	})

	t.Run("cancelled context stops retrying", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		calls := 0
		err := retryDBRead(ctx, func() error {
			calls++
			cancel()
			return errors.New("transient")
		})
		require.Error(t, err)
		assert.Equal(t, 1, calls)
	})
}

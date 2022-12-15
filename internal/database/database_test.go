package database

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func getDatabaseOpts() []DBConfigFunc {
	return []DBConfigFunc{
		WithDBName("api"),
		WithUser("postgres"),
		WithHost("localhost"),
		WithPort("5432"),
	}
}

// Test that the database initializer will respect a password function
func TestDatabasePasswordFnUsed(t *testing.T) {
	cnt := 0
	fn := func(ctx context.Context, cfg Config) (string, error) {
		cnt += 1
		return "abc", nil
	}
	opts := getDatabaseOpts()
	opts = append(opts, WithPasswordFn(fn))

	db, err := New(opts...)
	assert.Equal(t, 1, cnt)
	assert.Nil(t, err)
	assert.NotNil(t, db)
}

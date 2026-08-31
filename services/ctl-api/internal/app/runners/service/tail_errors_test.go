package service

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"net"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
)

func TestIsTransientTailProbeError(t *testing.T) {
	tests := map[string]struct {
		err       error
		transient bool
	}{
		"deadline":             {err: context.DeadlineExceeded, transient: true},
		"wrapped deadline":     {err: fmt.Errorf("query: %w", context.DeadlineExceeded), transient: true},
		"bad connection":       {err: driver.ErrBadConn, transient: true},
		"EOF":                  {err: io.EOF, transient: true},
		"unexpected EOF":       {err: io.ErrUnexpectedEOF, transient: true},
		"network timeout":      {err: &net.DNSError{IsTimeout: true}, transient: true},
		"connection exception": {err: &pgconn.PgError{Code: "08006"}, transient: true},
		"resource exhausted":   {err: &pgconn.PgError{Code: "53300"}, transient: true},
		"database shutdown":    {err: &pgconn.PgError{Code: "57P01"}, transient: true},
		"canceled":             {err: context.Canceled},
		"query error":          {err: errors.New("column does not exist")},
		"invalid SQL":          {err: &pgconn.PgError{Code: "42601"}},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.transient, isTransientTailProbeError(tt.err))
		})
	}
}

func TestTailProbeQueryError(t *testing.T) {
	queryErr := errors.New("driver timeout")

	t.Run("expired query", func(t *testing.T) {
		ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
		defer cancel()

		err := tailProbeQueryError(ctx, queryErr)
		assert.ErrorIs(t, err, context.DeadlineExceeded)
		assert.ErrorIs(t, err, queryErr)
		assert.True(t, isTransientTailProbeError(err))
	})

	t.Run("active query", func(t *testing.T) {
		err := tailProbeQueryError(context.Background(), queryErr)
		assert.Same(t, queryErr, err)
	})
}

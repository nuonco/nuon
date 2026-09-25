package hooks

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

const (
	dbReadRetryAttempts  = 3
	dbReadRetryBaseDelay = 100 * time.Millisecond
)

// retryDBRead runs a read-only DB operation with bounded retries so a
// transient query failure doesn't drop a notification. It must only wrap
// operations that happen strictly before any external delivery (Slack post,
// webhook send) — rerunning anything past that point would duplicate sends.
//
// gorm.ErrRecordNotFound is a domain outcome, not a transient failure, and is
// returned immediately. Context cancellation also short-circuits.
func retryDBRead(ctx context.Context, op func() error) error {
	var err error
	for attempt := 0; attempt < dbReadRetryAttempts; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(dbReadRetryBaseDelay << (attempt - 1)):
			}
		}
		err = op()
		if err == nil || errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if ctx.Err() != nil {
			return err
		}
	}
	return err
}

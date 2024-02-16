package migrations

import (
	"context"
)

func (a *Migrations) migration021NoopDatadogTest(ctx context.Context) error {
	// Noop migration to ensure events make it to datadog
	return nil
}

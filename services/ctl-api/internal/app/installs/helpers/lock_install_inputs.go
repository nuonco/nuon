package helpers

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

// LockInstallInputs serializes an install's inputs writers for the rest of the
// transaction. Rows are whole-snapshot appends and readers take the newest, so
// unserialized read-modify-append drops concurrent writes. The lock releases at the
// outermost commit, so callers taking several must order them.
func LockInstallInputs(ctx context.Context, tx *gorm.DB, installID string) error {
	if err := tx.WithContext(ctx).
		Exec("select pg_advisory_xact_lock(hashtext(?)::bigint)", "install_inputs:"+installID).
		Error; err != nil {
		return fmt.Errorf("unable to lock install inputs for %s: %w", installID, err)
	}
	return nil
}

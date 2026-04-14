package structured_errors

import (
	"encoding/json"
	"fmt"

	"gorm.io/gorm"
)

// Append adds errors to any model that has an `Errors` JSONB column.
// It performs a read-modify-write to append new errors to the existing slice.
func Append(db *gorm.DB, model any, modelID string, errs []CompositeError) error {
	if len(errs) == 0 {
		return nil
	}

	// Read current errors
	var existing CompositeErrors
	var raw json.RawMessage
	res := db.Model(model).Where("id = ?", modelID).Pluck("errors", &raw)
	if res.Error != nil {
		return fmt.Errorf("unable to read errors: %w", res.Error)
	}

	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &existing); err != nil {
			return fmt.Errorf("unable to unmarshal existing errors: %w", err)
		}
	}

	// Append new errors
	existing = append(existing, errs...)

	// Write back
	res = db.Model(model).Where("id = ?", modelID).Update("errors", existing)
	if res.Error != nil {
		return fmt.Errorf("unable to update errors: %w", res.Error)
	}

	return nil
}

// Clear moves current errors into their History field and resets the errors slice.
// This is used on retry to preserve error history while clearing the active errors.
func Clear(db *gorm.DB, model any, modelID string) error {
	// Read current errors
	var existing CompositeErrors
	var raw json.RawMessage
	res := db.Model(model).Where("id = ?", modelID).Pluck("errors", &raw)
	if res.Error != nil {
		return fmt.Errorf("unable to read errors: %w", res.Error)
	}

	if len(raw) == 0 {
		return nil
	}

	if err := json.Unmarshal(raw, &existing); err != nil {
		return fmt.Errorf("unable to unmarshal existing errors: %w", err)
	}

	if len(existing) == 0 {
		return nil
	}

	// Move each error into its own history and clear it
	for i := range existing {
		existing[i].History = append(existing[i].History, existing[i])
		existing[i].Summary = ""
		existing[i].Detail = ""
		existing[i].Metadata = nil
	}

	// Write cleared errors (with history preserved)
	cleared := CompositeErrors{}
	res = db.Model(model).Where("id = ?", modelID).Update("errors", cleared)
	if res.Error != nil {
		return fmt.Errorf("unable to clear errors: %w", res.Error)
	}

	return nil
}

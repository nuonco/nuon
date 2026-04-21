package labels

import (
	"encoding/json"

	"gorm.io/gorm"
)

// WithLabels returns a GORM scope that filters rows where the JSONB column
// contains all the specified key-value pairs (PostgreSQL @> containment).
// A value of "*" is treated as a wildcard — it matches any value for that key
// (uses jsonb_exists() to check key existence).
// Returns a no-op scope if lbls is nil or empty.
func WithLabels(column string, lbls Labels) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if len(lbls) == 0 {
			return db
		}

		// Split into exact matches and wildcard (key-only) matches.
		exact := make(Labels)
		var wildcardKeys []string
		for k, v := range lbls {
			if v == "*" {
				wildcardKeys = append(wildcardKeys, k)
			} else {
				exact[k] = v
			}
		}

		// Apply exact containment for key=value pairs.
		if len(exact) > 0 {
			jsonBytes, err := json.Marshal(exact)
			if err != nil {
				_ = db.AddError(err)
				return db
			}
			db = db.Where(column+" @> ?::jsonb", string(jsonBytes))
		}

		// Apply key-existence checks for wildcard entries.
		for _, key := range wildcardKeys {
			db = db.Where("jsonb_exists("+column+", ?)", key)
		}

		return db
	}
}

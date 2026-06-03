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

// WithoutLabels is the negative counterpart of WithLabels. It filters rows
// where the JSONB column does NOT match any of the specified key-value pairs.
// A value of "*" rejects rows where the key is present at all
// (jsonb_exists). Other values reject only rows where the exact pair is set.
// Each NotMatchLabels entry is independent; a row is excluded if ANY entry
// matches it (mirrors Selector.Matches in-Go semantics).
// Returns a no-op scope if lbls is nil or empty.
func WithoutLabels(column string, lbls Labels) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if len(lbls) == 0 {
			return db
		}
		for k, v := range lbls {
			if v == "*" {
				db = db.Where("NOT jsonb_exists("+column+", ?)", k)
				continue
			}
			pair, err := json.Marshal(Labels{k: v})
			if err != nil {
				_ = db.AddError(err)
				return db
			}
			db = db.Where("NOT ("+column+" @> ?::jsonb)", string(pair))
		}
		return db
	}
}

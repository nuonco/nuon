package labels

import (
	"encoding/json"

	"gorm.io/gorm"
)

func WithLabels(column string, lbls Labels) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if len(lbls) == 0 {
			return db
		}

		exact := make(Labels)
		var wildcardKeys []string
		for k, v := range lbls {
			if v == "*" {
				wildcardKeys = append(wildcardKeys, k)
			} else {
				exact[k] = v
			}
		}

		if len(exact) > 0 {
			jsonBytes, err := json.Marshal(exact)
			if err != nil {
				_ = db.AddError(err)
				return db
			}
			db = db.Where(column+" @> ?::jsonb", string(jsonBytes))
		}

		for _, key := range wildcardKeys {
			db = db.Where("jsonb_exists("+column+", ?)", key)
		}

		return db
	}
}

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

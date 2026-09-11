package generics

import (
	"encoding/json"
	"fmt"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

func ToJSON(val string) []byte {
	var contents []byte
	if len(val) > 0 {
		contents = []byte(val)
	}

	return contents
}

// MergeJSONBMetadata reads a JSONB column's "metadata" key, merges the provided
// key-value pairs, and writes the column back. Nil values remove their keys. The
// model must be a pointer to a GORM model (e.g. &app.QueueSignal{}).
func MergeJSONBMetadata(db *gorm.DB, model any, id string, field string, metadata map[string]any) error {
	// Build a jsonb_set chain that merges each key into the metadata sub-object.
	// First ensure the metadata key exists as an object, then set each key.
	expr := fmt.Sprintf(
		"jsonb_set(COALESCE(%s::jsonb, '{}'::jsonb), '{metadata}', COALESCE(%s::jsonb -> 'metadata', '{}'::jsonb))",
		field, field,
	)

	args := make([]any, 0, len(metadata))
	for k, v := range metadata {
		if v == nil {
			expr = fmt.Sprintf("(%s #- ?::text[])", expr)
			args = append(args, pq.Array([]string{"metadata", k}))
			continue
		}

		valJSON, err := json.Marshal(v)
		if err != nil {
			return fmt.Errorf("unable to marshal metadata value for key %s: %w", k, err)
		}
		expr = fmt.Sprintf("jsonb_set(%s, '{metadata,%s}', ?::jsonb)", expr, k)
		args = append(args, string(valJSON))
	}

	if res := db.
		Model(model).
		Where("id = ?", id).
		Update(field, gorm.Expr(expr, args...)); res.Error != nil {
		return fmt.Errorf("unable to merge metadata: %w", res.Error)
	}

	return nil
}

func WhereJSONBStatus(field string, status string) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(fmt.Sprintf("%s->>'status' = ?", field), status)
	}
}

func WhereJSONBStatusNot(field string, status string) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(fmt.Sprintf("%s->>'status' != ?", field), status)
	}
}

func WhereJSONBStatusIn(field string, statuses ...string) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(fmt.Sprintf("%s->>'status' IN ?", field), statuses)
	}
}

func WhereJSONBStatusNotIn(field string, statuses ...string) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(fmt.Sprintf("%s->>'status' NOT IN ?", field), statuses)
	}
}

func SetJSONBMetadataKey(db *gorm.DB, field, key string, value any) *gorm.DB {
	valJSON, err := json.Marshal(value)
	if err != nil {
		_ = db.AddError(fmt.Errorf("unable to marshal value for key %s: %w", key, err))
		return db
	}

	expr := fmt.Sprintf(
		"jsonb_set(jsonb_set(COALESCE(%s, '{}'::jsonb), '{metadata}', COALESCE(%s->'metadata', '{}'::jsonb)), ?::text[], ?::jsonb)",
		field,
		field,
	)
	return db.UpdateColumn(field, gorm.Expr(expr, pq.Array([]string{"metadata", key}), string(valJSON)))
}

func DeleteJSONBMetadataKey(db *gorm.DB, field, key string) *gorm.DB {
	expr := fmt.Sprintf(
		"jsonb_set(COALESCE(%s, '{}'::jsonb), '{metadata}', COALESCE(%s->'metadata', '{}'::jsonb) - ?)",
		field,
		field,
	)
	return db.UpdateColumn(field, gorm.Expr(expr, key))
}

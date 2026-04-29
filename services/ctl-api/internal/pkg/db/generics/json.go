package generics

import (
	"encoding/json"
	"fmt"

	"gorm.io/gorm"
)

func ToJSON(val string) []byte {
	var contents []byte
	if len(val) > 0 {
		contents = []byte(val)
	}

	return contents
}

// JSONBQuerier helps build JSONB queries
type JSONBQuerier struct {
	*gorm.DB
}

// NewJSONBQuery creates a new JSONB query helper
func NewJSONBQuery(db *gorm.DB) *JSONBQuerier {
	return &JSONBQuerier{db}
}

type JSONBQuery struct {
	Operator string
	Field    string
	Path     string
	Value    any
}

// WhereJSON adds a WHERE condition for JSONB field
func (jq *JSONBQuerier) WhereJSON(queryArgs JSONBQuery) *gorm.DB {
	var query string
	switch queryArgs.Operator {
	case "=":
		query = fmt.Sprintf("%s->>'%s' = ?", queryArgs.Field, queryArgs.Path)
	case "!=":
		query = fmt.Sprintf("%s->>'%s' != ?", queryArgs.Field, queryArgs.Path)
	case ">", "<", ">=", "<=":
		query = fmt.Sprintf("(%s->>'%s')::numeric %s ?", queryArgs.Field, queryArgs.Path, queryArgs.Operator)
	case "LIKE", "ILIKE":
		query = fmt.Sprintf("%s->>'%s' %s ?", queryArgs.Field, queryArgs.Path, queryArgs.Operator)
	case "IN":
		query = fmt.Sprintf("%s->>'%s' IN ?", queryArgs.Field, queryArgs.Path)
	case "@>":
		query = fmt.Sprintf("%s @> ?", queryArgs.Field)
	case "?":
		query = fmt.Sprintf("%s ? ?", queryArgs.Field)
	default:
		query = fmt.Sprintf("%s->>'%s' = ?", queryArgs.Field, queryArgs.Path)
	}

	return jq.Where(query, queryArgs.Value)
}

// WhereJSONPath queries nested JSON paths
func (jq *JSONBQuerier) WhereJSONPath(field string, path []string, operator string, value interface{}) *gorm.DB {
	pathStr := "{" + join(path, ",") + "}"
	query := fmt.Sprintf("%s#>>? %s ?", field, operator)
	return jq.Where(query, pathStr, value)
}

// MergeJSONBMetadata reads a JSONB column's "metadata" key, merges the provided
// key-value pairs, and writes the column back. The model must be a pointer to a
// GORM model (e.g. &app.QueueSignal{}).
func MergeJSONBMetadata(db *gorm.DB, model any, id string, field string, metadata map[string]any) error {
	// Build a jsonb_set chain that merges each key into the metadata sub-object.
	// Start from the existing column value, defaulting to '{}'.
	expr := fmt.Sprintf("COALESCE(%s::jsonb, '{}'::jsonb)", field)

	args := make([]any, 0, len(metadata))
	for k, v := range metadata {
		valJSON, err := json.Marshal(v)
		if err != nil {
			return fmt.Errorf("unable to marshal metadata value for key %s: %w", k, err)
		}
		expr = fmt.Sprintf("jsonb_set(%s, '{metadata,%s}', ?::jsonb)", expr, k)
		args = append(args, string(valJSON))
	}

	args = append(args, id)
	if res := db.
		Model(model).
		Where("id = ?", id).
		Update(field, gorm.Expr(expr, args...)); res.Error != nil {
		return fmt.Errorf("unable to merge metadata: %w", res.Error)
	}

	return nil
}

// Helper function to join strings
func join(strs []string, sep string) string {
	result := ""
	for i, s := range strs {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}

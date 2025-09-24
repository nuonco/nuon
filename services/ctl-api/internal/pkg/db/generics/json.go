package generics

import (
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

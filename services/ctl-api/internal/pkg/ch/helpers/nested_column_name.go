package helpers

import "fmt"

func NestedColumnName(parent, child string) string {
	// return a string formatted for SQL queries:
	// e.g.
	// - parent['child']
	// - parent['nested.child']
	// NOTE: only a single level of nesting is supported
	return fmt.Sprintf("%s['%s']", parent, child)
}

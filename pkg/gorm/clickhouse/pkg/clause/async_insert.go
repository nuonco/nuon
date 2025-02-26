package clause

import (
	"regexp"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AsyncInsert struct {
}

// Name where clause name
func (AsyncInsert) Name() string {
	return "SETTINGS"
}

// Build build where clause
func (a AsyncInsert) Build(builder clause.Builder) {
	builder.WriteString(" SETTINGS async_insert=1, wait_for_async_insert=1")
}

// ModifyStatement implements the StatementModifier interface
func (a AsyncInsert) ModifyStatement(stmt *gorm.Statement) {
	// Add ourselves to the list of clauses to be processed
	stmt.Clauses["SETTINGS"] = clause.Clause{Expression: a}
}

func (ai AsyncInsert) Merge(expr clause.Expression) {
}

// MergeClause implements the clause.Interface
func (a AsyncInsert) MergeClause(c *clause.Clause) {
	// Only set Expression if it's nil to avoid overwriting
	if c.Expression == nil {
		c.Expression = a
	}
}

// Register registers the AsyncInsert clause with GORM
func Register(db *gorm.DB) {
	// Register a callback that runs after the SQL is generated
	db.Callback().Create().After("gorm:create").Register("clickhouse:async_insert", func(db *gorm.DB) {
		// Check if our clause was added to this query
		if _, ok := db.Statement.Clauses["SETTINGS"]; ok && db.Statement.SQL.String() != "" {
			sql := db.Statement.SQL.String()

			// Use a regex to find the table name in the INSERT statement
			re := regexp.MustCompile(`(?i)INSERT\s+INTO\s+(?:\x60?([^\s\(]+)\x60?)`)
			matches := re.FindStringSubmatchIndex(sql)

			if len(matches) >= 4 {
				// matches[2] and matches[3] represent the position of the table name
				tableEndPos := matches[3]

				// Insert our SETTINGS clause right after the table name
				newSQL := sql[:tableEndPos] + " SETTINGS async_insert=1, wait_for_async_insert=1" + sql[tableEndPos:]

				// Reset the SQL buffer and write our new SQL
				db.Statement.SQL.Reset()
				db.Statement.SQL.WriteString(newSQL)

				// We don't need to modify the vars as they remain the same
				// The incorrect db.Statement.AddVar(db) line is removed
			}
		}
	})
}

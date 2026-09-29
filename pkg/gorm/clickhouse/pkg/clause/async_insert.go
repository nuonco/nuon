package clause

import (
	"regexp"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AsyncInsert struct {
}

func (AsyncInsert) Name() string {
	return "SETTINGS"
}

func (a AsyncInsert) Build(builder clause.Builder) {
	builder.WriteString(" SETTINGS async_insert=1, wait_for_async_insert=1")
}

func (a AsyncInsert) ModifyStatement(stmt *gorm.Statement) {
	stmt.Clauses["SETTINGS"] = clause.Clause{Expression: a}
}

func (ai AsyncInsert) Merge(expr clause.Expression) {
}

func (a AsyncInsert) MergeClause(c *clause.Clause) {
	if c.Expression == nil {
		c.Expression = a
	}
}

func Register(db *gorm.DB) {
	db.Callback().Create().After("gorm:create").Register("clickhouse:async_insert", func(db *gorm.DB) {
		if _, ok := db.Statement.Clauses["SETTINGS"]; ok && db.Statement.SQL.String() != "" {
			sql := db.Statement.SQL.String()

			re := regexp.MustCompile(`(?i)INSERT\s+INTO\s+(?:\x60?([^\s\(]+)\x60?)`)
			matches := re.FindStringSubmatchIndex(sql)

			if len(matches) >= 4 {
				tableEndPos := matches[3]

				newSQL := sql[:tableEndPos] + " SETTINGS async_insert=1, wait_for_async_insert=1" + sql[tableEndPos:]

				db.Statement.SQL.Reset()
				db.Statement.SQL.WriteString(newSQL)
			}
		}
	})
}

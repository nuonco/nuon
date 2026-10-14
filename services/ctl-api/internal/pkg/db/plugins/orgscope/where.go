package orgscope

import (
	"regexp"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	orgIDColumnPattern = regexp.MustCompile(`(?i)(^|[^a-z0-9_])"?org_id"?([^a-z0-9_]|$)`)
	orgIDEquality      = regexp.MustCompile(`(?i)(?:^|[^a-z0-9_])"?org_id"?\s*=\s*\?`)
	sqlOr              = regexp.MustCompile(`(?i)\bor\b`)
	sqlNotOrgID        = regexp.MustCompile(`(?i)\bnot\s+(?:[\w."]+\.)?"?org_id"?\s*=`)
)

func whereScopedToOrg(stmt *gorm.Statement, orgID string) bool {
	where, ok := stmt.Clauses["WHERE"]
	if !ok || where.Expression == nil {
		return false
	}
	return expressionScopedToOrg(where.Expression, orgID)
}

func expressionScopedToOrg(expr clause.Expression, orgID string) bool {
	if expr == nil {
		return false
	}
	switch e := expr.(type) {
	case clause.Where, clause.AndConditions:
		for _, nested := range nestedOf(e) {
			if expressionScopedToOrg(nested, orgID) {
				return true
			}
		}
		return false
	case clause.OrConditions:
		if len(e.Exprs) == 0 {
			return false
		}
		for _, nested := range e.Exprs {
			if !expressionScopedToOrg(nested, orgID) {
				return false
			}
		}
		return true
	case clause.NotConditions:
		return false
	case clause.Expr:
		return sqlScopedToOrg(e.SQL, e.Vars, orgID)
	case clause.NamedExpr:
		return sqlScopedToOrg(e.SQL, e.Vars, orgID)
	case clause.Eq:
		return columnScopedToOrg(e.Column, e.Value, orgID)
	case clause.IN:
		return columnScopedToOrg(e.Column, e.Values, orgID)
	default:
		return false
	}
}

func nestedOf(expr clause.Expression) []clause.Expression {
	switch e := expr.(type) {
	case clause.Where:
		return e.Exprs
	case clause.AndConditions:
		return e.Exprs
	default:
		return nil
	}
}

func columnScopedToOrg(col any, value any, orgID string) bool {
	if !columnIsOrgID(col) {
		return false
	}
	switch v := value.(type) {
	case string:
		return v == orgID
	case []string:
		return len(v) > 0 && allOrg(v, orgID)
	case []any:
		return len(v) > 0 && allOrg(v, orgID)
	default:
		return false
	}
}

func allOrg[T any](values []T, orgID string) bool {
	for _, value := range values {
		s, ok := any(value).(string)
		if !ok || s != orgID {
			return false
		}
	}
	return true
}

// sqlScopedToOrg reports whether raw SQL already forces org_id to orgID.
// A top-level AND term of org_id = orgID counts. An OR term does not, so the plugin still adds its own.
func sqlScopedToOrg(sql string, vars []any, orgID string) bool {
	if sqlNotOrgID.MatchString(sql) || !orgIDColumnPattern.MatchString(sql) {
		return false
	}
	for _, part := range splitTopLevelAnd(sql) {
		if sqlOr.MatchString(part.sql) || !assignmentsMatchOrg(sql, part, vars, orgID) {
			continue
		}
		return true
	}
	return false
}

type sqlSpan struct {
	sql   string
	start int
}

func splitTopLevelAnd(sql string) []sqlSpan {
	lower := strings.ToLower(sql)
	parts := make([]sqlSpan, 0, 2)
	depth := 0
	start := 0
	for i := 0; i < len(sql); i++ {
		switch sql[i] {
		case '(':
			depth++
		case ')':
			if depth > 0 {
				depth--
			}
		default:
			if depth == 0 && i+5 <= len(lower) && lower[i:i+5] == " and " {
				parts = append(parts, sqlSpan{sql: sql[start:i], start: start})
				i += 4
				start = i + 1
			}
		}
	}
	return append(parts, sqlSpan{sql: sql[start:], start: start})
}

func assignmentsMatchOrg(full string, part sqlSpan, vars []any, orgID string) bool {
	matches := orgIDEquality.FindAllStringSubmatchIndex(part.sql, -1)
	if len(matches) == 0 {
		return false
	}
	for _, match := range matches {
		end := part.start + match[1]
		placeholder := strings.Count(full[:end], "?") - 1
		if placeholder < 0 || placeholder >= len(vars) {
			return false
		}
		value, ok := vars[placeholder].(string)
		if !ok || value != orgID {
			return false
		}
	}
	return true
}

func columnIsOrgID(col any) bool {
	switch c := col.(type) {
	case string:
		return isOrgIDName(c)
	case clause.Column:
		return isOrgIDName(c.Name)
	default:
		return false
	}
}

func isOrgIDName(name string) bool {
	name = strings.Trim(name, `"`)
	if i := strings.LastIndex(name, "."); i >= 0 {
		name = name[i+1:]
	}
	return strings.EqualFold(name, "org_id")
}

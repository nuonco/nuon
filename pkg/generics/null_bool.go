package generics

import "database/sql"

func NewNullBool(val bool) sql.NullBool {
	return sql.NullBool{
		Bool:  val,
		Valid: true,
	}
}

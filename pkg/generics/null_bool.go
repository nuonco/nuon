package generics

import "database/sql"

func NewNullBool(val bool) sql.NullBool {
	return sql.NullBool{
		Bool:  val,
		Valid: true,
	}
}

func NewNullBoolFromPtr(val *bool) sql.NullBool {
	if val == nil {
		return sql.NullBool{Valid: false}
	}

	return sql.NullBool{
		Bool:  *val,
		Valid: true,
	}
}

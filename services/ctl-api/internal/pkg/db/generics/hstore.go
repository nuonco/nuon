package generics

import "github.com/jackc/pgx/v5/pgtype"

func ToStringMap(val pgtype.Hstore) map[string]string {
	vals := make(map[string]string, 0)
	for k, v := range val {
		vals[k] = *v
	}

	return vals
}

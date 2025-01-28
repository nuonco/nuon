package generics

import (
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/powertoolsdev/mono/pkg/generics"
)

func ToStringMap(val pgtype.Hstore) map[string]string {
	vals := make(map[string]string, 0)
	for k, v := range val {
		vals[k] = *v
	}

	return vals
}

func ToPtrStringMap(val map[string]string) map[string]*string {
	vals := make(map[string]*string, 0)
	for k, v := range val {
		vals[k] = generics.ToPtr(v)
	}

	return vals
}

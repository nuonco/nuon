package generics

import (
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/powertoolsdev/mono/pkg/generics"
)

func ToStringMapAny(val pgtype.Hstore) map[string]any {
	vals := make(map[string]any, 0)
	for k, v := range val {
		vals[k] = *v
	}

	return vals
}

func ToStringMap(val pgtype.Hstore) map[string]string {
	vals := make(map[string]string, 0)
	for k, v := range val {
		vals[k] = *v
	}

	return vals
}

func ToHstore(val map[string]string) pgtype.Hstore {
	return pgtype.Hstore(ToPtrStringMap(val))
}

func ToPtrStringMap(val map[string]string) map[string]*string {
	vals := make(map[string]*string, 0)
	for k, v := range val {
		vals[k] = generics.ToPtr(v)
	}

	return vals
}

package orgs

import "github.com/powertoolsdev/go-generics"

func ToOptionalID(val string) *string {
	if val == "" {
		return nil
	}

	return generics.ToPtr(val)
}

package converters

import "github.com/powertoolsdev/go-generics"

func ToOptionalStr(val string) *string {
	if val == "" {
		return nil
	}

	return generics.ToPtr(val)
}

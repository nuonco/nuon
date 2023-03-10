package servers

import (
	"fmt"
	"reflect"

	"github.com/powertoolsdev/go-common/shortid"
)

// ensureShortID ensures that the value provided is actually a shortID, regardless of whether a UUID or ShortID was pass
// ed in.
func ensureShortID(val string) (string, error) {
	// attempt to parse a uuid
	shortID, err := shortid.ParseString(val)
	if err == nil {
		return shortID, nil
	}

	// ensure shortID
	_, err = shortid.ToUUID(val)
	if err == nil {
		return val, nil
	}

	return "", fmt.Errorf("value is neither a shortID or UUID: %s", val)
}

func EnsureShortID(obj any) error {
	typ := reflect.TypeOf(obj).Elem()

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if field.Tag.Get("shortid") != "ensure" {
			continue
		}

		val := reflect.ValueOf(obj).Elem().Field(i).Interface()
		strVal, ok := val.(string)
		if !ok {
			return fmt.Errorf("only supports strings")
		}

		shortID, err := ensureShortID(strVal)
		if err != nil {
			return fmt.Errorf("unable to ensure to shortID: %w", err)
		}

		reflect.ValueOf(obj).Elem().Field(i).SetString(shortID)
	}

	return nil
}

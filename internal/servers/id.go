package servers

import (
	"fmt"

	"github.com/powertoolsdev/go-common/shortid"
)

func EnsureShortID(val string) (string, error) {
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

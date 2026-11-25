package auth

import (
	"errors"
	"net/http"
)

// IsUnauthorized reports whether ctl-api rejected the token. Any other failure returns
// false: sending those users to the auth service loops, because it sees a valid cookie
// and redirects straight back.
func IsUnauthorized(err error) bool {
	var coded interface{ IsCode(int) bool }
	if !errors.As(err, &coded) {
		return false
	}
	return coded.IsCode(http.StatusUnauthorized)
}

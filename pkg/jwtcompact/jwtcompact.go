// Package jwtcompact rejects non-compact JWTs before any claim peek.
//
// Flattened or general JSON Serialization can carry claim values that a naive
// "split on '.' and decode the middle segment" reader will treat as the payload,
// while signature verifiers that accept JSON Serialization use a different
// payload. Requiring charset-strict compact serialization closes that gap.
package jwtcompact

import (
	"errors"
	"regexp"
	"strings"
)

var compact = regexp.MustCompile(`^[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+$`)

// Require returns an error unless token is a compact JWT (three base64url
// segments separated by dots, nothing else).
func Require(token string) error {
	if !compact.MatchString(strings.TrimSpace(token)) {
		return errors.New("token must be compact jwt")
	}
	return nil
}

package oidc

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/mitchellh/mapstructure"
)

type Audience []string

type Claims struct {
	Iss   string         `json:"iss"`
	Sub   string         `json:"sub"`
	Aud   Audience       `json:"aud"`
	Exp   int64          `json:"exp"`
	Iat   int64          `json:"iat"`
	Nbf   int64          `json:"nbf"`
	Jti   string         `json:"jti"`
	Extra map[string]any `json:",remain"`
}

func DecodeClaims(raw map[string]any) (Claims, error) {
	var claims Claims
	if err := Decode(raw, &claims); err != nil {
		return Claims{}, err
	}

	return claims, nil
}

func Decode(raw map[string]any, out any) error {
	dec, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName:    "json",
		DecodeHook: audienceHook,
		Result:     out,
	})
	if err != nil {
		return fmt.Errorf("unable to create claims decoder: %w", err)
	}

	if err := dec.Decode(raw); err != nil {
		return fmt.Errorf("unable to decode claims: %w", err)
	}

	return nil
}

func audienceHook(from, to reflect.Type, data any) (any, error) {
	if to != reflect.TypeOf(Audience{}) {
		return data, nil
	}

	if s, ok := data.(string); ok {
		return Audience{s}, nil
	}

	return data, nil
}

func (c Claims) Lookup(name string) (string, bool) {
	var value string

	switch name {
	case "iss":
		value = c.Iss
	case "sub":
		value = c.Sub
	case "jti":
		value = c.Jti
	case "aud", "exp", "iat", "nbf":
		return "", false
	default:
		raw, ok := c.Extra[name]
		if !ok {
			return "", false
		}

		s, ok := raw.(string)
		return s, ok
	}

	return value, value != ""
}

func (c Claims) Names() []string {
	var names []string
	add := func(present bool, name string) {
		if present {
			names = append(names, name)
		}
	}

	add(c.Iss != "", "iss")
	add(c.Sub != "", "sub")
	add(len(c.Aud) > 0, "aud")
	add(c.Exp != 0, "exp")
	add(c.Iat != 0, "iat")
	add(c.Nbf != 0, "nbf")
	add(c.Jti != "", "jti")
	for name := range c.Extra {
		names = append(names, name)
	}

	sort.Strings(names)
	return names
}

// reservedClaims are validated by the JWT verifier itself and may not be used
// as claim conditions.
var reservedClaims = map[string]struct{}{
	"iss": {},
	"aud": {},
	"exp": {},
	"nbf": {},
	"iat": {},
}

func IsReservedClaim(name string) bool {
	_, ok := reservedClaims[name]
	return ok
}

func ClaimNames(claims any) map[string]struct{} {
	names := map[string]struct{}{}
	collectClaimNames(reflect.TypeOf(claims), names)
	return names
}

func collectClaimNames(t reflect.Type, names map[string]struct{}) {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		name, opts, _ := strings.Cut(field.Tag.Get("json"), ",")

		switch {
		case opts == "squash":
			collectClaimNames(field.Type, names)
		case name != "":
			names[name] = struct{}{}
		}
	}
}

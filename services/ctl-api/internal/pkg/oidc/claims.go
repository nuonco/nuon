package oidc

import (
	"fmt"
	"reflect"
	"sort"

	"github.com/mitchellh/mapstructure"
)

// Audience is the `aud` claim, which tokens carry as either a string or an
// array of strings.
type Audience []string

// Claims are the verified claims of an ID token. Registered claims are typed;
// everything else lands in Extra.
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

// DecodeClaims decodes a raw claim set into Claims.
func DecodeClaims(raw map[string]any) (Claims, error) {
	var claims Claims
	if err := Decode(raw, &claims); err != nil {
		return Claims{}, err
	}

	return claims, nil
}

// Decode decodes a raw claim set into out, a pointer to Claims or to a struct
// embedding it with `json:",squash"`.
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

// Lookup returns the string value of a claim. Claims that are absent, empty,
// or not strings are reported as absent.
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

// Names returns the sorted names of the claims present in the token.
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

package workloadjwt

import (
	"context"
	"strings"
	"testing"
)

func TestVerify_RequiresIssuerAndAudience(t *testing.T) {
	// An empty issuer or audience would make the validator accept anything the caller
	// forgot to constrain, so they are refused rather than defaulted.
	for _, tc := range []struct {
		name string
		req  Request
	}{
		{"no token", Request{Issuer: "https://sts.windows.net/x/", Audience: "aud"}},
		{"no issuer", Request{Token: "a.b.c", Audience: "aud"}},
		{"no audience", Request{Token: "a.b.c", Issuer: "https://sts.windows.net/x/"}},
		{"blank issuer", Request{Token: "a.b.c", Issuer: "   ", Audience: "aud"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := NewVerifier().Verify(context.Background(), tc.req); err == nil {
				t.Error("expected an error")
			}
		})
	}
}

func TestVerify_RejectsNonHTTPSIssuer(t *testing.T) {
	// A non-https issuer would let a caller point key discovery at plaintext, or at
	// something that is not a URL at all.
	for _, issuer := range []string{
		"http://sts.windows.net/tenant/",
		"file:///etc/passwd",
		"sts.windows.net/tenant/",
		"https://",
		"://nope",
	} {
		_, err := NewVerifier().Verify(context.Background(), Request{Token: "a.b.c", Issuer: issuer, Audience: "aud"})
		if err == nil || !strings.Contains(err.Error(), "issuer") {
			t.Errorf("expected issuer %q to be rejected, got %v", issuer, err)
		}
	}
}

func TestUnverifiedClaims(t *testing.T) {
	// {"tid":"abc","oid":"def"} base64url encoded, with throwaway header and signature.
	token := "eyJhbGciOiJSUzI1NiJ9.eyJ0aWQiOiJhYmMiLCJvaWQiOiJkZWYifQ.sig"

	claims, err := UnverifiedClaims(token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got, _ := StringClaim(claims, "tid"); got != "abc" {
		t.Errorf("got tid %q", got)
	}
	if got, _ := StringClaim(claims, "oid"); got != "def" {
		t.Errorf("got oid %q", got)
	}
}

func TestUnverifiedClaims_RejectsMalformed(t *testing.T) {
	for _, token := range []string{
		"",
		"not-a-jwt",
		"only.two",
		"a.b.c.d",
		"a.!!!notbase64!!!.c",
		// Flattened JSON JWS: go-jose would verify protected/payload/signature while
		// a naive Split on '.' can treat a dotted unknown property as the payload.
		`{"protected":"eyJhbGciOiJSUzI1NiJ9","payload":"eyJ0aWQiOiJ0YXJnZXQifQ","signature":"sig"}`,
		`{"protected":"eyJhbGciOiJSUzI1NiJ9","payload":"eyJ0aWQiOiJ0YXJnZXQifQ","signature":"sig","extra":".eyJ0aWQiOiJhdHRhY2tlciJ9."}`,
	} {
		if _, err := UnverifiedClaims(token); err == nil {
			t.Errorf("expected %q to be rejected", token)
		}
	}
}

func TestRequireCompactJWT(t *testing.T) {
	if err := RequireCompactJWT("eyJhbGciOiJSUzI1NiJ9.eyJ0aWQiOiJhYmMifQ.sig"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := RequireCompactJWT(`{"payload":"x"}`); err == nil {
		t.Fatal("expected JSON serialization to be rejected")
	}
}

func TestStringClaim(t *testing.T) {
	claims := map[string]any{
		"str":   "value",
		"num":   12345,
		"empty": "",
		"null":  nil,
	}

	if got, ok := StringClaim(claims, "str"); !ok || got != "value" {
		t.Errorf("got %q, %v", got, ok)
	}
	// A non-string, empty, or absent claim is treated as absent rather than coerced.
	for _, name := range []string{"num", "empty", "null", "missing"} {
		if _, ok := StringClaim(claims, name); ok {
			t.Errorf("expected claim %q to read as absent", name)
		}
	}
}

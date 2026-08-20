package workloadjwt

import (
	"context"
	"fmt"
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

func TestProvider_RejectsNonHTTPSIssuer(t *testing.T) {
	// A non-https issuer would let a caller point key discovery at plaintext, or at
	// something that is not a URL at all.
	v := NewVerifier()

	for _, issuer := range []string{
		"http://sts.windows.net/tenant/",
		"file:///etc/passwd",
		"sts.windows.net/tenant/",
		"https://",
		"://nope",
	} {
		if _, err := v.provider(issuer); err == nil {
			t.Errorf("expected issuer %q to be rejected", issuer)
		}
	}

	if len(v.providers) != 0 {
		t.Errorf("a rejected issuer must not be cached, got %d entries", len(v.providers))
	}
}

func TestProvider_CachesPerIssuer(t *testing.T) {
	v := NewVerifier()
	const issuer = "https://sts.windows.net/11111111-2222-3333-4444-555555555555/"

	first, err := v.provider(issuer)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	second, err := v.provider(issuer)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if first != second {
		t.Error("expected the same provider to be reused for one issuer")
	}
	if len(v.providers) != 1 {
		t.Errorf("expected 1 cached provider, got %d", len(v.providers))
	}
}

// Issuers come from stored state today, but an unbounded cache keyed on anything derived
// from a request is a memory leak waiting to happen.
func TestProvider_CacheIsBounded(t *testing.T) {
	v := NewVerifier()

	for i := range maxCachedIssuers + 50 {
		issuer := fmt.Sprintf("https://sts.windows.net/tenant-%d/", i)
		if _, err := v.provider(issuer); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	if len(v.providers) > maxCachedIssuers {
		t.Errorf("cache grew to %d, above the %d ceiling", len(v.providers), maxCachedIssuers)
	}
	if len(v.inserted) != len(v.providers) {
		t.Errorf("insertion order (%d) drifted from the cache (%d), so eviction leaks",
			len(v.inserted), len(v.providers))
	}

	// The oldest entries are the ones evicted.
	if _, ok := v.providers["https://sts.windows.net/tenant-0/"]; ok {
		t.Error("expected the oldest issuer to have been evicted")
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
	} {
		if _, err := UnverifiedClaims(token); err == nil {
			t.Errorf("expected %q to be rejected", token)
		}
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

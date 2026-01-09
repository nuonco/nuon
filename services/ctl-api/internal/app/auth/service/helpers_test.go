package service

import (
	"regexp"
	"testing"

	"github.com/nuonco/nuon/services/ctl-api/internal"
)

func TestGenerateStateNonce(t *testing.T) {
	// Generate a state nonce
	state, err := generateStateNonce()
	if err != nil {
		t.Fatalf("generateStateNonce() returned error: %v", err)
	}

	// Test length - should be reasonable length after base64 encoding and stripping
	// 32 bytes -> ~43 base64 chars, after stripping non-alphanumeric should still be substantial
	minLength := 20
	if len(state) < minLength {
		t.Errorf("generateStateNonce() returned state with length %d, want at least %d", len(state), minLength)
	}

	// Test that there are no non-alphanumeric characters
	nonAlphaNum := regexp.MustCompile("[^a-zA-Z0-9]")
	if nonAlphaNum.MatchString(state) {
		t.Errorf("generateStateNonce() returned state with non-alphanumeric characters: %q", state)
	}

	// Test uniqueness - generate another and ensure they're different
	state2, err := generateStateNonce()
	if err != nil {
		t.Fatalf("generateStateNonce() second call returned error: %v", err)
	}
	if state == state2 {
		t.Error("generateStateNonce() returned identical states on consecutive calls")
	}
}

func TestIsURLDomainAllowed(t *testing.T) {
	tests := []struct {
		name       string
		rootDomain string
		host       string
		want       bool
	}{
		// Production domain: sub.org.com
		{
			name:       "exact match",
			rootDomain: "sub.org.com",
			host:       "sub.org.com",
			want:       true,
		},
		{
			name:       "subdomain allowed",
			rootDomain: "sub.org.com",
			host:       "dashboard.sub.org.com",
			want:       true,
		},
		{
			name:       "nested subdomain allowed",
			rootDomain: "sub.org.com",
			host:       "api.dashboard.sub.org.com",
			want:       true,
		},
		{
			name:       "case insensitive match",
			rootDomain: "sub.org.com",
			host:       "DASHBOARD.SUB.ORG.COM",
			want:       true,
		},
		{
			name:       "with port allowed",
			rootDomain: "sub.org.com",
			host:       "dashboard.sub.org.com:8080",
			want:       true,
		},
		{
			name:       "different domain rejected",
			rootDomain: "sub.org.com",
			host:       "evil.com",
			want:       false,
		},
		{
			name:       "suffix attack rejected - no dot prefix",
			rootDomain: "sub.org.com",
			host:       "evilsub.org.com",
			want:       false,
		},
		{
			name:       "suffix attack rejected - domain as subdomain of attacker",
			rootDomain: "sub.org.com",
			host:       "sub.org.com.evil.com",
			want:       false,
		},
		{
			name:       "partial match rejected",
			rootDomain: "sub.org.com",
			host:       "org.com",
			want:       false,
		},
		// Localhost development
		{
			name:       "localhost allowed when root is localhost",
			rootDomain: "localhost",
			host:       "localhost",
			want:       true,
		},
		{
			name:       "localhost with port allowed",
			rootDomain: "localhost",
			host:       "localhost:3000",
			want:       true,
		},
		{
			name:       "127.0.0.1 allowed when root is localhost",
			rootDomain: "localhost",
			host:       "127.0.0.1",
			want:       true,
		},
		{
			name:       "127.0.0.1 with port allowed",
			rootDomain: "localhost",
			host:       "127.0.0.1:8080",
			want:       true,
		},
		{
			name:       "external domain rejected when root is localhost",
			rootDomain: "localhost",
			host:       "evil.com",
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &service{
				cfg: &internal.Config{
					RootDomain: tt.rootDomain,
				},
			}

			got := s.isURLDomainAllowed(tt.host)
			if got != tt.want {
				t.Errorf("isURLDomainAllowed(%q) with RootDomain=%q = %v, want %v",
					tt.host, tt.rootDomain, got, tt.want)
			}
		})
	}
}

package github

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSubjectRoundTrip(t *testing.T) {
	tests := map[string]struct {
		subject Subject
		want    string
	}{
		"mutable with ref": {
			subject: NewSubject("acme", "app", "ref:refs/heads/main"),
			want:    "repo:acme/app:ref:refs/heads/main",
		},
		"mutable without context": {
			subject: NewSubject("acme", "app", ""),
			want:    "repo:acme/app",
		},
		"immutable ids": {
			subject: NewSubject("acme", "app", "environment:production").WithIDs("123", "456"),
			want:    "repo:acme@123/app@456:environment:production",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.subject.String())

			parsed, ok := ParseSubject(tt.want)
			require.True(t, ok)
			require.Equal(t, tt.subject, parsed)
		})
	}
}

func TestParseSubjectOwner(t *testing.T) {
	tests := map[string]struct {
		pattern string
		want    string
		wantOK  bool
	}{
		"mutable":             {"repo:acme/app:*", "acme", true},
		"immutable":           {"repo:acme@123/app@456:*", "acme", true},
		"wildcard repo":       {"repo:acme/*:ref:refs/heads/main", "acme", true},
		"wildcard owner":      {"repo:*/app:**", "", false},
		"partial wildcard":    {"repo:acme*/app", "", false},
		"brace owner":         {"repo:{acme,evil}/app", "", false},
		"empty owner":         {"repo:/app", "", false},
		"colon in owner":      {"repo:acme:x/app", "", false},
		"no slash":            {"repo:acme", "", false},
		"not a repo subject":  {"environment:prod", "", false},
		"empty owner with id": {"repo:@1/app", "", false},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, ok := ParseSubjectOwner(tt.pattern)
			require.Equal(t, tt.wantOK, ok)
			require.Equal(t, tt.want, got)
		})
	}
}

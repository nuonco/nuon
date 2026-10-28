package service

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseTokenDuration(t *testing.T) {
	testCases := []struct {
		name    string
		raw     string
		want    time.Duration
		wantErr bool
	}{
		{name: "empty defaults to one year", raw: "", want: 8760 * time.Hour},
		{name: "valid hours", raw: "720h", want: 720 * time.Hour},
		{name: "valid short", raw: "30m", want: 30 * time.Minute},
		{name: "zero is rejected", raw: "0h", wantErr: true},
		{name: "negative is rejected", raw: "-1h", wantErr: true},
		{name: "unparseable is rejected", raw: "bogus", wantErr: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseTokenDuration(tc.raw)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

// The guard on deleteTokenServiceAccount. Static tokens created with their own
// dedicated service account take it with them when revoked; tokens issued against
// an account that outlives them — an install stack's, a runner's — must not. Getting
// this wrong deletes a live identity out from under running infrastructure.
func TestDedicatedTokenSubjectPrefix(t *testing.T) {
	const orgID = "org_abc"
	prefix := dedicatedTokenSubjectPrefix(orgID)

	testCases := []struct {
		name          string
		subject       string
		wantDedicated bool
	}{
		{
			name:          "account created for one token",
			subject:       prefix + "acc123",
			wantDedicated: true,
		},
		{
			name:          "install stack account",
			subject:       "istmj4srnucyq26cwdj3nceerb",
			wantDedicated: false,
		},
		{
			name:          "another org's dedicated account",
			subject:       dedicatedTokenSubjectPrefix("org_other") + "acc123",
			wantDedicated: false,
		},
		{
			name:          "empty subject",
			subject:       "",
			wantDedicated: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.wantDedicated, strings.HasPrefix(tc.subject, prefix))
		})
	}
}

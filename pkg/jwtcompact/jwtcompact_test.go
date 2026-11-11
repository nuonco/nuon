package jwtcompact

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRequire(t *testing.T) {
	t.Parallel()

	ok := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"none"}`)) + "." +
		base64.RawURLEncoding.EncodeToString([]byte(`{"iss":"https://example.com"}`)) + "." +
		base64.RawURLEncoding.EncodeToString([]byte("sig"))
	require.NoError(t, Require(ok))
	require.NoError(t, Require("  "+ok+"  "))

	cases := []string{
		"",
		"not-a-jwt",
		`{"payload":"x","signatures":[{"protected":"y","signature":"z"}]}`,
		ok + ".extra",
		"a.b",
		"a.b.c.d",
		"a.b!c.d",
		"a.b c.d",
	}
	for _, tc := range cases {
		require.Error(t, Require(tc), tc)
	}
}

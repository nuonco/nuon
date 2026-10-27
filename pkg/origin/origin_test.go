package origin

import "testing"

func TestIsHTTPLoopback(t *testing.T) {
	cases := []struct {
		raw  string
		want bool
	}{
		{raw: "http://localhost:4001", want: true},
		{raw: "http://127.0.0.1:4010", want: true},
		{raw: "http://[::1]:4000", want: true},
		{raw: "https://localhost:4000"},
		{raw: "http://localhost.evil.com"},
		{raw: "http://127.0.0.1.nip.io"},
		{raw: ""},
		{raw: "not a url"},
		{raw: "http://app.example.com"},
	}
	for _, tc := range cases {
		if got := IsHTTPLoopback(tc.raw); got != tc.want {
			t.Errorf("IsHTTPLoopback(%q) = %v, want %v", tc.raw, got, tc.want)
		}
	}
}

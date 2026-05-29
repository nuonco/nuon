package client

import "testing"

func TestResolveSiteURL(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		// Known regions.
		{"us1", "https://api.datadoghq.com"},
		{"us3", "https://api.us3.datadoghq.com"},
		{"us5", "https://api.us5.datadoghq.com"},
		{"eu1", "https://api.datadoghq.eu"},
		{"ap1", "https://api.ap1.datadoghq.com"},
		{"gov", "https://api.ddog-gov.com"},

		// Custom URLs pass through verbatim (with trailing slash trimmed).
		{"https://dd.internal.example.com", "https://dd.internal.example.com"},
		{"https://dd.internal.example.com/", "https://dd.internal.example.com"},

		// Whitespace tolerated.
		{"  us1  ", "https://api.datadoghq.com"},

		// Unknown / empty → defensive us1 fallback.
		{"", "https://api.datadoghq.com"},
		{"us9", "https://api.datadoghq.com"},
	}
	for _, c := range cases {
		got := ResolveSiteURL(c.in)
		if got != c.want {
			t.Errorf("ResolveSiteURL(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestAppURLForSite(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"us1", "https://app.datadoghq.com"},
		{"us3", "https://app.us3.datadoghq.com"},
		{"us5", "https://app.us5.datadoghq.com"},
		{"eu1", "https://app.datadoghq.eu"},
		{"ap1", "https://app.ap1.datadoghq.com"},
		{"gov", "https://app.ddog-gov.com"},

		// Custom URL → returned verbatim; caller decides UI deep-link.
		{"https://dd.internal.example.com/", "https://dd.internal.example.com"},
	}
	for _, c := range cases {
		got := AppURLForSite(c.in)
		if got != c.want {
			t.Errorf("AppURLForSite(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

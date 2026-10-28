package internal

import "testing"

func TestServerAPIURL(t *testing.T) {
	c := &Config{APIUrl: "https://api.example.com"}
	if got := c.ServerAPIURL(); got != c.APIUrl {
		t.Fatalf("ServerAPIURL() = %q, want the public url", got)
	}

	c.InternalAPIUrl = "http://ctl-api-public.ctl-api.svc.cluster.local:8081"
	if got := c.ServerAPIURL(); got != c.InternalAPIUrl {
		t.Fatalf("ServerAPIURL() = %q, want the internal url", got)
	}
}

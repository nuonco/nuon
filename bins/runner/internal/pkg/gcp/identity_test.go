package gcp

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestGetAccessToken_RequestsComputeReadonlyScope(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.RequestURI()
		w.Header().Set("Metadata-Flavor", "Google")
		w.Write([]byte(`{"access_token":"fake-token","expires_in":3600,"token_type":"Bearer"}`))
	}))
	defer srv.Close()

	t.Setenv("GCE_METADATA_HOST", srv.Listener.Addr().String())

	if _, err := GetAccessToken(t.Context()); err != nil {
		t.Fatalf("GetAccessToken() error = %v", err)
	}

	parsed, err := url.Parse(gotPath)
	if err != nil {
		t.Fatalf("unable to parse requested path %q: %v", gotPath, err)
	}
	const wantScope = "https://www.googleapis.com/auth/compute.readonly"
	if got := parsed.Query().Get("scopes"); got != wantScope {
		t.Fatalf("scopes query param = %q, want %q", got, wantScope)
	}
}

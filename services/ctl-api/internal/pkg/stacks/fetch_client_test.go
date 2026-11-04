package stacks

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTemplateFetchClient_RefusesPlainHTTP(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("request must not reach the server")
	}))
	defer srv.Close()

	_, err := NewTemplateFetchClient().Get(srv.URL)
	require.ErrorContains(t, err, "must use https")
}

func TestTemplateFetchClient_RefusesNonPublicHosts(t *testing.T) {
	client := NewTemplateFetchClient()
	for _, u := range []string{
		"https://127.0.0.1/template.json",
		"https://169.254.169.254/latest/meta-data/",
		"https://10.0.0.5/template.json",
		"https://[fd00:ec2::254]/template.json",
		"https://localhost/template.json",
	} {
		_, err := client.Get(u)
		require.Error(t, err, u)
		require.ErrorContains(t, err, "non-public", u)
	}
}

func TestTemplateFetchClient_RefusesCredentialsAndOtherSchemes(t *testing.T) {
	client := NewTemplateFetchClient()

	_, err := client.Get("https://user:pass@example.com/template.json")
	require.ErrorContains(t, err, "without credentials")

	_, err = client.Get("ftp://example.com/template.json")
	require.ErrorContains(t, err, "unsupported URL scheme")
}

func TestTemplateFetchClient_RefusesRedirectOffHTTPS(t *testing.T) {
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("redirect target must not be fetched")
	}))
	defer target.Close()

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target.URL, http.StatusFound)
	}))
	defer srv.Close()

	// An insecure client is needed to reach the TLS test server at all; the
	// redirect must still be refused when the client is strict about scheme.
	client := NewTemplateFetchClient()
	client.Transport = templateTransport{base: srv.Client().Transport, allowInsecure: false}

	_, err := client.Get(srv.URL)
	require.ErrorContains(t, err, "must use https")
}

func TestTemplateFetchClient_InsecureAllowsLocalhost(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"ok":true}`)
	}))
	defer srv.Close()

	resp, err := NewLocalTemplateFetchClient().Get(srv.URL)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, `{"ok":true}`, string(body))
}

func TestTemplateFetchClient_CapsBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Stream without a Content-Length so the cap is enforced on the read path.
		w.Header().Set("Transfer-Encoding", "chunked")
		chunk := strings.Repeat("x", 64<<10)
		for written := 0; written <= MaxTemplateFetchBytes; written += len(chunk) {
			if _, err := io.WriteString(w, chunk); err != nil {
				return
			}
		}
	}))
	defer srv.Close()

	resp, err := NewLocalTemplateFetchClient().Get(srv.URL)
	require.NoError(t, err)
	defer resp.Body.Close()

	_, err = io.ReadAll(resp.Body)
	require.ErrorContains(t, err, "exceeds")
}

func TestTemplateFetchClient_CapsBodyByContentLength(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "99999999999")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	_, err := NewLocalTemplateFetchClient().Get(srv.URL)
	require.ErrorContains(t, err, "exceeds")
}

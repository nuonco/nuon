package oidcverify

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPublicIP(t *testing.T) {
	cases := map[string]bool{
		"8.8.8.8":         true,
		"1.1.1.1":         true,
		"127.0.0.1":       false,
		"10.1.2.3":        false,
		"172.16.5.5":      false,
		"192.168.1.1":     false,
		"169.254.169.254": false, // link local, cloud metadata
		"100.64.1.1":      false, // carrier-grade NAT
		"192.0.0.8":       false, // IETF protocol assignments
		"198.18.0.1":      false, // benchmarking
		"0.0.0.0":         false,
		"224.0.0.1":       false, // multicast
		"::1":             false,
		"fe80::1":         false,
		"2001:db8::1":     false, // documentation
		"2606:4700::1":    true,
	}

	for raw, want := range cases {
		require.Equal(t, want, publicIP(net.ParseIP(raw)), raw)
	}
}

// localHTTPSServer serves TLS on 127.0.0.1 with a certificate the returned
// client trusts, so tests exercise the real dial path.
func localHTTPSServer(t *testing.T, handler http.Handler) (*httptest.Server, *http.Client) {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "oidc-test"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		DNSNames:     []string{"localhost"},
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	require.NoError(t, err)

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyDER, err := x509.MarshalECPrivateKey(key)
	require.NoError(t, err)
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})

	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	require.NoError(t, err)

	pool := x509.NewCertPool()
	require.True(t, pool.AppendCertsFromPEM(certPEM))

	server := httptest.NewUnstartedServer(handler)
	server.TLS = &tls.Config{Certificates: []tls.Certificate{cert}}
	server.StartTLS()
	t.Cleanup(server.Close)

	trusted := server.Client()
	trusted.Transport.(*http.Transport).TLSClientConfig = &tls.Config{RootCAs: pool}

	return server, trusted
}

func TestHTTPClientBlocksLocalIssuers(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"ok": "yes"})
	})

	t.Run("secure client refuses loopback", func(t *testing.T) {
		server, trusted := localHTTPSServer(t, handler)

		client := newHTTPClient(false)
		base := client.Transport.(limitedTransport).base.(*http.Transport)
		base.TLSClientConfig = trusted.Transport.(*http.Transport).TLSClientConfig

		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL, nil)
		require.NoError(t, err)

		_, err = client.Do(req)
		require.ErrorContains(t, err, "non-public address")
	})

	t.Run("insecure client reaches loopback", func(t *testing.T) {
		server, trusted := localHTTPSServer(t, handler)

		client := newHTTPClient(true)
		base := client.Transport.(limitedTransport).base.(*http.Transport)
		base.TLSClientConfig = trusted.Transport.(*http.Transport).TLSClientConfig

		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL, nil)
		require.NoError(t, err)

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("http is refused", func(t *testing.T) {
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://example.com", nil)
		require.NoError(t, err)

		_, err = newHTTPClient(false).Do(req)
		require.ErrorContains(t, err, "require https")
	})

	t.Run("too many redirects are refused", func(t *testing.T) {
		var server *httptest.Server
		server, _ = localHTTPSServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, server.URL+"/jwks", http.StatusFound)
		}))

		client := newHTTPClient(true)
		base := client.Transport.(limitedTransport).base.(*http.Transport)
		base.TLSClientConfig = server.Client().Transport.(*http.Transport).TLSClientConfig

		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL, nil)
		require.NoError(t, err)

		_, err = client.Do(req)
		require.ErrorContains(t, err, "redirect")
	})

	t.Run("oversized response is refused", func(t *testing.T) {
		server, trusted := localHTTPSServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Length", "10485761")
			w.WriteHeader(http.StatusOK)
		}))

		client := newHTTPClient(true)
		base := client.Transport.(limitedTransport).base.(*http.Transport)
		base.TLSClientConfig = trusted.Transport.(*http.Transport).TLSClientConfig

		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL, nil)
		require.NoError(t, err)

		_, err = client.Do(req)
		require.ErrorContains(t, err, "too large")
	})
}

package publicnet

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/net/dns/dnsmessage"
)

func TestIsPublicIP(t *testing.T) {
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

		"100.100.100.200":        false, // alibaba metadata
		"fd00:ec2::254":          false, // aws metadata
		"fd20:ce::254":           false, // gcp metadata
		"::ffff:169.254.169.254": false, // ipv4-mapped
		"64:ff9b::a9fe:a9fe":     false, // nat64
		"2002:a9fe:a9fe::1":      false, // 6to4
		"2001:0:a9fe:a9fe::1":    false, // teredo
		"0.1.2.3":                false,
		"240.0.0.1":              false,
		"255.255.255.255":        false,
		"203.0.113.5":            false, // documentation
		"2a00:1450:4001::1":      true,
	}

	for raw, want := range cases {
		require.Equal(t, want, IsPublicIP(net.ParseIP(raw)), raw)
	}
}

func TestHTTPClientRejectsUnsafeRequests(t *testing.T) {
	var requests atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		_, _ = io.WriteString(w, "must not be fetched")
	}))
	t.Cleanup(server.Close)
	t.Setenv("HTTP_PROXY", server.URL)
	t.Setenv("HTTPS_PROXY", server.URL)
	client := NewHTTPClient()
	t.Cleanup(client.CloseIdleConnections)
	for _, target := range []string{
		server.URL, "http://10.0.0.1/", "http://169.254.169.254/", "http://100.100.100.200/",
		"http://[::1]/", "http://[::ffff:169.254.169.254]/", "http://[64:ff9b::a9fe:a9fe]/",
		"http://[2002:a9fe:a9fe::1]/", "http://[2001:0:a9fe:a9fe::1]/",
		"http://user:password@8.8.8.8/", "file:///outside-canary", "http:opaque",
	} {
		t.Run(target, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
			require.NoError(t, err)
			response, err := client.Do(req)
			if response != nil {
				response.Body.Close()
			}
			require.Error(t, err)
		})
	}
	require.Zero(t, requests.Load())
}

type redirectTransport func(*http.Request) (*http.Response, error)

func (f redirectTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestHTTPClientRedirectPolicy(t *testing.T) {
	for name, tc := range map[string]struct {
		target    string
		wantError bool
	}{
		"public":      {target: "https://1.1.1.1/final"},
		"private":     {target: "http://10.0.0.1/final", wantError: true},
		"metadata":    {target: "http://169.254.169.254/final", wantError: true},
		"credentials": {target: "https://user:password@1.1.1.1/final", wantError: true},
		"file":        {target: "file:///outside-canary", wantError: true},
		"loop":        {target: "http://8.8.8.8/start?secret=canary", wantError: true},
	} {
		t.Run(name, func(t *testing.T) {
			client := NewHTTPClient()
			t.Cleanup(client.CloseIdleConnections)
			requests := 0
			client.Transport = redirectTransport(func(req *http.Request) (*http.Response, error) {
				requests++
				if requests > 1 {
					require.Empty(t, req.Header.Get("Referer"))
				}
				status, headers := http.StatusOK, make(http.Header)
				if req.URL.Path == "/start" {
					status = http.StatusFound
					headers.Set("Location", tc.target)
				}
				return &http.Response{StatusCode: status, Header: headers, Body: io.NopCloser(strings.NewReader("public contents")), Request: req}, nil
			})
			response, err := client.Get("http://8.8.8.8/start?secret=canary")
			if response != nil {
				defer response.Body.Close()
			}
			if tc.wantError {
				require.Error(t, err)
				if name == "loop" {
					require.Equal(t, 10, requests)
				} else {
					require.Equal(t, 1, requests)
				}
			} else {
				require.NoError(t, err)
				contents, err := io.ReadAll(response.Body)
				require.NoError(t, err)
				require.Equal(t, "public contents", string(contents))
				require.Equal(t, 2, requests)
			}
		})
	}
}

func TestPublicDNSAndPinnedDial(t *testing.T) {
	dns, err := net.ListenPacket("udp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { dns.Close() })
	var mixed atomic.Bool
	go func() {
		buf := make([]byte, 4096)
		for {
			n, peer, err := dns.ReadFrom(buf)
			if err != nil {
				return
			}
			var message dnsmessage.Message
			if err := message.Unpack(buf[:n]); err != nil {
				continue
			}
			message.Response, message.Authoritative, message.RecursionAvailable = true, true, true
			for _, q := range message.Questions {
				if q.Type == dnsmessage.TypeA {
					header := dnsmessage.ResourceHeader{Name: q.Name, Type: dnsmessage.TypeA, Class: dnsmessage.ClassINET}
					message.Answers = append(message.Answers, dnsmessage.Resource{Header: header, Body: &dnsmessage.AResource{A: [4]byte{8, 8, 8, 8}}})
					if mixed.Load() {
						message.Answers = append(message.Answers, dnsmessage.Resource{Header: header, Body: &dnsmessage.AResource{A: [4]byte{127, 0, 0, 1}}})
					}
				}
			}
			if response, err := message.Pack(); err == nil {
				_, _ = dns.WriteTo(response, peer)
			}
		}
	}()
	previous := net.DefaultResolver
	net.DefaultResolver = &net.Resolver{PreferGo: true, Dial: func(ctx context.Context, _, _ string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, "udp", dns.LocalAddr().String())
	}}
	t.Cleanup(func() { net.DefaultResolver = previous })
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	require.NoError(t, RequirePublicHost(ctx, "source.example.test"))
	var dialed string
	dial := DialContext(&net.Dialer{ControlContext: func(_ context.Context, _, address string, _ syscall.RawConn) error {
		dialed = address
		return fmt.Errorf("stopped before connecting")
	}})
	_, err = dial(ctx, "tcp", "source.example.test:80")
	require.ErrorContains(t, err, "stopped before connecting")
	require.Equal(t, "8.8.8.8:80", dialed)
	mixed.Store(true)
	dialed = ""
	_, err = dial(ctx, "tcp", "source.example.test:80")
	require.ErrorContains(t, err, "non-public address")
	require.Empty(t, dialed)
}

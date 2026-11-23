package registryclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
	"oras.land/oras-go/v2/registry/remote/retry"

	"github.com/nuonco/nuon/pkg/publicnet"
)

func TestForUsesDefaultClientWithoutTheMark(t *testing.T) {
	require.Same(t, retry.DefaultClient, For(context.Background()))
}

func TestForRefusesNonPublicAddressesWhenMarked(t *testing.T) {
	var requests atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
	}))
	defer server.Close()

	ctx := publicnet.WithPublicOnly(context.Background())
	client := For(ctx)
	require.NotSame(t, retry.DefaultClient, client)

	for _, target := range []string{
		server.URL,
		"http://169.254.169.254/latest/meta-data/",
		"http://168.63.129.16/metadata",
		"http://10.0.0.1/v2/",
	} {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
		require.NoError(t, err)
		_, err = client.Do(req)
		require.ErrorIs(t, err, publicnet.ErrNonPublic, target)
	}
	require.Zero(t, requests.Load())
}

func TestForRefusesRedirectToNonPublicAddress(t *testing.T) {
	var requests atomic.Int64
	internal := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
	}))
	defer internal.Close()

	// The marked client cannot reach a loopback server to receive a redirect, so
	// the first hop is served by a stub and only the redirect dials for real.
	client := *For(publicnet.WithPublicOnly(context.Background()))
	client.Transport = &redirectFirst{location: internal.URL, next: client.Transport}

	_, err := client.Get("https://registry.example.com/v2/")
	require.ErrorIs(t, err, publicnet.ErrNonPublic)
	require.Zero(t, requests.Load())
}

type redirectFirst struct {
	location string
	next     http.RoundTripper
	served   atomic.Bool
}

func (r *redirectFirst) RoundTrip(req *http.Request) (*http.Response, error) {
	if r.served.CompareAndSwap(false, true) {
		return &http.Response{
			StatusCode: http.StatusTemporaryRedirect,
			Header:     http.Header{"Location": []string{r.location}},
			Body:       http.NoBody,
			Request:    req,
		}, nil
	}
	return r.next.RoundTrip(req)
}

// Package registryclient picks the HTTP client used for OCI registry requests.
package registryclient

import (
	"context"
	"net"
	"net/http"
	"time"

	"oras.land/oras-go/v2/registry/remote/retry"

	"github.com/nuonco/nuon/pkg/publicnet"
)

// publicOnly dials only public addresses. The dial runs for every connection,
// so it also covers redirects and the token realm a registry names in its
// WWW-Authenticate header. Proxy environment variables are ignored: a proxy
// would make the connection on our behalf and skip that check.
var publicOnly = &http.Client{
	Transport: retry.NewTransport(&http.Transport{
		DialContext:           publicnet.DialContext(&net.Dialer{Timeout: 30 * time.Second, KeepAlive: 30 * time.Second}),
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: time.Second,
	}),
}

// For returns the client registry requests made under ctx should use. Registry
// hosts on the control plane come from app config, so a context marked with
// publicnet.WithPublicOnly gets a client that refuses loopback, private,
// link-local and metadata addresses. The runner reaches private registries in
// the customer's network and keeps the default client.
func For(ctx context.Context) *http.Client {
	if publicnet.PublicOnly(ctx) {
		return publicOnly
	}
	return retry.DefaultClient
}

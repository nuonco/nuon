package get

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"

	"github.com/nuonco/nuon/pkg/publicnet"
	"github.com/nuonco/nuon/pkg/sourcefs"
)

type Options struct {
	FieldTimeout time.Duration
	RootDir      string
	// Source is borrowed; when set, it is the authority for local file reads.
	Source      *sourcefs.FS
	OnLocalFile func(path string, contents []byte) error
	// CheckRemote, when set, is called before each remote fetch. field names the
	// config field as <Type>.<Field>. A non-nil error aborts the parse.
	CheckRemote func(field, rawURL string) error
}

type get struct {
	dst        any
	opts       *Options
	httpClient *http.Client
}

const maxRedirects = 10

// newLocalParseClient is the HTTP client for parses with no source filesystem,
// such as the CLI's. The config's author runs it on their own machine, so a URL
// named in the config may be a private host, as with go-getter's default
// client. A redirect is different: the server picks the target. A redirect
// away from the host the config named is followed only to a public address,
// so an allowed or otherwise harmless URL cannot send the fetch to a metadata
// or internal endpoint, and the check attached with
// publicnet.WithRedirectCheck runs on every redirect target.
func newLocalParseClient() *http.Client {
	dialer := &net.Dialer{Timeout: 30 * time.Second, KeepAlive: 30 * time.Second}
	publicDial := publicnet.DialContext(dialer)

	transport := http.DefaultTransport.(*http.Transport).Clone()
	// A pooled connection skips the dial, and the dial is where a redirect
	// target's address is checked.
	transport.DisableKeepAlives = true
	transport.DialContext = func(ctx context.Context, network, address string) (net.Conn, error) {
		if publicnet.PublicOnly(ctx) {
			return publicDial(ctx, network, address)
		}
		return dialer.DialContext(ctx, network, address)
	}

	return &http.Client{
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= maxRedirects {
				return fmt.Errorf("stopped after %d redirects", maxRedirects)
			}
			if err := publicnet.CheckRedirectTarget(req); err != nil {
				return err
			}
			if req.URL.Host == via[0].URL.Host {
				return nil
			}
			proxy, err := transport.Proxy(req)
			if err != nil {
				return err
			}
			if proxy != nil {
				// The dial goes to the proxy, so the target cannot be pinned.
				return publicnet.RequirePublicHost(req.Context(), req.URL.Hostname())
			}
			*req = *req.WithContext(publicnet.WithPublicOnly(req.Context()))
			return nil
		},
	}
}

func Parse(ctx context.Context, obj any, opts *Options) error {
	v := validator.New()
	if err := v.StructCtx(ctx, opts); err != nil {
		return err
	}
	options := *opts
	if options.Source != nil && options.RootDir == "" {
		options.RootDir = options.Source.Root()
	}

	getter := &get{
		dst:  obj,
		opts: &options,
	}
	switch {
	case options.Source != nil:
		getter.httpClient = publicnet.NewHTTPClient()
		defer getter.httpClient.CloseIdleConnections()
	default:
		getter.httpClient = newLocalParseClient()
	}

	if err := getter.GetAll(ctx); err != nil {
		return err
	}

	return nil
}

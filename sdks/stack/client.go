package stack

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"

	genclient "github.com/nuonco/nuon/sdks/stack/client"
	"github.com/nuonco/nuon/sdks/stack/client/operations"
)

const (
	maxAttempts    = 5
	maxRetryDelay  = 8 * time.Second
	initialDelay   = 500 * time.Millisecond
	defaultTimeout = 10 * time.Second
)

//go:generate ./generate.sh

// Client calls the runner API's stacks namespace. The transport and wire types
// are generated from the ctl-api swagger spec; this interface is the stable
// surface embedders and stack-cli consume.
type Client interface {
	// FetchConfig reads the install's rendered stack config.
	FetchConfig(ctx context.Context) (*Config, error)
	// PhoneHome reports stack outputs to the endpoint the config named.
	PhoneHome(ctx context.Context, phoneHomeURL string, payload map[string]any) error
}

type client struct {
	ops       operations.ClientService
	authInfo  runtime.ClientAuthInfoWriter
	installID string
	opts      Options
}

var _ Client = (*client)(nil)

// newDefaultTransport builds an *http.Transport that does not share state with
// http.DefaultTransport, which other code in-process mutates globally.
func newDefaultTransport() *http.Transport {
	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          16,
		MaxIdleConnsPerHost:   8,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}

// newOps builds a generated operations client rooted at rawURL. Callers pass the
// runner API base for reads and the server-supplied phone-home host for reports,
// so a report always lands where the config said it should.
func newOps(rawURL string, hc *http.Client) (operations.ClientService, error) {
	u, err := url.Parse(strings.TrimSuffix(rawURL, "/"))
	if err != nil {
		return nil, fmt.Errorf("parse api url %q: %w", rawURL, err)
	}
	if u.Host == "" {
		return nil, fmt.Errorf("api url %q has no host", rawURL)
	}

	if hc == nil {
		hc = &http.Client{Transport: newDefaultTransport(), Timeout: defaultTimeout}
	}

	schemes := []string{u.Scheme}
	if u.Scheme == "" {
		schemes = []string{"https"}
	}

	tr := httptransport.NewWithClient(u.Host, genclient.DefaultBasePath, schemes, hc)
	return operations.New(tr, strfmt.Default), nil
}

func newClient(ctx context.Context, opts Options) (*client, error) {
	if err := opts.validate(); err != nil {
		return nil, err
	}

	token, err := resolveToken(ctx, opts)
	if err != nil {
		return nil, err
	}

	ops, err := newOps(opts.APIURL, opts.HTTPClient)
	if err != nil {
		return nil, err
	}

	return &client{
		ops:       ops,
		authInfo:  bearerAuth(token),
		installID: opts.InstallID,
		opts:      opts,
	}, nil
}

func (c *client) FetchConfig(ctx context.Context) (*Config, error) {
	params := operations.NewGetStackConfigParamsWithContext(ctx).
		WithInstallID(c.installID)

	var cfg *Config
	err := retry(ctx, func() error {
		res, err := c.ops.GetStackConfig(params, c.authInfo)
		if err != nil {
			return err
		}
		if res.Payload == nil || res.Payload.Config == nil {
			return errNoConfig
		}
		cfg, err = configFromModel(res.Payload.Config)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("fetch stack config: %w", err)
	}

	return cfg, nil
}

// errNoConfig is terminal: a 200 with no config block will not change on retry.
var errNoConfig = errors.New("runner api returned no config block")

// retry makes up to maxAttempts attempts with capped exponential backoff.
// Transport errors and 5xx are retried; 4xx is returned immediately, since a
// rejected credential is rejected identically every time.
func retry(ctx context.Context, fn func() error) error {
	var lastErr error
	delay := initialDelay

	for attempt := 0; attempt < maxAttempts; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
			if delay *= 2; delay > maxRetryDelay {
				delay = maxRetryDelay
			}
		}

		err := fn()
		if err == nil {
			return nil
		}
		if !isRetryable(err) {
			return err
		}
		lastErr = err
	}

	return fmt.Errorf("gave up after %d attempts: %w", maxAttempts, lastErr)
}

// isRetryable reports whether err is worth another attempt. Generated response
// errors carry their status code; transport errors carry none and are retried.
func isRetryable(err error) bool {
	if errors.Is(err, errNoConfig) {
		return false
	}

	// Generated response types expose Code(); the runtime's own error carries it
	// as a field.
	var coded interface{ Code() int }
	if errors.As(err, &coded) {
		return !isClientError(coded.Code())
	}

	var apiErr *runtime.APIError
	if errors.As(err, &apiErr) {
		return !isClientError(apiErr.Code)
	}

	return true
}

func isClientError(code int) bool {
	return code >= 400 && code < 500
}

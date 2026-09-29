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
	"github.com/nuonco/nuon/sdks/stack/models"
)

const (
	maxAttempts    = 5
	maxRetryDelay  = 8 * time.Second
	initialDelay   = 500 * time.Millisecond
	defaultTimeout = 10 * time.Second
)

//go:generate ./generate.sh

type Client interface {
	FetchConfig(ctx context.Context) (*models.AppInstallerSDKConfig, error)
	PhoneHome(ctx context.Context, phoneHomeURL string, payload map[string]any) error
}

type client struct {
	ops       operations.ClientService
	authInfo  runtime.ClientAuthInfoWriter
	installID string
	opts      Options
}

var _ Client = (*client)(nil)

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

	basePath := u.Path
	if basePath == "" {
		basePath = genclient.DefaultBasePath
	}

	tr := httptransport.NewWithClient(u.Host, basePath, schemes, hc)
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

func (c *client) FetchConfig(ctx context.Context) (*models.AppInstallerSDKConfig, error) {
	params := operations.NewGetStackConfigParamsWithContext(ctx).
		WithInstallID(c.installID)

	var cfg *models.AppInstallerSDKConfig
	err := retry(ctx, func() error {
		res, err := c.ops.GetStackConfig(params, c.authInfo)
		if err != nil {
			return err
		}
		if res.Payload == nil || res.Payload.Config == nil {
			return errNoConfig
		}
		cfg = res.Payload.Config
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("fetch stack config: %w", err)
	}

	return cfg, nil
}

var errNoConfig = errors.New("runner api returned no config block")

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

func isRetryable(err error) bool {
	if errors.Is(err, errNoConfig) {
		return false
	}

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

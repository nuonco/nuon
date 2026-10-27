package oidcverify

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/nuonco/nuon/pkg/publicnet"
)

// ValidateIssuerURL checks that raw is safe to store as a trust policy issuer:
// an https URL with a host and no credentials, query, or fragment, that does
// not address a private or internal network. allowInsecure permits http and
// local/internal hosts, for tests and local development.
func ValidateIssuerURL(ctx context.Context, raw string, allowInsecure bool) error {
	parsed, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("invalid issuer URL: %w", err)
	}

	if parsed.Hostname() == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return fmt.Errorf("issuer URL must include a host and no credentials, query, or fragment")
	}

	switch parsed.Scheme {
	case "https":
	case "http":
		if !allowInsecure {
			return fmt.Errorf("issuer URL must use https")
		}
	default:
		return fmt.Errorf("issuer URL must use https")
	}

	if !allowInsecure {
		if strings.EqualFold(parsed.Hostname(), "localhost") {
			return fmt.Errorf("issuer URL must not address a local network")
		}
		if err := publicnet.RequirePublicHost(ctx, parsed.Hostname()); err != nil {
			return fmt.Errorf("issuer URL must not address a local network: %w", err)
		}
	}

	return nil
}

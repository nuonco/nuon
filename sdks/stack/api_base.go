package stack

import (
	"fmt"
	"net"
	"net/url"
	"strings"
)

// APIBase is the origin the stack client calls. Callers still pass the runner
// API (https://runner.stage.nuon.co). The stack API is that same host with the
// first label replaced, and the route is the generated /v1/stacks/... path, so
// anything after the origin is dropped.
func APIBase(raw string) (string, error) {
	u, err := url.Parse(strings.TrimRight(strings.TrimSpace(raw), "/"))
	if err != nil {
		return "", fmt.Errorf("parse api url %q: %w", raw, err)
	}
	if u.Hostname() == "" {
		return "", fmt.Errorf("api url %q has no host", raw)
	}

	host := u.Hostname()
	if label, rest, ok := strings.Cut(host, "."); ok && label == "runner" && rest != "" {
		host = "stack." + rest
	}
	if port := u.Port(); port != "" {
		u.Host = net.JoinHostPort(host, port)
	} else {
		u.Host = host
	}
	u.Path = ""
	u.RawPath = ""
	u.RawQuery = ""
	u.Fragment = ""
	if u.Scheme == "" {
		u.Scheme = "https"
	}
	return u.String(), nil
}

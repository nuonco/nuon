package publicnet

import (
	"context"
	"net/http"
	"net/url"
)

type redirectCheckKey struct{}

// WithRedirectCheck attaches a check that clients built in this package, and
// in templatefetch, run on every redirect target before following it. It is
// how a caller applies a policy that the client itself knows nothing about,
// such as the operator's allowed hosts. The check does not run on the first
// URL: the caller checks that before making the request.
func WithRedirectCheck(ctx context.Context, check func(target *url.URL) error) context.Context {
	if check == nil {
		return ctx
	}
	return context.WithValue(ctx, redirectCheckKey{}, check)
}

// CheckRedirectTarget runs the check attached to the request's context, if
// any. Call it from an http.Client's CheckRedirect.
func CheckRedirectTarget(req *http.Request) error {
	check, _ := req.Context().Value(redirectCheckKey{}).(func(*url.URL) error)
	if check == nil {
		return nil
	}
	return check(req.URL)
}

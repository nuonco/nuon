package stack

import (
	"fmt"
	"net/url"
	"strings"
)

// parseURL splits a create-run URL into its base and phone_home_id pieces.
// The path must end in `/v1/stack-runs/{phone_home_id}` — anything before
// that becomes part of the base URL. This tolerates the dashboard's
// reverse-proxy prefix (`/api/v1/...`) without making the dashboard care
// about which deployment shape it's rendering for.
//
// Trailing slashes are tolerated. URLs with a `/kind/...` suffix are rejected
// — the kind is appended by the SDK based on the operation, not the user.
// Query strings and fragments are stripped.
func parseURL(raw string) (base, phoneHomeID string, err error) {
	if strings.TrimSpace(raw) == "" {
		return "", "", fmt.Errorf("empty create-run URL")
	}
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return "", "", fmt.Errorf("parse url: %w", err)
	}
	if u.Scheme == "" || u.Host == "" {
		return "", "", fmt.Errorf("create-run URL must include scheme and host (got %q)", raw)
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) < 3 || parts[len(parts)-3] != "v1" || parts[len(parts)-2] != "stack-runs" || parts[len(parts)-1] == "" {
		return "", "", fmt.Errorf("create-run URL path must end in /v1/stack-runs/{phone_home_id} (got %q)", u.Path)
	}
	phoneHomeID = parts[len(parts)-1]
	prefix := strings.Join(parts[:len(parts)-3], "/")
	base = u.Scheme + "://" + u.Host
	if prefix != "" {
		base = base + "/" + prefix
	}
	return base, phoneHomeID, nil
}

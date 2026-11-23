// Package urlpolicy decides whether a URL taken from app config may be used,
// against allowlists a control plane operator configures per category.
//
// The CLI and ctl-api both evaluate URLs with this package, so a URL that
// passes `nuon apps validate` is not rejected by the server for a matching
// reason.
package urlpolicy

import (
	"fmt"
	"net"
	"net/url"
	"strings"
	"unicode"

	"golang.org/x/net/idna"
)

const (
	schemeHTTP  = "http"
	schemeHTTPS = "https"

	wildcardPrefix = "*."
)

// Entry is one allowlist entry: a URL prefix. A bare host is deliberately not
// the unit of trust, because hosts such as raw.githubusercontent.com or
// s3.amazonaws.com serve content from anyone.
type Entry struct {
	raw string

	scheme string
	// host is lowercase and IDNA-normalized, without the wildcard prefix.
	host     string
	wildcard bool
	port     string
	// path is the decoded path prefix. It always starts with "/".
	path string
}

// ParseOptions controls how entries and candidate URLs are parsed.
type ParseOptions struct {
	// AllowHTTP admits http entries. Only local development sets it.
	AllowHTTP bool
}

// ParseEntry parses an allowlist entry. The accepted forms are:
//
//	https://bucket.s3.us-west-2.amazonaws.com/        origin plus path prefix
//	https://raw.githubusercontent.com/acme/runner/    owner-scoped prefix
//	https://*.example.com/templates/                  any subdomain of example.com
//	example.com                                       shorthand for https://example.com/
func ParseEntry(raw string, opts ParseOptions) (Entry, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return Entry{}, fmt.Errorf("entry is empty")
	}

	withScheme := trimmed
	if !strings.Contains(trimmed, "://") {
		withScheme = schemeHTTPS + "://" + trimmed
	}

	// url.Parse rejects "*" in a host on some paths and IDNA rejects it always,
	// so the wildcard label is swapped for a placeholder while parsing.
	wildcard := false
	if rest, ok := strings.CutPrefix(withScheme, schemeHTTPS+"://"+wildcardPrefix); ok {
		wildcard = true
		withScheme = schemeHTTPS + "://" + rest
	} else if rest, ok := strings.CutPrefix(withScheme, schemeHTTP+"://"+wildcardPrefix); ok {
		wildcard = true
		withScheme = schemeHTTP + "://" + rest
	}

	u, err := url.Parse(withScheme)
	if err != nil {
		return Entry{}, fmt.Errorf("entry %q is not a valid URL prefix: %w", raw, err)
	}

	switch u.Scheme {
	case schemeHTTPS:
	case schemeHTTP:
		if !opts.AllowHTTP {
			return Entry{}, fmt.Errorf("entry %q must use https", raw)
		}
	default:
		return Entry{}, fmt.Errorf("entry %q must use https, got scheme %q", raw, u.Scheme)
	}
	if u.Opaque != "" {
		return Entry{}, fmt.Errorf("entry %q is not a valid URL prefix", raw)
	}
	if u.User != nil {
		return Entry{}, fmt.Errorf("entry %q must not contain credentials", raw)
	}
	if u.RawQuery != "" || u.ForceQuery || u.Fragment != "" {
		return Entry{}, fmt.Errorf("entry %q must not contain a query or fragment", raw)
	}
	if strings.Contains(u.Hostname(), "*") {
		return Entry{}, fmt.Errorf("entry %q may only use a wildcard as the leading label (https://*.example.com/)", raw)
	}

	host, err := normalizeHost(u.Hostname())
	if err != nil {
		return Entry{}, fmt.Errorf("entry %q has an invalid host: %w", raw, err)
	}
	if wildcard && net.ParseIP(host) != nil {
		return Entry{}, fmt.Errorf("entry %q cannot combine a wildcard with an IP address", raw)
	}
	if reason := unsafePath(u); reason != "" {
		return Entry{}, fmt.Errorf("entry %q has an invalid path: %s", raw, reason)
	}

	path := u.Path
	if path == "" {
		path = "/"
	}

	return Entry{
		raw:      trimmed,
		scheme:   u.Scheme,
		host:     host,
		wildcard: wildcard,
		port:     effectivePort(u),
		path:     path,
	}, nil
}

// String returns the entry in canonical form.
func (e Entry) String() string {
	host := e.host
	if e.port != defaultPort(e.scheme) {
		host = net.JoinHostPort(host, e.port)
	} else if strings.Contains(host, ":") {
		host = "[" + host + "]"
	}
	if e.wildcard {
		host = wildcardPrefix + host
	}
	return e.scheme + "://" + host + (&url.URL{Path: e.path}).EscapedPath()
}

// matches reports whether the already-vetted candidate falls under the entry.
func (e Entry) matches(c candidate) bool {
	if c.scheme != e.scheme || c.port != e.port {
		return false
	}
	if e.wildcard {
		if !strings.HasSuffix(c.host, "."+e.host) {
			return false
		}
	} else if c.host != e.host {
		return false
	}
	return pathHasPrefix(c.path, e.path)
}

// pathHasPrefix reports whether path is under prefix on a segment boundary, so
// "/runner" admits "/runner" and "/runner/x" but not "/runner-evil".
func pathHasPrefix(path, prefix string) bool {
	if strings.HasSuffix(prefix, "/") {
		return strings.HasPrefix(path, prefix)
	}
	return path == prefix || strings.HasPrefix(path, prefix+"/")
}

type candidate struct {
	scheme string
	host   string
	port   string
	path   string
}

// parseCandidate vets a URL from app config. The returned reason is empty when
// the URL is well-formed enough to be compared against entries.
func parseCandidate(raw string) (candidate, string) {
	if reason := unsafeCharacters(raw); reason != "" {
		return candidate{}, reason
	}
	u, err := url.Parse(raw)
	if err != nil {
		return candidate{}, "it is not a valid URL"
	}
	if u.Scheme != schemeHTTPS && u.Scheme != schemeHTTP {
		if u.Scheme == "" {
			return candidate{}, "it must be an absolute https URL"
		}
		return candidate{}, fmt.Sprintf("scheme %q is not allowed; it must be an https URL", u.Scheme)
	}
	if u.Opaque != "" || u.Hostname() == "" {
		return candidate{}, "it must be an absolute https URL"
	}
	if u.User != nil {
		return candidate{}, "it must not contain credentials"
	}
	host, err := normalizeHost(u.Hostname())
	if err != nil {
		return candidate{}, "its host is not valid"
	}
	if reason := unsafePath(u); reason != "" {
		return candidate{}, reason
	}

	path := u.Path
	if path == "" {
		path = "/"
	}
	// The fragment is ignored: it never reaches the server.
	return candidate{
		scheme: u.Scheme,
		host:   host,
		port:   effectivePort(u),
		path:   path,
	}, ""
}

// shellMetacharacters are rejected in a candidate URL. The value is stored as
// written and later lands in stack templates and scripts, so a URL under an
// allowed prefix must not be able to carry a second command with it.
const shellMetacharacters = "\"'`\\$;|<>"

// unsafeCharacters vets the value as written, before any trimming or parsing,
// because that is the value consumers receive.
func unsafeCharacters(raw string) string {
	for _, r := range raw {
		switch {
		case unicode.IsSpace(r):
			return "it must not contain whitespace"
		case unicode.IsControl(r), unicode.Is(unicode.Cf, r):
			return "it must not contain control or invisible characters"
		case strings.ContainsRune(shellMetacharacters, r):
			return fmt.Sprintf("it must not contain the character %q", r)
		}
	}
	return ""
}

// unsafePath rejects paths a server or an intermediary could resolve to a
// location other than the one the prefix comparison sees.
func unsafePath(u *url.URL) string {
	escaped := strings.ToLower(u.EscapedPath())
	if strings.Contains(escaped, "%2f") || strings.Contains(escaped, "%5c") {
		return "its path must not contain an encoded slash or backslash"
	}
	if strings.Contains(u.Path, `\`) {
		return "its path must not contain a backslash"
	}
	for _, segment := range strings.Split(u.Path, "/") {
		if segment == ".." {
			return `its path must not contain ".." segments`
		}
	}
	return ""
}

func normalizeHost(host string) (string, error) {
	host = strings.TrimSuffix(strings.ToLower(host), ".")
	if host == "" {
		return "", fmt.Errorf("host is empty")
	}
	if ip := net.ParseIP(host); ip != nil {
		return ip.String(), nil
	}
	ascii, err := idna.Lookup.ToASCII(host)
	if err != nil {
		return "", err
	}
	return ascii, nil
}

func effectivePort(u *url.URL) string {
	if port := u.Port(); port != "" {
		return port
	}
	return defaultPort(u.Scheme)
}

func defaultPort(scheme string) string {
	if scheme == schemeHTTP {
		return "80"
	}
	return "443"
}

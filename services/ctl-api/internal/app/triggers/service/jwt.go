package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strings"
	"time"

	"github.com/auth0/go-jwt-middleware/v2/jwks"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type eventJWTClaims struct {
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	expectedEmail string
}

func (c *eventJWTClaims) Validate(context.Context) error {
	if c.expectedEmail != "" && (c.Email != c.expectedEmail || !c.EmailVerified) {
		return errors.New("JWT email identity is not verified")
	}
	return nil
}

const (
	oidcJWKSCacheDuration = 5 * time.Minute
	maxOIDCResponseSize   = 1 << 20
	maxBearerTokenSize    = 16 << 10
)

var nonPublicOIDCPrefixes = []netip.Prefix{
	netip.MustParsePrefix("100.64.0.0/10"),
	netip.MustParsePrefix("192.0.0.0/24"),
	netip.MustParsePrefix("198.18.0.0/15"),
	netip.MustParsePrefix("2001:db8::/32"),
}

func validateOIDCIssuer(value string) error {
	issuer, err := url.Parse(value)
	if err != nil || issuer.Scheme != "https" || issuer.Hostname() == "" || issuer.User != nil || issuer.RawQuery != "" || issuer.Fragment != "" {
		return errors.New("bearer_jwt issuer must be an HTTPS URL without credentials, query, or fragment")
	}
	if strings.EqualFold(issuer.Hostname(), "localhost") {
		return errors.New("bearer_jwt issuer may not address a local network")
	}
	if ip := net.ParseIP(issuer.Hostname()); ip != nil && !publicOIDCIP(ip) {
		return errors.New("bearer_jwt issuer may not address a local network")
	}
	return nil
}

func publicOIDCIP(ip net.IP) bool {
	if ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified() || ip.IsMulticast() {
		return false
	}
	addr, ok := netip.AddrFromSlice(ip)
	if !ok {
		return false
	}
	addr = addr.Unmap()
	for _, prefix := range nonPublicOIDCPrefixes {
		if prefix.Contains(addr) {
			return false
		}
	}
	return true
}

type oidcTransport struct {
	base http.RoundTripper
}

func (t oidcTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.Scheme != "https" {
		return nil, errors.New("OIDC requests require HTTPS")
	}
	resp, err := t.base.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	if resp.ContentLength > maxOIDCResponseSize {
		resp.Body.Close()
		return nil, errors.New("OIDC response is too large")
	}
	resp.Body = struct {
		io.Reader
		io.Closer
	}{Reader: io.LimitReader(resp.Body, maxOIDCResponseSize), Closer: resp.Body}
	return resp, nil
}

func newOIDCHTTPClient() *http.Client {
	dialer := &net.Dialer{Timeout: 5 * time.Second, KeepAlive: 30 * time.Second}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = nil
	transport.DialContext = func(ctx context.Context, network, address string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(address)
		if err != nil {
			return nil, err
		}
		ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
		if err != nil {
			return nil, err
		}
		for _, resolved := range ips {
			if !publicOIDCIP(resolved.IP) {
				return nil, errors.New("OIDC endpoint resolved to a local network")
			}
		}
		if len(ips) == 0 {
			return nil, errors.New("OIDC endpoint did not resolve")
		}
		return dialer.DialContext(ctx, network, net.JoinHostPort(ips[0].IP.String(), port))
	}
	return &http.Client{
		Timeout:   10 * time.Second,
		Transport: oidcTransport{base: transport},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 3 || req.URL.Scheme != "https" {
				return errors.New("invalid OIDC redirect")
			}
			return nil
		},
	}
}

func (s *service) jwtProvider(issuer string) (*jwks.CachingProvider, error) {
	issuerURL, err := url.Parse(issuer)
	if err != nil {
		return nil, err
	}
	s.jwtMu.Lock()
	defer s.jwtMu.Unlock()
	if provider, ok := s.jwtProviders[issuer]; ok {
		return provider, nil
	}
	provider := jwks.NewCachingProvider(issuerURL, oidcJWKSCacheDuration, jwks.WithCustomClient(newOIDCHTTPClient()))
	s.jwtProviders[issuer] = provider
	return provider, nil
}

func (s *service) verifyBearerJWT(ctx *gin.Context, trigger *app.Trigger) error {
	authorization := ctx.GetHeader("Authorization")
	if !strings.HasPrefix(authorization, "Bearer ") || len(authorization) == len("Bearer ") {
		return errors.New("missing bearer token")
	}
	token := strings.TrimPrefix(authorization, "Bearer ")
	if len(token) > maxBearerTokenSize || strings.Count(token, ".") != 2 {
		return errors.New("invalid bearer token format")
	}
	issuer := trigger.AuthConfig.Issuer
	if issuer == "https://accounts.google.com" {
		parts := strings.Split(token, ".")
		claimsJSON, decodeErr := base64.RawURLEncoding.DecodeString(parts[1])
		if decodeErr != nil {
			return decodeErr
		}
		var claims struct {
			Issuer string `json:"iss"`
		}
		if json.Unmarshal(claimsJSON, &claims) != nil || (claims.Issuer != "https://accounts.google.com" && claims.Issuer != "accounts.google.com") {
			return errors.New("invalid Google JWT issuer")
		}
		issuer = claims.Issuer
	}
	provider, err := s.jwtProvider(trigger.AuthConfig.Issuer)
	if err != nil {
		return err
	}
	tokenValidator, err := validator.New(
		provider.KeyFunc,
		validator.RS256,
		issuer,
		trigger.AuthConfig.Audience,
		validator.WithAllowedClockSkew(time.Minute),
		validator.WithCustomClaims(func() validator.CustomClaims {
			return &eventJWTClaims{expectedEmail: trigger.AuthConfig.ExpectedEmail}
		}),
	)
	if err != nil {
		return fmt.Errorf("create JWT validator: %w", err)
	}
	validated, err := tokenValidator.ValidateToken(ctx, token)
	if err != nil {
		return err
	}
	claims, ok := validated.(*validator.ValidatedClaims)
	if !ok {
		return errors.New("unexpected JWT claims")
	}
	if err := requireJWTExpiry(claims); err != nil {
		return err
	}
	if trigger.AuthConfig.ExpectedSubject != "" && claims.RegisteredClaims.Subject != trigger.AuthConfig.ExpectedSubject {
		return errors.New("unexpected JWT subject")
	}
	return nil
}

func requireJWTExpiry(claims *validator.ValidatedClaims) error {
	if claims.RegisteredClaims.Expiry == 0 {
		return errors.New("JWT is missing expiry")
	}
	return nil
}

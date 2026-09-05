package nuonjwtauthextension

import (
	"fmt"
	"net"
	"net/url"
)

const defaultAudience = "urn:nuon:telemetry"

type Config struct {
	Issuer   string `mapstructure:"issuer"`
	Audience string `mapstructure:"audience"`
	JWKSURL  string `mapstructure:"jwks_url"`
}

func (c *Config) Validate() error {
	if err := validateHTTPSOrLoopbackURL("issuer", c.Issuer); err != nil {
		return err
	}
	if c.Audience == "" {
		return fmt.Errorf("audience is required")
	}
	if err := validateHTTPSOrLoopbackURL("JWKS URL", c.JWKSURL); err != nil {
		return err
	}
	return nil
}

func validateHTTPSOrLoopbackURL(name, value string) error {
	parsed, err := url.Parse(value)
	if err != nil || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return fmt.Errorf("%s must be an absolute URL without userinfo, query, or fragment", name)
	}
	if parsed.Scheme == "https" {
		return nil
	}
	if parsed.Scheme != "http" {
		return fmt.Errorf("%s must use HTTPS", name)
	}

	host := parsed.Hostname()
	if host == "localhost" {
		return nil
	}
	if ip := net.ParseIP(host); ip != nil && ip.IsLoopback() {
		return nil
	}
	return fmt.Errorf("%s may use HTTP only on loopback", name)
}

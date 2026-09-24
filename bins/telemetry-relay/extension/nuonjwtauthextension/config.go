package nuonjwtauthextension

import (
	"fmt"
	"net/url"
)

const defaultAudience = "urn:nuon:telemetry"

type Config struct {
	Issuer            string `mapstructure:"issuer"`
	Audience          string `mapstructure:"audience"`
	JWKSURL           string `mapstructure:"jwks_url"`
	JWKSAllowInsecure bool   `mapstructure:"jwks_allow_insecure"`
}

func (c *Config) Validate() error {
	if err := validateEndpointURL("issuer", c.Issuer, c.JWKSAllowInsecure); err != nil {
		return err
	}
	if c.Audience == "" {
		return fmt.Errorf("audience is required")
	}
	if err := validateEndpointURL("JWKS URL", c.JWKSURL, c.JWKSAllowInsecure); err != nil {
		return err
	}
	return nil
}

func validateEndpointURL(name, value string, allowInsecure bool) error {
	parsed, err := url.Parse(value)
	if err != nil || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return fmt.Errorf("%s must be an absolute URL without userinfo, query, or fragment", name)
	}
	if parsed.Scheme == "https" || (parsed.Scheme == "http" && allowInsecure) {
		return nil
	}
	return fmt.Errorf("%s must use HTTPS, or HTTP with jwks_allow_insecure enabled", name)
}

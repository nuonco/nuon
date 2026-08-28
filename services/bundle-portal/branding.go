package main

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

var hexColorPattern = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

type portalBranding struct {
	Name         string `json:"name"`
	LogoURL      string `json:"logo_url,omitempty"`
	FaviconURL   string `json:"favicon_url,omitempty"`
	PrimaryColor string `json:"primary_color"`
	SupportURL   string `json:"support_url,omitempty"`
}

func defaultPortalBranding() portalBranding {
	return portalBranding{Name: "Deployment portal", PrimaryColor: "#8040bf"}
}

func (b portalBranding) Validate() error {
	if strings.TrimSpace(b.Name) == "" {
		return fmt.Errorf("--brand-name cannot be empty")
	}
	if !hexColorPattern.MatchString(b.PrimaryColor) {
		return fmt.Errorf("--brand-primary-color must be a six-digit hex value")
	}
	for name, value := range map[string]string{
		"--brand-logo-url":    b.LogoURL,
		"--brand-favicon-url": b.FaviconURL,
		"--brand-support-url": b.SupportURL,
	} {
		if value == "" {
			continue
		}
		parsed, err := url.Parse(value)
		isHTTPS := err == nil && parsed.Scheme == "https" && parsed.Host != ""
		isSameOrigin := err == nil && parsed.Scheme == "" && parsed.Host == "" && strings.HasPrefix(value, "/") && !strings.HasPrefix(value, "//")
		if !isHTTPS && !isSameOrigin {
			return fmt.Errorf("%s must be a same-origin path or HTTPS URL", name)
		}
	}
	return nil
}

package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPortalBrandingValidation(t *testing.T) {
	require.NoError(t, (portalBranding{Name: "Acme", PrimaryColor: "#123456", LogoURL: "/logo.svg", SupportURL: "https://support.example.com"}).Validate())
	require.Error(t, (portalBranding{Name: "Acme", PrimaryColor: "purple"}).Validate())
	require.Error(t, (portalBranding{Name: "Acme", PrimaryColor: "#123456", LogoURL: "http://example.com/logo.svg"}).Validate())
	require.Error(t, (portalBranding{Name: "Acme", PrimaryColor: "#123456", LogoURL: "https:logo.svg"}).Validate())
	require.Error(t, (portalBranding{Name: "Acme", PrimaryColor: "#123456", LogoURL: "//example.com/logo.svg"}).Validate())
}

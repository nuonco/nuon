package azure

import (
	"strings"
	"testing"

	"github.com/nuonco/nuon/pkg/azure"
)

func TestIMDSTokenRequestURL(t *testing.T) {
	arm := imdsTokenRequestURL("")
	if arm != imdsTokenRequestURL(azure.ManagementAudience) {
		t.Fatalf("empty resource and the ARM audience differ: %s vs %s", arm, imdsTokenRequestURL(azure.ManagementAudience))
	}
	if !strings.Contains(arm, "resource=https://management.azure.com/") {
		t.Fatalf("ARM url = %s", arm)
	}

	const audience = "api://11111111-1111-1111-1111-111111111111"
	got := imdsTokenRequestURL(audience)
	if !strings.Contains(got, "resource=api%3A%2F%2F11111111-1111-1111-1111-111111111111") {
		t.Fatalf("app audience url = %s", got)
	}
	if strings.Contains(got, "management.azure.com") {
		t.Fatalf("app audience url still requests ARM: %s", got)
	}
}

package acr

import (
	"fmt"
	"regexp"
	"strings"
)

// loginServerPattern is a single DNS label under azurecr.io. Registries created
// with a domain name label scope carry a hyphenated hash suffix, so hyphens are
// allowed inside the label.
var loginServerPattern = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?\.azurecr\.io$`)

// ValidateLoginServer rejects anything that is not a bare <registry>.azurecr.io
// host, and any host listed in reserved. The token exchange posts an Entra token
// to this host, so it must only ever be an Azure Container Registry.
func ValidateLoginServer(loginServer string, reserved ...string) error {
	host := strings.ToLower(loginServer)
	if !loginServerPattern.MatchString(host) {
		return fmt.Errorf("invalid ACR login server %q: must be <registry>.azurecr.io with no scheme, port, or path", loginServer)
	}

	for _, r := range reserved {
		if r == "" {
			continue
		}
		if host == normalizeHost(r) {
			return fmt.Errorf("ACR login server %q is reserved and cannot be used as an image source", loginServer)
		}
	}

	return nil
}

// ValidateImageHost requires the image to live on the registry the token is
// minted for.
func ValidateImageHost(loginServer, imageURL string) error {
	host, _, _ := strings.Cut(imageURL, "/")
	if !strings.EqualFold(host, loginServer) {
		return fmt.Errorf("image_url %q must be on the ACR login server %q", imageURL, loginServer)
	}
	return nil
}

func normalizeHost(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.TrimPrefix(s, "https://")
	s = strings.TrimPrefix(s, "http://")
	return strings.TrimSuffix(s, "/")
}

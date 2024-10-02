package registry

import (
	"fmt"
	"strings"
)

type AccessInfoAuth struct {
	// base 64 version
	Encoded string

	// required if encoded not set
	Username string
	Password string

	// set if the authentication should use a custom endpoint
	ServerAddress string
}

type AccessInfo struct {
	Image    string
	Insecure bool

	Auth *AccessInfoAuth
}

// NOTE(jm): in some cases, the image is not a fully resolved path, and we need to use a proxy path or other way of
// referencing the image, such as when a docker index image is used.
func (a *AccessInfo) RepositoryURI() string {
	if a.Auth == nil {
		return a.Image
	}
	if a.Auth.ServerAddress == "" {
		return a.Image
	}

	img := a.Image
	if !strings.HasPrefix(a.Image, a.Auth.ServerAddress) {
		img = fmt.Sprintf("%s/%s", a.Auth.ServerAddress, a.Image)
	}

	// trim any http or https heading
	img = strings.TrimPrefix(img, "https://")
	img = strings.TrimPrefix(img, "http://")

	return img
}

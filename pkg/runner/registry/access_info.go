package registry

import (
	"fmt"
	"strings"
)

type AccessInfoAuth struct {
	Encoded string

	Username string
	Password string

	ServerAddress string
}

type AccessInfo struct {
	Image    string
	Insecure bool

	Auth *AccessInfoAuth
}

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

	img = strings.TrimPrefix(img, "https://")
	img = strings.TrimPrefix(img, "http://")

	return img
}

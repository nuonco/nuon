package version

import "github.com/powertoolsdev/mono/pkg/api/client"

type Service struct {
	api client.Client
}

func New() *Service {
	return &Service{}
}

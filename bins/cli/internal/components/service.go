package components

import "github.com/powertoolsdev/mono/pkg/api/client"

type Service struct {
	api client.Client
}

func New(apiClient client.Client) *Service {
	return &Service{
		api: apiClient,
	}
}

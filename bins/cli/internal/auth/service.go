package auth

import (
	"github.com/nuonco/nuon-go"
)

type Service struct {
	api nuon.Client
}

func New(apiClient nuon.Client) *Service {
	return &Service{
		api: apiClient,
	}
}

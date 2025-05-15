package dev

import (
	"github.com/go-playground/validator/v10"
	"github.com/nuonco/nuon-go"
	"github.com/powertoolsdev/mono/bins/cli/internal/config"
)

type Service struct {
	v           *validator.Validate
	api         nuon.Client
	cfg         *config.Config
	branchName  string
	autoApprove bool
}

func New(v *validator.Validate, apiClient nuon.Client, cfg *config.Config) *Service {
	return &Service{
		v:   v,
		api: apiClient,
		cfg: cfg,
	}
}

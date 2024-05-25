package protos

import (
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/terraformcloud"
)

type Adapter struct {
	cfg         *internal.Config
	orgsOutputs *terraformcloud.OrgsOutputs
}

func New(v *validator.Validate,
	cfg *internal.Config,
	orgsOutputs *terraformcloud.OrgsOutputs) (*Adapter, error) {
	return &Adapter{
		orgsOutputs: orgsOutputs,
		cfg:         cfg,
	}, nil
}

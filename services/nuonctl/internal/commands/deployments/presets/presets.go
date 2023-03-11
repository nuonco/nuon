package presets

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/common/shortid"
	componentv1 "github.com/powertoolsdev/mono/pkg/protos/components/generated/types/component/v1"
)

type preset struct {
	v  *validator.Validate
	ID string `validate:"required"`
}

func New(v *validator.Validate, name string, opts ...presetOption) (*componentv1.Component, error) {
	p := &preset{
		v:  v,
		ID: shortid.New(),
	}
	for idx, opt := range opts {
		if err := opt(p); err != nil {
			return nil, fmt.Errorf("option %d failed: %w", idx, err)
		}
	}

	if err := p.v.Struct(p); err != nil {
		return nil, fmt.Errorf("unable to validate temporal: %w", err)
	}

	var fn func() (*componentv1.Component, error)
	switch name {
	// httpbin deployed from a public image (ie: kennethreitz/httpbin)
	case "public_external_image_httpbin":
		fn = p.publicExternalImageHttpbin
	// httpbin deployed in a private ECR repository
	case "private_external_image_httpbin":
		fn = p.privateExternalImageHttpbin
	// public docker repo containing httpbin (ie: github.com/kennethreitz/httpbin)
	case "public_docker_httpbin":
		fn = p.publicDockerHttpbin
	// private docker repo that contains httpbin
	case "private_docker_httpbin":
		fn = p.privateDockerHttpbin
	case "public_helm_chart":
		fn = p.publicHelmChart
	default:
		return nil, fmt.Errorf("invalid preset: %s", name)
	}

	return fn()
}

type presetOption func(*preset) error

func WithID(id string) presetOption {
	return func(p *preset) error {
		p.ID = id
		return nil
	}
}

package multi

import (
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/kube"
)

type Config struct {
	AddressTemplate    string            `validate:"required"`
	SecretNameTemplate string            `validate:"required"`
	SecretNamespace    string            `validate:"required"`
	SecretKey          string            `validate:"required"`
	ClusterInfo        *kube.ClusterInfo `validate:"required"`
}

type multiClient struct {
	Config *Config `validate:"required"`

	v *validator.Validate
}

type clientOption func(*multiClient) error

func New(v *validator.Validate, opts ...clientOption) (*multiClient, error) {
	p := &multiClient{v: v}

	for _, opt := range opts {
		if err := opt(p); err != nil {
			return nil, err
		}
	}

	if err := p.v.Struct(p); err != nil {
		return nil, err
	}

	return p, nil
}

func WithConfig(cfg *Config) clientOption {
	return func(p *multiClient) error {
		p.Config = cfg
		return nil
	}
}

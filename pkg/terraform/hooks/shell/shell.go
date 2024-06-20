package shell

import (
	"fmt"

	"github.com/go-playground/validator/v10"

	"github.com/powertoolsdev/mono/pkg/aws/credentials"
	"github.com/powertoolsdev/mono/pkg/terraform/hooks"
)

const (
	hookPreApply   string = "pre-apply.sh"
	hookPostApply  string = "post-apply.sh"
	hookErrorApply string = "error-apply.sh"

	hookPreDestroy   string = "pre-destroy.sh"
	hookPostDestroy  string = "post-destroy.sh"
	hookErrorDestroy string = "error-destroy.sh"
)

type shell struct {
	v       *validator.Validate
	rootDir string

	EnvVars map[string]string `validate:"required"`
	Auth    *credentials.Config
}

var _ hooks.Hooks = (*shell)(nil)

func New(v *validator.Validate, opts ...shellOption) (*shell, error) {
	s := &shell{
		v: v,
	}

	for idx, opt := range opts {
		if err := opt(s); err != nil {
			return nil, fmt.Errorf("unable to set %d option: %w", idx, err)
		}
	}
	if err := s.v.Struct(s); err != nil {
		return nil, err
	}

	return s, nil
}

type shellOption func(*shell) error

func WithEnvVars(envVars map[string]string) shellOption {
	return func(v *shell) error {
		v.EnvVars = envVars
		return nil
	}
}

func WithRunAuth(cfg *credentials.Config) shellOption {
	return func(v *shell) error {
		v.Auth = cfg
		return nil
	}
}

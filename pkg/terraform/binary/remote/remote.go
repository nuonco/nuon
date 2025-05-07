package remote

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hc-install/releases"
	"github.com/powertoolsdev/mono/pkg/terraform/binary"
)

type remote struct {
	v *validator.Validate

	Version *version.Version `validate:"required"`

	version *releases.ExactVersion
}

var _ binary.Binary = (*remote)(nil)

type remoteOption func(*remote) error

func New(v *validator.Validate, opts ...remoteOption) (*remote, error) {
	auth := &remote{
		v: v,
	}

	for _, opt := range opts {
		if err := opt(auth); err != nil {
			return nil, err
		}
	}

	if err := auth.v.Struct(auth); err != nil {
		return nil, err
	}
	return auth, nil
}

func WithVersion(v string) remoteOption {
	return func(s *remote) error {
		ver, err := version.NewVersion(v)
		if err != nil {
			return fmt.Errorf("invalid version: %w", err)
		}
		s.Version = ver
		return nil
	}
}

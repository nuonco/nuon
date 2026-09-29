package local

import (
	"context"
	"fmt"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/go-hclog"

	"github.com/nuonco/nuon/pkg/terraform/binary"
)

type local struct {
	v *validator.Validate

	Path string `validate:"required"`
}

var _ binary.Binary = (*local)(nil)

type localOption func(*local) error

func New(v *validator.Validate, opts ...localOption) (*local, error) {
	l := &local{v: v}
	for _, opt := range opts {
		if err := opt(l); err != nil {
			return nil, err
		}
	}
	if err := l.v.Struct(l); err != nil {
		return nil, err
	}
	return l, nil
}

func WithPath(p string) localOption {
	return func(l *local) error {
		l.Path = p
		return nil
	}
}

func (l *local) Install(_ context.Context, _ hclog.Logger, _ string) (string, error) {
	info, err := os.Stat(l.Path)
	if err != nil {
		return "", fmt.Errorf("bundled terraform binary at %s: %w", l.Path, err)
	}
	if info.IsDir() {
		return "", fmt.Errorf("bundled terraform binary path %s is a directory", l.Path)
	}
	if info.Mode()&0o111 == 0 {
		if err := os.Chmod(l.Path, 0o755); err != nil {
			return "", fmt.Errorf("unable to chmod bundled terraform binary at %s: %w", l.Path, err)
		}
	}
	return l.Path, nil
}

func (l *local) Init(_ context.Context) error {
	return nil
}

func (l *local) Source() string {
	return "local"
}

func (l *local) Uninstall(_ context.Context) error {
	return nil
}

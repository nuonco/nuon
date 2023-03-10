package install

import (
	"context"
	"fmt"
	"log"

	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
)

type terraformInstaller struct {
	Logger  *log.Logger `validate:"required"`
	Dir     string      `validate:"required,dir,min=1"`
	Version string      `validate:"required,min=5"`

	// internal state
	installer installer
	remover   remover
	validator *validator.Validate
}

type terraformInstallerOption func(*terraformInstaller) error

// New instantiates a new terraform installer
func New(v *validator.Validate, opts ...terraformInstallerOption) (*terraformInstaller, error) {
	t := &terraformInstaller{}

	if v == nil {
		return nil, fmt.Errorf("error instantiating terraform installer: validator is nil")
	}
	t.validator = v

	for _, opt := range opts {
		if err := opt(t); err != nil {
			return nil, err
		}
	}

	if err := t.validator.Struct(t); err != nil {
		return nil, err
	}

	i, err := t.initInstaller()
	if err != nil {
		return nil, err
	}

	t.installer = i
	t.remover = i

	return t, nil
}

// WithLogger sets the logger to use while installing
func WithLogger(l *log.Logger) terraformInstallerOption {
	return func(t *terraformInstaller) error {
		t.Logger = l
		return nil
	}
}

// WithInstallDir sets the directory that terraform should be installed into
// typically should be os.TempDir()
func WithInstallDir(d string) terraformInstallerOption {
	return func(t *terraformInstaller) error {
		t.Dir = d
		return nil
	}
}

// WithVersion sets the version of terraform to install
func WithVersion(v string) terraformInstallerOption {
	return func(t *terraformInstaller) error {
		t.Version = v
		return nil
	}
}

type installer interface {
	Install(context.Context) (string, error)
}

// Install installs terraform
func (t *terraformInstaller) Install(ctx context.Context) (string, error) {
	// TODO(jdt): check for existing install?
	execPath, err := t.installer.Install(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to install terraform: %w", err)
	}

	return execPath, nil
}

type remover interface {
	Remove(context.Context) error
}

// Cleanup cleans up the terraform install
func (t *terraformInstaller) Cleanup() error {
	// NOTE(jdt): not sure that background context is the best but not sure what would be more appropriate?
	// this would typically get called as the process is exiting?
	return t.remover.Remove(context.Background())
}

func (t *terraformInstaller) initInstaller() (*releases.ExactVersion, error) {
	ver, err := version.NewVersion(t.Version)
	if err != nil {
		return nil, err
	}
	installer := &releases.ExactVersion{
		Product:    product.Terraform,
		Version:    ver,
		InstallDir: t.Dir,
	}
	installer.SetLogger(t.Logger)
	return installer, nil
}

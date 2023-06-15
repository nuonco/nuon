package oci

// package oci exposes methods for working with oci archives
import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/powertoolsdev/mono/pkg/terraform/archive"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/file"
)

// Package oci exposes an archive that loads a terraform archive from an oci artifact
var _ archive.Archive = (*oci)(nil)

type oci struct {
	v *validator.Validate

	Image *Image `validate:"required"`
	Auth  *Auth  `validate:"required"`

	store *file.Store

	// the following fields are exposed to make this more easily tested
	testSrc oras.ReadOnlyTarget
	tmpDir  string
}

type ociOption func(*oci) error

func New(v *validator.Validate, opts ...ociOption) (*oci, error) {
	s := &oci{
		v:      v,
		tmpDir: filepath.Join(os.TempDir(), fmt.Sprintf("oci-archive-%s", uuid.NewString())),
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

// WithAuth sets the iam role and ecr token to use with this archive
func WithAuth(auth *Auth) ociOption {
	return func(o *oci) error {
		if err := auth.Validate(o.v); err != nil {
			return fmt.Errorf("invalid auth: %w", err)
		}

		o.Auth = auth
		return nil
	}
}

// WithImage sets the image
func WithImage(img *Image) ociOption {
	return func(o *oci) error {
		o.Image = img
		return nil
	}
}

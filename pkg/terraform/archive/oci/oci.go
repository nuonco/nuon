package oci

// package oci exposes methods for working with oci archives
import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/terraform/archive"
)

// Package oci exposes an archive that loads a terraform archive from an oci artifact
var _ archive.Archive = (*oci)(nil)

type oci struct {
	v *validator.Validate

	// RoleARN is used to load the oci artifact, and assumes that the user name is AWS (because all of our
	// repos/uses are ECR).
	AuthToken string

	// TODO(jm): we will eventually load oci archives using an IAM role, to fetch from the local repo, but for now
	// we don't support that, so we just fetch using the hard coded credentials.
	RoleARN         string
	RoleSessionName string
}

type ociOption func(*oci) error

func New(v *validator.Validate, opts ...ociOption) (*oci, error) {
	s := &oci{
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

// WithRoleArn sets the bucket role
func WithRoleARN(arn string) ociOption {
	return func(s *oci) error {
		s.RoleARN = arn
		return nil
	}
}

// WithRoleSessionName sets the role session name
func WithRoleSessionName(name string) ociOption {
	return func(s *oci) error {
		s.RoleSessionName = name
		return nil
	}
}

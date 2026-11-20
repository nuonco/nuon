// Package remote implements a binary.Binary that downloads the OpenTofu CLI
// via github.com/opentofu/tofudl and installs it into the workspace's bins/
// directory. It is the OpenTofu counterpart of
// pkg/terraform/binary/remote, which downloads terraform from
// releases.hashicorp.com via hc-install — OpenTofu does not publish to
// releases.hashicorp.com, so it needs its own downloader.
//
// The binary.Binary interface itself (and the local path-based installer,
// pkg/terraform/binary/local) is reused from pkg/terraform/binary unchanged.
package remote

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/opentofu/tofudl"

	"github.com/nuonco/nuon/pkg/terraform/binary"
)

type remote struct {
	v *validator.Validate

	// Version is the OpenTofu version to download, pinned at construction
	// time. tofudl requires bare `X.Y.Z` (optionally `-alpha/-beta/-rc`);
	// an optional leading `v` from callers is stripped in WithVersion.
	Version tofudl.Version `validate:"required"`

	// installedPath is set by Install and removed by Uninstall.
	installedPath string
}

var _ binary.Binary = (*remote)(nil)

type remoteOption func(*remote) error

func New(v *validator.Validate, opts ...remoteOption) (*remote, error) {
	r := &remote{
		v: v,
	}

	for _, opt := range opts {
		if err := opt(r); err != nil {
			return nil, err
		}
	}

	if err := r.v.Struct(r); err != nil {
		return nil, err
	}
	return r, nil
}

// Source returns the binary source label used by tracing in
// pkg/opentofu/workspace.LoadBinary. "remote" means the binary is downloaded
// via tofudl; the cost shows up under the tofu.binary_install span emitted by
// Install.
func (r *remote) Source() string {
	return "remote"
}

// WithVersion pins the OpenTofu version to download. A single leading "v" is
// accepted (the convention go-version-based callers use) and stripped, since
// tofudl's version format is bare X.Y.Z. The version must parse as a valid
// OpenTofu version or New fails.
func WithVersion(version string) remoteOption {
	return func(r *remote) error {
		v := version
		if len(v) > 0 && v[0] == 'v' {
			v = v[1:]
		}
		if err := tofudl.Version(v).Validate(); err != nil {
			return fmt.Errorf("invalid version: %w", err)
		}
		r.Version = tofudl.Version(v)
		return nil
	}
}

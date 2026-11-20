package validation

import (
	"fmt"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/opentofu/tofudl"

	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

const MinTerraformVersion = "1.8.0"

const MinOpentofuVersion = "1.13.0"

// ValidateOpentofuMinVersion validates an OpenTofu version string against the
// minimum constraint and the format tofudl can download (the runner installs
// binaries via tofudl, which rejects versions like "1.13" or "+meta" builds).
// There is no upper bound because OpenTofu has no registry lookup in v1.
func ValidateOpentofuMinVersion(version string) error {
	if version == "" {
		return stderr.ErrUser{
			Err:         fmt.Errorf("opentofu version is required"),
			Description: "OpenTofu version cannot be empty",
		}
	}

	normalized := strings.TrimPrefix(version, "v")
	if err := tofudl.Version(normalized).Validate(); err != nil {
		return stderr.ErrUser{
			Err:         fmt.Errorf("invalid opentofu version: %s", version),
			Description: fmt.Sprintf("OpenTofu version '%s' is not a downloadable release (use X.Y.Z, e.g. 1.13.9)", version),
		}
	}

	constraint, err := semver.NewConstraint(fmt.Sprintf(">= %s", MinOpentofuVersion))
	if err != nil {
		return stderr.ErrUser{
			Err:         fmt.Errorf("unable to parse version constraint: %w", err),
			Description: fmt.Sprintf("Failed to parse OpenTofu minimum version constraint: %s", MinOpentofuVersion),
		}
	}

	ver, err := semver.NewVersion(version)
	if err != nil {
		return stderr.ErrUser{
			Err:         fmt.Errorf("invalid opentofu version: %s", version),
			Description: fmt.Sprintf("OpenTofu version '%s' is not a valid semver version", version),
		}
	}

	if !constraint.Check(ver) {
		return stderr.ErrUser{
			Err:         fmt.Errorf("opentofu version %s is below minimum required version %s", version, MinOpentofuVersion),
			Description: fmt.Sprintf("OpenTofu version '%s' must be >= %s", version, MinOpentofuVersion),
		}
	}

	return nil
}

// ValidateTerraformVersion validates a Terraform version string against min/max constraints.
func ValidateTerraformVersion(version string, minVersion string, maxVersion string) error {
	if version == "" {
		return stderr.ErrUser{
			Err:         fmt.Errorf("terraform version is required"),
			Description: "Terraform version cannot be empty",
		}
	}

	// Parse version constraints
	minConstraint := fmt.Sprintf(">= %s", minVersion)
	maxConstraint := fmt.Sprintf("<= %s", maxVersion)
	constraint, err := semver.NewConstraint(fmt.Sprintf("%s, %s", minConstraint, maxConstraint))
	if err != nil {
		return stderr.ErrUser{
			Err:         fmt.Errorf("unable to parse version constraints: %w", err),
			Description: fmt.Sprintf("Failed to parse Terraform version constraints (min: %s, max: %s)", minVersion, maxVersion),
		}
	}

	// Parse provided version
	ver, err := semver.NewVersion(version)
	if err != nil {
		return stderr.ErrUser{
			Err:         fmt.Errorf("invalid terraform version: %s", version),
			Description: fmt.Sprintf("Terraform version '%s' is not a valid semver version", version),
		}
	}

	// Check if version satisfies constraints
	if !constraint.Check(ver) {
		return stderr.ErrUser{
			Err:         fmt.Errorf("terraform version out of range: %s", version),
			Description: fmt.Sprintf("Terraform version '%s' does not satisfy constraints: must be between %s and %s", version, minVersion, maxVersion),
		}
	}

	return nil
}

// ValidateTerraformMinVersion validates a Terraform version string against only the minimum constraint.
// Used by the DB syncer which does not have access to tfClient.GetLatestVersion() for the upper bound.
func ValidateTerraformMinVersion(version string) error {
	if version == "" {
		return stderr.ErrUser{
			Err:         fmt.Errorf("terraform version is required"),
			Description: "Terraform version cannot be empty",
		}
	}

	constraint, err := semver.NewConstraint(fmt.Sprintf(">= %s", MinTerraformVersion))
	if err != nil {
		return stderr.ErrUser{
			Err:         fmt.Errorf("unable to parse version constraint: %w", err),
			Description: fmt.Sprintf("Failed to parse Terraform minimum version constraint: %s", MinTerraformVersion),
		}
	}

	ver, err := semver.NewVersion(version)
	if err != nil {
		return stderr.ErrUser{
			Err:         fmt.Errorf("invalid terraform version: %s", version),
			Description: fmt.Sprintf("Terraform version '%s' is not a valid semver version", version),
		}
	}

	if !constraint.Check(ver) {
		return stderr.ErrUser{
			Err:         fmt.Errorf("terraform version %s is below minimum required version %s", version, MinTerraformVersion),
			Description: fmt.Sprintf("Terraform version '%s' must be >= %s", version, MinTerraformVersion),
		}
	}

	return nil
}

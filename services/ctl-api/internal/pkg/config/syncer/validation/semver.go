package validation

import (
	"fmt"

	"github.com/Masterminds/semver/v3"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

// ValidateTerraformVersion validates a Terraform version string against min/max constraints.
// Duplicates logic from services/ctl-api/internal/app/components/service/create_terraform_module_component_config.go
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

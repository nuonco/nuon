package validation

import (
	"fmt"
	"regexp"

	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

var (
	interpolatedNameRegex = regexp.MustCompile(`^[a-z0-9_{}\.]*$`)
	entityNameRegex       = regexp.MustCompile(`^[a-z0-9_-]*$`)
	dnsRFC1035Regex       = regexp.MustCompile(`^[a-z]([-a-z0-9]*[a-z0-9])?$`)
)

// ValidateInterpolatedName validates a name that allows interpolation syntax.
// Allows: lowercase letters, numbers, underscores, dots, and curly braces
// Duplicates logic from services/ctl-api/internal/pkg/validator/interpolated_name.go
func ValidateInterpolatedName(name string) error {
	if name == "" {
		return nil // Empty is allowed for optional fields
	}

	if !interpolatedNameRegex.MatchString(name) {
		return stderr.ErrUser{
			Err:         fmt.Errorf("invalid name: %s", name),
			Description: fmt.Sprintf("Name '%s' must contain only lowercase letters, numbers, underscores, dots, and curly braces (for interpolation)", name),
		}
	}

	return nil
}

// ValidateEntityName validates a standard entity name.
// Allows: lowercase letters, numbers, underscores, and hyphens
// Duplicates logic from services/ctl-api/internal/pkg/validator/entity_name.go
func ValidateEntityName(name string) error {
	if name == "" {
		return stderr.ErrUser{
			Err:         fmt.Errorf("name is required"),
			Description: "Name cannot be empty",
		}
	}

	if !entityNameRegex.MatchString(name) {
		return stderr.ErrUser{
			Err:         fmt.Errorf("invalid name: %s", name),
			Description: fmt.Sprintf("Name '%s' must contain only lowercase letters, numbers, underscores, and hyphens", name),
		}
	}

	return nil
}

// ValidateDNSName validates a DNS RFC 1035 label (used for Helm chart names, K8s resources, etc).
// Must be 1-63 characters, start with a letter, and contain only lowercase letters, numbers, and hyphens.
func ValidateDNSName(name string, minLen, maxLen int) error {
	if name == "" {
		return stderr.ErrUser{
			Err:         fmt.Errorf("name is required"),
			Description: "Name cannot be empty",
		}
	}

	if len(name) < minLen {
		return stderr.ErrUser{
			Err:         fmt.Errorf("name too short: %s", name),
			Description: fmt.Sprintf("Name '%s' must be at least %d characters", name, minLen),
		}
	}

	if len(name) > maxLen {
		return stderr.ErrUser{
			Err:         fmt.Errorf("name too long: %s", name),
			Description: fmt.Sprintf("Name '%s' cannot exceed %d characters", name, maxLen),
		}
	}

	if !dnsRFC1035Regex.MatchString(name) {
		return stderr.ErrUser{
			Err:         fmt.Errorf("invalid DNS name: %s", name),
			Description: fmt.Sprintf("Name '%s' must start with a lowercase letter and contain only lowercase letters, numbers, and hyphens", name),
		}
	}

	return nil
}

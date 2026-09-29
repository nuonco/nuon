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
	dnsRFC1123Regex       = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`)
)

func ValidateDNSSubdomain(name string) error {
	if len(name) > 253 {
		return stderr.ErrUser{
			Err:         fmt.Errorf("name too long: %s", name),
			Description: fmt.Sprintf("Name '%s' cannot exceed 253 characters", name),
		}
	}
	if !dnsRFC1123Regex.MatchString(name) {
		return stderr.ErrUser{
			Err:         fmt.Errorf("invalid DNS subdomain: %s", name),
			Description: fmt.Sprintf("Name '%s' must be a valid DNS RFC 1123 subdomain", name),
		}
	}
	return nil
}

func ValidateInterpolatedName(name string) error {
	if name == "" {
		return nil
	}

	if !interpolatedNameRegex.MatchString(name) {
		return stderr.ErrUser{
			Err:         fmt.Errorf("invalid name: %s", name),
			Description: fmt.Sprintf("Name '%s' must contain only lowercase letters, numbers, underscores, dots, and curly braces (for interpolation)", name),
		}
	}

	return nil
}

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

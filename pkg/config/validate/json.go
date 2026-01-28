package validate

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/xeipuuv/gojsonschema"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/schema"
)

// componentIndexPattern matches error fields like "components.9" or "components.9.name"
var componentIndexPattern = regexp.MustCompile(`^components\.(\d+)`)

// policyIndexPattern matches error fields like "policies.policy.6" or "policies.policy.6.type"
var policyIndexPattern = regexp.MustCompile(`^policies\.policy\.(\d+)`)

func ValidateJSONSchema(ctx context.Context, c *config.AppConfig) error {
	errs, err := schema.Validate(ctx, c)
	if err != nil {
		return err
	}

	if len(errs) < 1 {
		return nil
	}

	// Format errors with file names where possible
	formattedErrs := formatValidationErrors(errs, c)

	return config.ErrConfig{
		Description: strings.Join(formattedErrs, "\n"),
	}
}

// formatValidationErrors converts schema validation errors to user-friendly messages,
// replacing component and policy array indices with their source file names where available.
func formatValidationErrors(errs []gojsonschema.ResultError, c *config.AppConfig) []string {
	result := make([]string, 0, len(errs))

	for _, err := range errs {
		field := err.Field()
		desc := err.Description()

		// Try to replace component index with file name
		if matches := componentIndexPattern.FindStringSubmatch(field); len(matches) == 2 {
			if idx, parseErr := strconv.Atoi(matches[1]); parseErr == nil {
				if idx >= 0 && idx < len(c.Components) && c.Components[idx] != nil {
					sourceFile := c.Components[idx].GetSourceFile()
					if sourceFile != "" {
						// Replace "components.9" with the file name
						suffix := strings.TrimPrefix(field, matches[0])
						if suffix != "" {
							field = sourceFile + suffix
						} else {
							field = sourceFile
						}
					}
				}
			}
		}

		// Try to replace policy index with file name and line number
		if matches := policyIndexPattern.FindStringSubmatch(field); len(matches) == 2 {
			if idx, parseErr := strconv.Atoi(matches[1]); parseErr == nil {
				if c.Policies != nil && idx >= 0 && idx < len(c.Policies.Policies) {
					policy := c.Policies.Policies[idx]
					sourceFile := policy.GetSourceFile()
					sourceLine := policy.GetSourceLine()
					if sourceFile != "" {
						suffix := strings.TrimPrefix(field, matches[0])
						if sourceLine > 0 {
							// Use line number format: "file:L<line>"
							if suffix != "" {
								field = fmt.Sprintf("%s:L%d%s", sourceFile, sourceLine, suffix)
							} else {
								field = fmt.Sprintf("%s:L%d", sourceFile, sourceLine)
							}
						} else {
							// Fallback to policy number if no line info
							policyNum := idx + 1
							if suffix != "" {
								field = fmt.Sprintf("%s (policy %d)%s", sourceFile, policyNum, suffix)
							} else {
								field = fmt.Sprintf("%s (policy %d)", sourceFile, policyNum)
							}
						}
					}
				}
			}
		}

		result = append(result, fmt.Sprintf("%s: %s", field, desc))
	}

	return result
}

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

var componentIndexPattern = regexp.MustCompile(`^components\.(\d+)`)

var policyIndexPattern = regexp.MustCompile(`^policies\.policy\.(\d+)`)

func ValidateJSONSchema(ctx context.Context, c *config.AppConfig) error {
	errs, err := schema.Validate(ctx, c)
	if err != nil {
		return err
	}

	if len(errs) < 1 {
		return nil
	}

	formattedErrs := formatValidationErrors(errs, c)

	return config.ErrConfig{
		Description: strings.Join(formattedErrs, "\n"),
	}
}

func formatValidationErrors(errs []gojsonschema.ResultError, c *config.AppConfig) []string {
	result := make([]string, 0, len(errs))

	for _, err := range errs {
		field := err.Field()
		desc := err.Description()

		if matches := componentIndexPattern.FindStringSubmatch(field); len(matches) == 2 {
			if idx, parseErr := strconv.Atoi(matches[1]); parseErr == nil {
				if idx >= 0 && idx < len(c.Components) && c.Components[idx] != nil {
					sourceFile := c.Components[idx].GetSourceFile()
					if sourceFile != "" {
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

		if matches := policyIndexPattern.FindStringSubmatch(field); len(matches) == 2 {
			if idx, parseErr := strconv.Atoi(matches[1]); parseErr == nil {
				if c.Policies != nil && idx >= 0 && idx < len(c.Policies.Policies) {
					policy := c.Policies.Policies[idx]
					sourceFile := policy.GetSourceFile()
					sourceLine := policy.GetSourceLine()
					if sourceFile != "" {
						suffix := strings.TrimPrefix(field, matches[0])
						if sourceLine > 0 {
							if suffix != "" {
								field = fmt.Sprintf("%s:L%d%s", sourceFile, sourceLine, suffix)
							} else {
								field = fmt.Sprintf("%s:L%d", sourceFile, sourceLine)
							}
						} else {
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

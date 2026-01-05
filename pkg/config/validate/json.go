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
// replacing component array indices with their source file names where available.
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

		result = append(result, fmt.Sprintf("%s: %s", field, desc))
	}

	return result
}

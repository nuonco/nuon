package validate

import (
	"context"
	"strings"

	"github.com/powertoolsdev/mono/pkg/config"
	"github.com/powertoolsdev/mono/pkg/config/schema"
	"github.com/powertoolsdev/mono/pkg/generics"
)

func ValidateJSONSchema(ctx context.Context, c *config.AppConfig) error {
	errs, err := schema.Validate(ctx, c)
	if err != nil {
		return err
	}

	if len(errs) < 1 {
		return nil
	}
	return config.ErrConfig{
		Description: strings.Join(generics.SliceToStrings(errs), "\n"),
	}
}

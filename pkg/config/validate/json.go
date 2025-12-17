package validate

import (
	"context"
	"strings"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/schema"
	"github.com/nuonco/nuon/pkg/generics"
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

package activities

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/go-playground/validator/v10"
	"go.temporal.io/sdk/temporal"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/parse"
	"github.com/nuonco/nuon/pkg/config/validate"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/branchrunerrors"
)

// @temporal-gen-v2 activity
// @start-to-close-timeout 5m
// @as-wrapper
// @by-field sourceDir
// @local
func (a *Activities) fetchIntermediateConfig(ctx context.Context, sourceDir string) (*config.AppConfig, error) {
	defer os.RemoveAll(sourceDir)

	v := validator.New()
	parseResult, err := parse.ParseDirWithSource(ctx, parse.ParseConfig{
		Dirname:       sourceDir,
		V:             v,
		FileProcessor: func(name string, obj map[string]any) map[string]any { return obj },
	})
	if err != nil {
		return nil, fmt.Errorf("unable to parse config from repo: %w", err)
	}
	cfg := parseResult.Config

	if cfg.CustomerManaged != nil {
		cfg.SourceArchive = parseResult.Source
	}

	if err := validate.Validate(ctx, v, cfg); err != nil {
		var configErr config.ErrConfig
		if errors.As(err, &configErr) {
			return nil, temporal.NewNonRetryableApplicationError(
				configErr.Description,
				branchrunerrors.ConfigValidationFailedTemporalType,
				err,
			)
		}
		return nil, err
	}

	return cfg, nil
}

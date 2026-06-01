package activities

import (
	"context"
	"fmt"
	"os"

	"github.com/go-playground/validator/v10"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/parse"
)

// @temporal-gen-v2 activity
// @start-to-close-timeout 5m
// @as-wrapper
// @by-field sourceDir
func (a *Activities) fetchIntermediateConfig(ctx context.Context, sourceDir string) (*config.AppConfig, error) {
	defer os.RemoveAll(sourceDir)

	cfg, err := parse.ParseDir(ctx, parse.ParseConfig{
		Dirname:       sourceDir,
		V:             validator.New(),
		FileProcessor: func(name string, obj map[string]any) map[string]any { return obj },
	})
	if err != nil {
		return nil, fmt.Errorf("unable to parse config from repo: %w", err)
	}

	return cfg, nil
}

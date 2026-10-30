package activities

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/go-playground/validator/v10"
	"go.temporal.io/sdk/temporal"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/parse"
	"github.com/nuonco/nuon/pkg/config/validate"
	"github.com/nuonco/nuon/pkg/workspace"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/branchrunerrors"
)

// @temporal-gen-v2 activity
// @start-to-close-timeout 5m
// @as-wrapper
// @local
func (a *Activities) fetchIntermediateConfig(ctx context.Context, sourceDir string, workspaceID string) (_ *config.AppConfig, retErr error) {
	if !filepath.IsLocal(workspaceID) || workspaceID == "." || strings.ContainsAny(workspaceID, `/\`) {
		return nil, fmt.Errorf("invalid workspace ID %q", workspaceID)
	}
	v := validator.New()
	ws, err := workspace.New(v, workspace.WithID(workspaceID))
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := ws.Cleanup(context.WithoutCancel(ctx)); err != nil {
			retErr = errors.Join(retErr, err)
		}
	}()
	parent, err := filepath.EvalSymlinks(filepath.Dir(ws.Root()))
	if err != nil {
		return nil, err
	}

	parseResult, err := parse.ParseDirWithSource(ctx, parse.ParseConfig{
		Dirname:       sourceDir,
		SourceRoot:    filepath.Join(parent, workspaceID),
		V:             v,
		FileProcessor: func(name string, obj map[string]any) map[string]any { return obj },
		SkipBranches:  true,
		SkipInstalls:  true,
	})
	if err != nil {
		return nil, configValidationFailed(err)
	}
	cfg := parseResult.Config
	cfg.SourceArchive = parseResult.Source

	if err := validate.Validate(ctx, v, cfg); err != nil {
		return nil, configValidationFailed(err)
	}

	return cfg, nil
}

func configValidationFailed(err error) error {
	return temporal.NewNonRetryableApplicationError(
		branchrunerrors.UserDetail(err),
		branchrunerrors.ConfigValidationFailedTemporalType,
		err,
	)
}

package activities

import (
	"context"
	"encoding/json"
	"fmt"

	pkgconfig "github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
)

type GetInstallsVCSConfigOutput struct {
	InstallsDir  string `json:"installs_dir"`
	HasVCSConfig bool   `json:"has_vcs_config"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
// @as-wrapper
// @by-field appID
func (a *Activities) getInstallsVCSConfig(ctx context.Context, appID string) (*GetInstallsVCSConfigOutput, error) {
	var latestConfig app.AppConfig
	if err := a.db.WithContext(ctx).
		Where(app.AppConfig{AppID: appID}).
		Order("created_at DESC").
		First(&latestConfig).Error; err != nil {
		return nil, fmt.Errorf("unable to get latest app config: %w", err)
	}

	if latestConfig.IntermediateConfig == nil || !latestConfig.IntermediateConfig.IsSet() {
		return &GetInstallsVCSConfigOutput{}, nil
	}

	blobCtx := blobstore.WithBlobService(ctx, a.blobSvc)
	intermediateJSON, err := latestConfig.IntermediateConfig.Get(blobCtx)
	if err != nil || intermediateJSON == "" {
		return &GetInstallsVCSConfigOutput{}, nil
	}

	var parsed pkgconfig.AppConfig
	if err := json.Unmarshal([]byte(intermediateJSON), &parsed); err != nil {
		return nil, fmt.Errorf("unable to parse intermediate config: %w", err)
	}

	if parsed.InstallsConfig == nil {
		return &GetInstallsVCSConfigOutput{}, nil
	}

	ic := parsed.InstallsConfig
	dir := ""
	if ic.ConnectedRepo != nil {
		dir = ic.ConnectedRepo.Directory
	} else if ic.PublicRepo != nil {
		dir = ic.PublicRepo.Directory
	}

	return &GetInstallsVCSConfigOutput{
		InstallsDir:  dir,
		HasVCSConfig: true,
	}, nil
}

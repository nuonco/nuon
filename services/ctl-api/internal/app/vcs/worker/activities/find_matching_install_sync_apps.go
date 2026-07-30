package activities

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	pkgconfig "github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
)

type FindMatchingInstallSyncAppsRequest struct {
	OrgID  string `json:"org_id" validate:"required"`
	Repo   string `json:"repo" validate:"required"`
	Branch string `json:"branch" validate:"required"`
}

type MatchingInstallSyncApp struct {
	AppID string `json:"app_id"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 2m
func (a *Activities) FindMatchingInstallSyncApps(ctx context.Context, req FindMatchingInstallSyncAppsRequest) ([]MatchingInstallSyncApp, error) {
	var apps []app.App
	if err := a.db.WithContext(ctx).
		Where(app.App{OrgID: req.OrgID}).
		Find(&apps).Error; err != nil {
		return nil, fmt.Errorf("unable to get apps for org: %w", err)
	}

	var matches []MatchingInstallSyncApp

	for _, ap := range apps {
		var latestConfig app.AppConfig
		if err := a.db.WithContext(ctx).
			Where(app.AppConfig{AppID: ap.ID}).
			Order("created_at DESC").
			First(&latestConfig).Error; err != nil {
			continue
		}

		if latestConfig.IntermediateConfig == nil || !latestConfig.IntermediateConfig.IsSet() {
			continue
		}

		blobCtx := blobstore.WithBlobService(ctx, a.blobSvc)
		intermediateJSON, err := latestConfig.IntermediateConfig.Get(blobCtx)
		if err != nil || intermediateJSON == "" {
			continue
		}

		var parsed pkgconfig.AppConfig
		if err := json.Unmarshal([]byte(intermediateJSON), &parsed); err != nil {
			continue
		}

		if parsed.InstallsConfig == nil {
			continue
		}

		if matchesInstallsConfig(parsed.InstallsConfig, req.Repo, req.Branch) {
			matches = append(matches, MatchingInstallSyncApp{AppID: ap.ID})
		}
	}

	return matches, nil
}

func matchesInstallsConfig(ic *pkgconfig.InstallsConfig, repo, branch string) bool {
	if ic.ConnectedRepo != nil {
		if ic.ConnectedRepo.Repo == repo && ic.ConnectedRepo.Branch == branch {
			return true
		}
	}

	if ic.PublicRepo != nil {
		if ic.PublicRepo.Repo == repo && ic.PublicRepo.Branch == branch {
			return true
		}
		if ic.PublicRepo.Repo == "https://github.com/"+repo+".git" && ic.PublicRepo.Branch == branch {
			return true
		}
		if strings.TrimSuffix(ic.PublicRepo.Repo, ".git") == "https://github.com/"+repo && ic.PublicRepo.Branch == branch {
			return true
		}
	}

	return false
}

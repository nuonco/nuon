package service

import (
	"fmt"
	"strings"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/views"
)

func parseCloudPlatformsFilter(raw string) []app.CloudPlatform {
	if raw == "" {
		return nil
	}

	seen := map[app.CloudPlatform]struct{}{}
	var platforms []app.CloudPlatform
	for _, part := range strings.Split(raw, ",") {
		p := strings.TrimSpace(strings.ToLower(part))
		switch app.CloudPlatform(p) {
		case app.CloudPlatformAWS, app.CloudPlatformAzure, app.CloudPlatformGCP, app.CloudPlatformUnknown:
			plat := app.CloudPlatform(p)
			if _, ok := seen[plat]; ok {
				continue
			}
			seen[plat] = struct{}{}
			platforms = append(platforms, plat)
		}
	}
	return platforms
}

func applyCloudPlatformFilter(tx *gorm.DB, db *gorm.DB, platforms []app.CloudPlatform) *gorm.DB {
	if len(platforms) == 0 {
		return tx
	}
	if len(platforms) == 4 {
		return tx
	}

	includeUnknown := false
	var selectedTypes []string
	for _, platform := range platforms {
		if platform == app.CloudPlatformUnknown {
			includeUnknown = true
			continue
		}
		selectedTypes = append(selectedTypes, app.RunnerTypesForCloudPlatform(platform)...)
	}

	if !includeUnknown && len(selectedTypes) == 0 {
		return tx
	}

	installTable := views.TableOrViewName(db, &app.Install{}, "")
	typeExpr := fmt.Sprintf(`COALESCE(
		(SELECT type FROM app_runner_configs WHERE app_config_id = %[1]s.app_config_id AND deleted_at = 0 LIMIT 1),
		(SELECT type FROM app_runner_configs WHERE id = %[1]s.app_runner_config_id AND deleted_at = 0 LIMIT 1),
		''
	)`, installTable)

	knownTypes := append(
		append(app.RunnerTypesForCloudPlatform(app.CloudPlatformAWS), app.RunnerTypesForCloudPlatform(app.CloudPlatformAzure)...),
		app.RunnerTypesForCloudPlatform(app.CloudPlatformGCP)...,
	)

	switch {
	case includeUnknown && len(selectedTypes) > 0:
		return tx.Where(typeExpr+" IN ? OR "+typeExpr+" NOT IN ? OR "+typeExpr+" = ''", selectedTypes, knownTypes)
	case includeUnknown:
		return tx.Where(typeExpr+" NOT IN ? OR "+typeExpr+" = ''", knownTypes)
	default:
		return tx.Where(typeExpr+" IN ?", selectedTypes)
	}
}

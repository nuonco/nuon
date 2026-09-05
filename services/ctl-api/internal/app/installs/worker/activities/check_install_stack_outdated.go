package activities

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/deployerrors"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/configdiff"
)

func (a *Activities) checkInstallStackOutdated(ctx context.Context, req InstallPreflightRequest) ([]*compositeerrors.CompositeErrorData, error) {
	if !req.CheckStackOutdated {
		return nil, nil
	}

	var install app.Install
	if err := a.db.WithContext(ctx).
		Select("id", "app_config_id").
		Where(app.Install{ID: req.InstallID}).
		First(&install).Error; err != nil {
		return nil, fmt.Errorf("unable to get install: %w", err)
	}

	var versions []app.InstallStackVersion
	if err := a.db.WithContext(ctx).
		Select("id", "app_config_id", "status", "created_at").
		Where(app.InstallStackVersion{InstallID: req.InstallID}).
		Order("created_at DESC").
		Find(&versions).Error; err != nil {
		return nil, fmt.Errorf("unable to get install stack versions: %w", err)
	}

	desiredAppConfigID := req.DesiredAppConfigID
	if desiredAppConfigID == "" {
		desiredAppConfigID = install.AppConfigID
	}

	activeAppConfigID := activeStackAppConfigID(versions)
	outdated := activeAppConfigID == "" && len(versions) > 0
	if activeAppConfigID != "" && activeAppConfigID != desiredAppConfigID {
		applied, err := a.getAppStackConfig(ctx, activeAppConfigID)
		if err != nil {
			return nil, fmt.Errorf("unable to get applied stack config: %w", err)
		}
		desired, err := a.getAppStackConfig(ctx, desiredAppConfigID)
		if err != nil {
			return nil, fmt.Errorf("unable to get desired stack config: %w", err)
		}
		outdated = !configdiff.StackConfigEqual(applied, desired)
	}
	if !outdated {
		return nil, nil
	}

	finding, err := compositeerrors.New(
		&deployerrors.InstallStackOutdatedError{InstallID: req.InstallID},
		compositeerrors.WithSource((&app.Workflow{}).TableName(), req.FlowID),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to build install stack preflight warning: %w", err)
	}
	return []*compositeerrors.CompositeErrorData{finding}, nil
}

func activeStackAppConfigID(versions []app.InstallStackVersion) string {
	versionsByID := make(map[string]app.InstallStackVersion, len(versions))
	for _, version := range versions {
		versionsByID[version.ID] = version
	}

	for _, version := range versions {
		if version.Status.Status != app.InstallStackVersionStatusActive {
			continue
		}

		applied := version
		seen := map[string]struct{}{applied.ID: {}}
		for {
			appliedFromID, ok := applied.Status.Metadata["applied_from_version_id"].(string)
			if !ok || appliedFromID == "" {
				return applied.AppConfigID
			}
			if _, ok := seen[appliedFromID]; ok {
				return applied.AppConfigID
			}
			seen[appliedFromID] = struct{}{}
			next, found := versionsByID[appliedFromID]
			if !found {
				return applied.AppConfigID
			}
			applied = next
		}
	}

	return ""
}

func (a *Activities) getAppStackConfig(ctx context.Context, appConfigID string) (app.AppStackConfig, error) {
	var config app.AppStackConfig
	err := a.db.WithContext(ctx).
		Where(app.AppStackConfig{AppConfigID: appConfigID}).
		First(&config).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return app.AppStackConfig{}, nil
	}
	return config, err
}

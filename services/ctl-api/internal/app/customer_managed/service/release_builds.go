package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"

	"gorm.io/gorm"

	customermanaged "github.com/nuonco/nuon/pkg/customer_managed"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	customermanagedapp "github.com/nuonco/nuon/services/ctl-api/internal/app/customer_managed"
)

type preconditionError struct{ msg string }

func (e preconditionError) Error() string { return e.msg }

func canonicalBundleRunbooks(runbooks []customermanaged.RunbookTemplate) ([]customermanaged.RunbookTemplate, string, error) {
	if len(runbooks) == 0 {
		return nil, "", nil
	}
	canonical := append([]customermanaged.RunbookTemplate(nil), runbooks...)
	sort.Slice(canonical, func(i, j int) bool {
		if canonical[i].ID == canonical[j].ID {
			return canonical[i].Name < canonical[j].Name
		}
		return canonical[i].ID < canonical[j].ID
	})
	raw, err := json.Marshal(canonical)
	if err != nil {
		return nil, "", err
	}
	digest := sha256.Sum256(raw)
	return canonical, hex.EncodeToString(digest[:]), nil
}

type bundleBuildSelection struct {
	sandboxBuildID    string
	componentBuildIDs map[string]string
}

func (s *service) resolveActiveBuilds(ctx context.Context, orgID, appID string, cfg *app.AppConfig) (bundleBuildSelection, error) {
	selection := bundleBuildSelection{componentBuildIDs: make(map[string]string, len(cfg.ComponentConfigConnections))}
	for _, connection := range cfg.ComponentConfigConnections {
		var build app.ComponentBuild
		var err error
		if connection.LatestBuildID.Valid {
			build, err = s.activeBuildForConnection(ctx, orgID, connection.LatestBuildID.String, connection)
		} else {
			err = s.db.WithContext(ctx).Where(app.ComponentBuild{OrgID: orgID, ComponentConfigConnectionID: connection.ID, Status: app.ComponentBuildStatusActive}).Order("created_at DESC").First(&build).Error
		}
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return selection, preconditionError{msg: fmt.Sprintf("no active build for component %s in app config %s", connection.ComponentName, cfg.ID)}
			}
			return selection, err
		}
		selection.componentBuildIDs[connection.ID] = build.ID
	}
	sandboxBuild, found, err := s.resolveReleasedSandboxBuild(ctx, orgID, appID, cfg.SandboxConfig)
	if err != nil {
		return selection, err
	}
	if !found {
		err = s.db.WithContext(ctx).Where(app.AppSandboxBuild{OrgID: orgID, AppID: appID, AppConfigID: cfg.ID, AppSandboxConfigID: cfg.SandboxConfig.ID, Status: app.AppSandboxBuildStatusActive}).Order("created_at DESC").First(&sandboxBuild).Error
	}
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return selection, preconditionError{msg: fmt.Sprintf("no active sandbox build for app config %s: run a sandbox build for this config before creating a release", cfg.ID)}
		}
		return selection, err
	}
	selection.sandboxBuildID = sandboxBuild.ID
	return selection, nil
}

func (s *service) activeBuildForConnection(ctx context.Context, orgID, buildID string, connection app.ComponentConfigConnection) (app.ComponentBuild, error) {
	var build app.ComponentBuild
	if err := s.db.WithContext(ctx).
		Where(app.ComponentBuild{ID: buildID, OrgID: orgID, Status: app.ComponentBuildStatusActive}).
		First(&build).Error; err != nil {
		return build, err
	}
	var buildComponentID string
	if err := s.db.WithContext(ctx).
		Model(&app.ComponentConfigConnection{}).
		Where(app.ComponentConfigConnection{ID: build.ComponentConfigConnectionID, OrgID: orgID}).
		Pluck("component_id", &buildComponentID).Error; err != nil {
		return build, err
	}
	if buildComponentID != connection.ComponentID {
		return build, gorm.ErrRecordNotFound
	}
	return build, nil
}

func (s *service) resolveReleasedSandboxBuild(ctx context.Context, orgID, appID string, sandboxConfig app.AppSandboxConfig) (app.AppSandboxBuild, bool, error) {
	definition, err := customermanagedapp.CanonicalObject(sandboxConfig)
	if err != nil {
		return app.AppSandboxBuild{}, false, fmt.Errorf("canonicalize sandbox: %w", err)
	}
	var members []app.AppReleaseMember
	if err := s.db.WithContext(ctx).
		Preload("Release").
		Where(app.AppReleaseMember{OrgID: orgID, Kind: "sandbox", ConfigDigest: customermanagedapp.ObjectDigest(definition)}).
		Find(&members).Error; err != nil {
		return app.AppSandboxBuild{}, false, err
	}
	sort.Slice(members, func(i, j int) bool {
		return members[i].Release.CreatedAt.After(members[j].Release.CreatedAt)
	})
	for _, member := range members {
		if member.Release.AppID != appID || member.Release.Status != app.AppReleaseStatusReady {
			continue
		}
		var build app.AppSandboxBuild
		if err := s.db.WithContext(ctx).Where(app.AppSandboxBuild{
			ID: member.BuildID, OrgID: orgID, AppID: appID, Status: app.AppSandboxBuildStatusActive,
		}).First(&build).Error; err == nil {
			return build, true, nil
		} else if err != gorm.ErrRecordNotFound {
			return app.AppSandboxBuild{}, false, err
		}
	}
	return app.AppSandboxBuild{}, false, nil
}

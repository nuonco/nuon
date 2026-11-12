package alignment

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/config/refs"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/configdiff"
)

func unscoped(db *gorm.DB) *gorm.DB {
	return db.Unscoped()
}

func (s *Helpers) loadDependencyState(ctx context.Context, st *installAlignmentState) error {
	st.appliedBuilds = map[string]*app.ComponentBuild{}
	for _, ic := range st.components {
		if deploy := st.appliedDeploys[ic.ID]; deploy != nil {
			st.appliedBuilds[ic.ComponentID] = deploy.ComponentBuild
		}
	}

	var buildIDs, sandboxBuildIDs []string
	for _, deploy := range st.appliedDeploys {
		for id, buildID := range deploy.DependencyBuildIDs {
			switch {
			case buildID == "":
			case id == "sandbox":
				sandboxBuildIDs = append(sandboxBuildIDs, buildID)
			default:
				buildIDs = append(buildIDs, buildID)
			}
		}
	}

	st.consumedBuilds = map[string]*app.ComponentBuild{}
	if len(buildIDs) > 0 {
		var builds []*app.ComponentBuild
		if err := s.db.WithContext(ctx).Unscoped().
			Preload("ComponentConfigConnection", unscoped).
			Where("id IN ?", buildIDs).
			Find(&builds).Error; err != nil {
			return fmt.Errorf("unable to get consumed component builds: %w", err)
		}
		for _, build := range builds {
			st.consumedBuilds[build.ID] = build
		}
	}

	appliedSandboxBuildID := ""
	if st.appliedSandbox != nil {
		id, err := s.installs.AppliedSandboxBuildID(ctx, st.appliedSandbox)
		if err != nil {
			return fmt.Errorf("unable to get applied sandbox build: %w", err)
		}
		appliedSandboxBuildID = id
		if id != "" {
			sandboxBuildIDs = append(sandboxBuildIDs, id)
		}
	}
	st.consumedSandboxBuilds = map[string]*app.AppSandboxBuild{}
	if len(sandboxBuildIDs) > 0 {
		var builds []*app.AppSandboxBuild
		if err := s.db.WithContext(ctx).Unscoped().
			Preload("AppSandboxConfig", unscoped).
			Preload("AppSandboxConfig.ConnectedGithubVCSConfig", unscoped).
			Preload("AppSandboxConfig.PublicGitVCSConfig", unscoped).
			Where("id IN ?", sandboxBuildIDs).
			Find(&builds).Error; err != nil {
			return fmt.Errorf("unable to get consumed sandbox builds: %w", err)
		}
		for _, build := range builds {
			st.consumedSandboxBuilds[build.ID] = build
		}
	}
	if st.appliedSandbox != nil {
		st.appliedSandboxBuild = st.consumedSandboxBuilds[appliedSandboxBuildID]
		if st.appliedSandboxBuild == nil {
			st.appliedSandboxBuild = &app.AppSandboxBuild{ID: appliedSandboxBuildID}
		}
	}

	if err := s.loadSandboxChanged(ctx, st); err != nil {
		return err
	}

	if st.stack != nil && st.install.AppConfigID != "" {
		applied := st.stack.AppConfigRef.AppliedConfigID
		if applied != "" && applied != st.install.AppConfigID {
			st.stackChanged, st.stackErr = configdiff.StackChanged(blobstore.WithBlobService(ctx, s.blobSvc), s.db, applied, st.install.AppConfigID)
			if st.stackErr != nil {
				s.l.Warn("unable to compare install stack config",
					zap.String("install_id", st.install.ID),
					zap.Error(st.stackErr))
			}
		}
	}
	return nil
}

func (s *Helpers) loadSandboxChanged(ctx context.Context, st *installAlignmentState) error {
	if st.appliedSandbox == nil || st.expectedSandbox == nil {
		return nil
	}
	var cfgs []app.AppSandboxConfig
	if err := s.db.WithContext(ctx).Unscoped().
		Preload("ConnectedGithubVCSConfig", unscoped).
		Preload("PublicGitVCSConfig", unscoped).
		Where("id IN ?", []string{st.appliedSandbox.AppSandboxConfigID, st.expectedSandbox.ID}).
		Find(&cfgs).Error; err != nil {
		return fmt.Errorf("unable to get sandbox configs: %w", err)
	}
	var applied, expected *app.AppSandboxConfig
	for i := range cfgs {
		if cfgs[i].ID == st.appliedSandbox.AppSandboxConfigID {
			applied = &cfgs[i]
		}
		if cfgs[i].ID == st.expectedSandbox.ID {
			expected = &cfgs[i]
		}
	}
	st.appliedSandboxConfig = applied

	next, err := configdiff.LatestActiveSandboxBuild(ctx, s.db, st.install.AppConfigID)
	if err != nil {
		return fmt.Errorf("unable to get selected sandbox build: %w", err)
	}
	configChanged := applied == nil || expected == nil ||
		(applied.ID != expected.ID && !configdiff.SandboxConfigsEqual(*applied, *expected))
	st.sandboxChanged = configChanged || !configdiff.SandboxSourcesEqual(st.appliedSandboxBuild, next)
	return nil
}

func (st *installAlignmentState) dependenciesAppliedSince(componentID string, conn *app.ComponentConfigConnection, deploy *app.InstallDeploy) (changed, verified bool) {
	deps, sandbox := st.directDependencies(componentID, conn)
	if len(deps) == 0 && !sandbox {
		return false, true
	}
	receipt := deploy.DependencyBuildIDs
	if receipt == nil {
		return false, false
	}
	for _, id := range deps {
		if configdiff.DependencyAppliedSince(st.consumedBuilds[receipt[id]], st.appliedBuilds[id]) {
			return true, true
		}
	}
	if sandbox && configdiff.SandboxAppliedSince(st.consumedSandboxBuilds[receipt["sandbox"]], st.appliedSandboxBuild, st.appliedSandboxConfig) {
		return true, true
	}
	return false, true
}

func (st *installAlignmentState) directDependencies(componentID string, conn *app.ComponentConfigConnection) (componentIDs []string, sandbox bool) {
	if conn == nil {
		return nil, false
	}
	idsByName := make(map[string]string, len(st.targets))
	for id, target := range st.targets {
		if target.Config != nil && target.Config.Component.Name != "" {
			idsByName[target.Config.Component.Name] = id
		}
	}
	seen := map[string]bool{componentID: true}
	add := func(id string) {
		if id != "" && !seen[id] {
			seen[id] = true
			componentIDs = append(componentIDs, id)
		}
	}
	for _, id := range conn.ComponentDependencyIDs {
		add(id)
	}
	for _, ref := range conn.Refs {
		switch ref.Type {
		case refs.RefTypeComponents:
			add(idsByName[ref.Name])
		case refs.RefTypeSandbox:
			sandbox = true
		}
	}
	return componentIDs, sandbox
}

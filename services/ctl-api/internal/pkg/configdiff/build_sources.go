package configdiff

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/config"
	pkgdiff "github.com/nuonco/nuon/pkg/config/diff"
	"github.com/nuonco/nuon/pkg/config/refs"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func ComponentSourcesEqual(old, next *app.ComponentBuild) bool {
	if old == nil || next == nil {
		return false
	}
	if old.ID != "" && old.ID == next.ID {
		return true
	}
	if old.SourceChecksum != "" && old.SourceChecksum == next.SourceChecksum &&
		old.SourceIdentity.EligibleChecksum() == old.SourceChecksum && next.SourceIdentity.EligibleChecksum() == next.SourceChecksum {
		return true
	}
	return old.SourceIdentity == nil && next.SourceIdentity == nil && old.SourceDigest != "" && strings.HasPrefix(old.SourceDigest, "sha256:") &&
		old.SourceDigest == next.SourceDigest
}

func SandboxSourcesEqual(old, next *app.AppSandboxBuild) bool {
	return old != nil && next != nil &&
		((old.ID != "" && old.ID == next.ID) ||
			(old.SourceChecksum != "" && old.SourceChecksum == next.SourceChecksum &&
				old.SourceIdentity.EligibleChecksum() == old.SourceChecksum && next.SourceIdentity.EligibleChecksum() == next.SourceChecksum))
}

type configBuildSources struct {
	components map[string]*app.ComponentBuild
	sandbox    *app.AppSandboxBuild
}

func loadConfigBuildSources(ctx context.Context, db *gorm.DB, cfg *app.AppConfig) (*configBuildSources, error) {
	sources := &configBuildSources{components: map[string]*app.ComponentBuild{}}
	if cfg == nil {
		return sources, nil
	}
	ids := make([]string, 0, len(cfg.ComponentConfigConnections))
	for _, conn := range cfg.ComponentConfigConnections {
		if id := cccBuildID(&conn); id != "" {
			ids = append(ids, id)
			sources.components[conn.ComponentID] = &app.ComponentBuild{ID: id}
		}
	}
	if len(ids) > 0 {
		var builds []*app.ComponentBuild
		if err := db.WithContext(ctx).Unscoped().Where("id IN ?", ids).Find(&builds).Error; err != nil {
			return nil, fmt.Errorf("unable to get component source identities: %w", err)
		}
		byID := make(map[string]*app.ComponentBuild, len(builds))
		for i := range builds {
			byID[builds[i].ID] = builds[i]
		}
		for componentID, build := range sources.components {
			if loaded := byID[build.ID]; loaded != nil {
				sources.components[componentID] = loaded
			}
		}
	}
	var err error
	sources.sandbox, err = LatestActiveSandboxBuild(ctx, db, cfg.ID)
	return sources, err
}

func applySourceChanges(ctx context.Context, db *gorm.DB, diff *app.InstallConfigDiff, baselines CompositeBaselines, loaded map[string]*app.AppConfig, next *app.AppConfig) error {
	var dependencyBuildIDs, sandboxBuildIDs []string
	for _, receipt := range baselines.DependencyBuildIDs {
		for id, buildID := range receipt {
			if buildID == "" {
				continue
			}
			if id == "sandbox" {
				sandboxBuildIDs = append(sandboxBuildIDs, buildID)
			} else {
				dependencyBuildIDs = append(dependencyBuildIDs, buildID)
			}
		}
	}
	dependencyBuilds := map[string]*app.ComponentBuild{}
	if len(dependencyBuildIDs) > 0 {
		var builds []*app.ComponentBuild
		if err := db.WithContext(ctx).Unscoped().Where("id IN ?", dependencyBuildIDs).Find(&builds).Error; err != nil {
			return fmt.Errorf("unable to get consumed component sources: %w", err)
		}
		for i := range builds {
			dependencyBuilds[builds[i].ID] = builds[i]
		}
	}
	sandboxBuilds := map[string]*app.AppSandboxBuild{}
	if len(sandboxBuildIDs) > 0 {
		var builds []*app.AppSandboxBuild
		if err := db.WithContext(ctx).Unscoped().Where("id IN ?", sandboxBuildIDs).Find(&builds).Error; err != nil {
			return fmt.Errorf("unable to get consumed sandbox sources: %w", err)
		}
		for i := range builds {
			sandboxBuilds[builds[i].ID] = builds[i]
		}
	}
	cache := map[string]*configBuildSources{}
	get := func(id string) (*configBuildSources, error) {
		if sources, ok := cache[id]; ok {
			return sources, nil
		}
		sources, err := loadConfigBuildSources(ctx, db, loaded[id])
		if err == nil {
			cache[id] = sources
		}
		return sources, err
	}
	nextSources, err := get(next.ID)
	if err != nil {
		return err
	}
	oldSandbox, err := get(baselines.Sandbox)
	if err != nil {
		return err
	}
	if baselines.SandboxBuildID != nil {
		oldSandbox = &configBuildSources{sandbox: &app.AppSandboxBuild{ID: *baselines.SandboxBuildID}}
		if *baselines.SandboxBuildID != "" {
			if err := db.WithContext(ctx).Unscoped().Where(app.AppSandboxBuild{ID: *baselines.SandboxBuildID}).First(oldSandbox.sandbox).Error; err != nil {
				return fmt.Errorf("unable to get applied sandbox source: %w", err)
			}
		}
	}
	if sandboxSourceChanged(oldSandbox.sandbox, nextSources.sandbox) {
		diff.SandboxChanged = true
		diff.SandboxNewID = next.SandboxConfig.ID
		if old := loaded[baselines.Sandbox]; old != nil {
			diff.SandboxOldID = old.SandboxConfig.ID
		}
	}
	diff.SandboxBuildOldID = ""
	diff.SandboxBuildNewID = ""
	if oldSandbox.sandbox != nil {
		diff.SandboxBuildOldID = oldSandbox.sandbox.ID
	}
	if nextSources.sandbox != nil {
		diff.SandboxBuildNewID = nextSources.sandbox.ID
	}
	diff.SandboxBuildChanged = diff.SandboxBuildOldID != diff.SandboxBuildNewID
	graphs := map[string]*pkgdiff.Diff{}
	fallbackImpacts := map[string]map[pkgdiff.NodeID][]pkgdiff.ImpactReason{}
	changed := append([]app.ComponentDiffEntry(nil), diff.Changed...)
	unchanged := make([]app.ComponentDiffEntry, 0, len(diff.Unchanged))
	entries := append(changed, diff.Unchanged...)
	diff.Changed = nil
	for _, entry := range entries {
		baselineID, ok := baselines.Components[entry.ComponentID]
		if !ok {
			baselineID = baselines.Fallback
		}
		oldSources, err := get(baselineID)
		if err != nil {
			return err
		}
		oldBuild := oldSources.components[entry.ComponentID]
		if applied, ok := baselines.ComponentBuilds[entry.ComponentID]; ok {
			oldBuild = applied
		}
		graphKey := baselineID
		receipt := baselines.DependencyBuildIDs[entry.ComponentID]
		oldCfg := loaded[baselineID]
		if receipt != nil || baselines.ComponentBuilds != nil {
			graphKey += ":" + entry.ComponentID
			oldSources = &configBuildSources{components: map[string]*app.ComponentBuild{}, sandbox: sandboxBuilds[receipt["sandbox"]]}
			for _, conn := range next.ComponentConfigConnections {
				oldSources.components[conn.ComponentID] = dependencyBuilds[receipt[conn.ComponentID]]
			}
			oldSources.components[entry.ComponentID] = oldBuild
		}
		entry.OldBuildID = ""
		if oldBuild != nil {
			entry.OldBuildID = oldBuild.ID
		}
		entry.BuildChanged = entry.OldBuildID != entry.NewBuildID && (entry.OldBuildID != "" || entry.NewBuildID != "")
		graph, ok := graphs[graphKey]
		if !ok {
			roots := sourceChangedRoots(oldSources, nextSources, next)
			for _, conn := range next.ComponentConfigConnections {
				if oldConn := connOn(oldCfg, conn.ComponentID); oldConn == nil || !checksumsEqual(oldConn, &conn) {
					roots = append(roots, config.ComponentResourceID(conn.ComponentName))
				}
			}
			if oldCfg == nil || !sandboxConfigEqual(oldCfg.SandboxConfig, next.SandboxConfig) {
				roots = append(roots, config.SandboxResourceID)
			}
			graph, err = intermediateConfigDiff(ctx, loaded[baselineID], next, roots...)
			if err != nil {
				return err
			}
			if graph == nil {
				fallbackImpacts[graphKey] = persistedDependencyImpacts(oldCfg, next, roots)
			}
			graphs[graphKey] = graph
		}
		if graph != nil {
			if node := graph.FindResource(config.ComponentResourceID(entry.ComponentName)); node != nil && node.Impacted {
				entry.ImpactReasons = node.ImpactReasons
			}
		} else if reasons := fallbackImpacts[graphKey][config.ComponentResourceID(entry.ComponentName)]; len(reasons) > 0 {
			entry.ImpactReasons = reasons
		}
		nextBuild := nextSources.components[entry.ComponentID]
		if entry.OldChecksum != "" && entry.OldChecksum == entry.NewChecksum && len(entry.ImpactReasons) == 0 &&
			!componentSourceChanged(oldBuild, nextBuild) {
			unchanged = append(unchanged, entry)
		} else {
			diff.Changed = append(diff.Changed, entry)
		}
	}
	diff.Unchanged = unchanged
	return nil
}

func componentSourceChanged(old, next *app.ComponentBuild) bool {
	return !ComponentSourcesEqual(old, next)
}

func sandboxSourceChanged(old, next *app.AppSandboxBuild) bool {
	return !SandboxSourcesEqual(old, next)
}

func sourceChangedRoots(old, next *configBuildSources, cfg *app.AppConfig) []pkgdiff.NodeID {
	var roots []pkgdiff.NodeID
	for _, conn := range cfg.ComponentConfigConnections {
		if componentSourceChanged(old.components[conn.ComponentID], next.components[conn.ComponentID]) {
			roots = append(roots, config.ComponentResourceID(conn.ComponentName))
		}
	}
	if sandboxSourceChanged(old.sandbox, next.sandbox) {
		roots = append(roots, config.SandboxResourceID)
	}
	return roots
}

func persistedDependencyImpacts(old, next *app.AppConfig, roots []pkgdiff.NodeID) map[pkgdiff.NodeID][]pkgdiff.ImpactReason {
	graph := pkgdiff.NewGraph()
	for _, cfg := range []*app.AppConfig{old, next} {
		if cfg == nil {
			continue
		}
		byID := connsByComponentID(cfg)
		for _, conn := range cfg.ComponentConfigConnections {
			dependent := config.ComponentResourceID(conn.ComponentName)
			for _, id := range conn.ComponentDependencyIDs {
				if dependency := byID[id]; dependency != nil {
					graph.AddDependency(dependent, config.ComponentResourceID(dependency.ComponentName), pkgdiff.EdgeReasonComponentDependency)
				}
			}
			for _, raw := range conn.References {
				ref := refs.NewFromString(raw)
				switch ref.Type {
				case refs.RefTypeComponents:
					graph.AddDependency(dependent, config.ComponentResourceID(ref.Name), pkgdiff.EdgeReasonComponentReference)
				case refs.RefTypeSandbox:
					graph.AddDependency(dependent, config.SandboxResourceID, pkgdiff.EdgeReasonSandboxOutput)
				}
			}
		}
	}
	return graph.Propagate(roots...)
}

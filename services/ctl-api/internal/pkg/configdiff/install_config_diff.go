// Package configdiff computes the difference between two app config versions
// as it applies to a single install.
package configdiff

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/config"
	pkgdiff "github.com/nuonco/nuon/pkg/config/diff"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func preload(db *gorm.DB) *gorm.DB {
	return db.
		Preload("ComponentConfigConnections").
		Preload("ComponentConfigConnections.Component").
		Preload("PermissionsConfig").
		Preload("PermissionsConfig.Roles").
		Preload("PermissionsConfig.Roles.Policies").
		Preload("PermissionsConfig.NamedPolicies").
		Preload("BreakGlassConfig").
		Preload("BreakGlassConfig.Roles").
		Preload("BreakGlassConfig.Roles.Policies").
		Preload("SecretsConfig").
		Preload("SecretsConfig.Secrets").
		Preload("SecretsConfig.Secrets.KubernetesSyncTargets").
		Preload("SandboxConfig").
		Preload("SandboxConfig.PublicGitVCSConfig").
		Preload("SandboxConfig.ConnectedGithubVCSConfig").
		Preload("RunnerConfig").
		Preload("StackConfig")
}

// ComputeInstallConfigDiff diffs oldAppConfigID against newAppConfigID. An empty
// oldAppConfigID means the install has never been pinned to a config, so
// everything counts as added.
func ComputeInstallConfigDiff(ctx context.Context, db *gorm.DB, oldAppConfigID, newAppConfigID string) (*app.InstallConfigDiff, error) {
	var newAppCfg app.AppConfig
	if err := preload(db.WithContext(ctx)).First(&newAppCfg, "id = ?", newAppConfigID).Error; err != nil {
		return nil, fmt.Errorf("unable to get new app config: %w", err)
	}

	diff := &app.InstallConfigDiff{
		Added:     []app.ComponentDiffEntry{},
		Removed:   []app.ComponentDiffEntry{},
		Changed:   []app.ComponentDiffEntry{},
		Unchanged: []app.ComponentDiffEntry{},
	}

	newConnByComponent := make(map[string]*app.ComponentConfigConnection, len(newAppCfg.ComponentConfigConnections))
	for i := range newAppCfg.ComponentConfigConnections {
		ccc := &newAppCfg.ComponentConfigConnections[i]
		newConnByComponent[ccc.ComponentID] = ccc
	}

	if oldAppConfigID != "" {
		var oldAppCfg app.AppConfig
		err := preload(db.WithContext(ctx)).First(&oldAppCfg, "id = ?", oldAppConfigID).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("unable to get old app config: %w", err)
		}
		if err == nil {
			graphDiff, err := intermediateConfigDiff(ctx, &oldAppCfg, &newAppCfg)
			if err != nil {
				return nil, err
			}
			oldConnByComponent := make(map[string]*app.ComponentConfigConnection, len(oldAppCfg.ComponentConfigConnections))
			for i := range oldAppCfg.ComponentConfigConnections {
				ccc := &oldAppCfg.ComponentConfigConnections[i]
				oldConnByComponent[ccc.ComponentID] = ccc
			}

			for componentID, oldConn := range oldConnByComponent {
				newConn, exists := newConnByComponent[componentID]
				if !exists {
					diff.Removed = append(diff.Removed, app.ComponentDiffEntry{
						ComponentID:   componentID,
						ComponentName: oldConn.ComponentName,
						ComponentType: string(oldConn.Type),
						OldChecksum:   oldConn.Checksum,
					})
					continue
				}

				entry := componentDiffEntry(oldConn, newConn)
				if graphDiff != nil {
					if component := graphDiff.FindResource(config.ComponentResourceID(entry.ComponentName)); component != nil && component.Impacted {
						entry.ImpactReasons = append([]pkgdiff.ImpactReason(nil), component.ImpactReasons...)
					}
				}
				if checksumsEqual(oldConn, newConn) {
					if entry.BuildChanged || len(entry.ImpactReasons) > 0 {
						diff.Changed = append(diff.Changed, entry)
					} else {
						diff.Unchanged = append(diff.Unchanged, entry)
					}
				} else {
					diff.Changed = append(diff.Changed, entry)
				}

				delete(newConnByComponent, componentID)
			}

			for _, newConn := range newConnByComponent {
				entry := componentDiffEntry(nil, newConn)
				diff.Added = append(diff.Added, entry)
			}

			// Every sync writes fresh sandbox and stack config rows, so their IDs
			// always differ between versions — only a content change is a real change.
			if oldAppCfg.SandboxConfig.ID != newAppCfg.SandboxConfig.ID &&
				!sandboxConfigEqual(oldAppCfg.SandboxConfig, newAppCfg.SandboxConfig) {
				diff.SandboxChanged = true
				diff.SandboxOldID = oldAppCfg.SandboxConfig.ID
				diff.SandboxNewID = newAppCfg.SandboxConfig.ID
			}
			if graphDiff != nil {
				stack := graphDiff.FindResource(config.StackResourceID)
				if stack != nil && stack.Summary().HasChanged {
					diff.StackImpacts = graphStackImpacts(stack, &oldAppCfg, &newAppCfg)
					diff.StackImpactReasons = append([]pkgdiff.ImpactReason(nil), stack.ImpactReasons...)
				}
			} else {
				// Historical configs without intermediate blobs retain the
				// database projection fallback.
				diff.StackImpacts = stackImpactChanges(&oldAppCfg, &newAppCfg)
			}
			if len(diff.StackImpacts) > 0 {
				diff.StackChanged = true
				diff.StackOldID = oldAppCfg.StackConfig.ID
				diff.StackNewID = newAppCfg.StackConfig.ID
			}

			oldSandboxBuildID, err := latestActiveSandboxBuildID(ctx, db, oldAppConfigID)
			if err != nil {
				return nil, err
			}
			newSandboxBuildID, err := latestActiveSandboxBuildID(ctx, db, newAppConfigID)
			if err != nil {
				return nil, err
			}
			if oldSandboxBuildID != newSandboxBuildID {
				diff.SandboxBuildChanged = true
				diff.SandboxBuildOldID = oldSandboxBuildID
				diff.SandboxBuildNewID = newSandboxBuildID
			}

			return diff, nil
		}
	}

	for _, newConn := range newConnByComponent {
		entry := componentDiffEntry(nil, newConn)
		diff.Added = append(diff.Added, entry)
	}
	if newAppCfg.SandboxConfig.ID != "" {
		diff.SandboxChanged = true
		diff.SandboxNewID = newAppCfg.SandboxConfig.ID
	}
	if newAppCfg.StackConfig.ID != "" {
		diff.StackChanged = true
		diff.StackNewID = newAppCfg.StackConfig.ID
		diff.StackImpacts = stackImpactChanges(nil, &newAppCfg)
	}
	if newSandboxBuildID, err := latestActiveSandboxBuildID(ctx, db, newAppConfigID); err != nil {
		return nil, err
	} else if newSandboxBuildID != "" {
		diff.SandboxBuildChanged = true
		diff.SandboxBuildNewID = newSandboxBuildID
	}

	return diff, nil
}

// CompositeBaselines is the applied app config for each install entity. An empty
// id means that entity has never been applied. Callers resolve a missing entity
// ref to the install's deployed config before calling, so a legacy install is
// not treated as brand new.
type CompositeBaselines struct {
	Fallback   string
	Stack      string
	Sandbox    string
	Components map[string]string
}

// ComputeCompositeInstallConfigDiff unions a stack slice, a sandbox slice, and
// one slice per component. Each slice is diffed against that entity's baseline,
// not against a single install pin.
func ComputeCompositeInstallConfigDiff(ctx context.Context, db *gorm.DB, baselines CompositeBaselines, newAppConfigID string) (*app.InstallConfigDiff, error) {
	var newAppCfg app.AppConfig
	if err := preload(db.WithContext(ctx)).First(&newAppCfg, "id = ?", newAppConfigID).Error; err != nil {
		return nil, fmt.Errorf("unable to get new app config: %w", err)
	}

	loaded, err := loadBaselineConfigs(ctx, db, baselines, newAppConfigID, &newAppCfg)
	if err != nil {
		return nil, err
	}

	diff := &app.InstallConfigDiff{
		Added:     []app.ComponentDiffEntry{},
		Removed:   []app.ComponentDiffEntry{},
		Changed:   []app.ComponentDiffEntry{},
		Unchanged: []app.ComponentDiffEntry{},
	}

	graphs := map[string]*pkgdiff.Diff{}
	graphBetween := func(oldID string) (*pkgdiff.Diff, error) {
		if oldID == "" || oldID == newAppConfigID {
			return nil, nil
		}
		if graph, ok := graphs[oldID]; ok {
			return graph, nil
		}
		old := loaded[oldID]
		if old == nil {
			return nil, nil
		}
		graph, err := intermediateConfigDiff(ctx, old, &newAppCfg)
		if err != nil {
			return nil, err
		}
		graphs[oldID] = graph
		return graph, nil
	}

	stackGraph, err := graphBetween(baselines.Stack)
	if err != nil {
		return nil, err
	}
	if err := fillStackSlice(diff, loaded[baselines.Stack], &newAppCfg, stackGraph); err != nil {
		return nil, err
	}
	if err := fillSandboxSlice(ctx, db, diff, loaded[baselines.Sandbox], &newAppCfg, baselines.Sandbox, newAppConfigID); err != nil {
		return nil, err
	}

	if err := fillComponentSlices(diff, baselines, loaded, &newAppCfg, graphBetween); err != nil {
		return nil, err
	}
	sortComponentEntries(diff)
	return diff, nil
}

func loadBaselineConfigs(ctx context.Context, db *gorm.DB, baselines CompositeBaselines, newAppConfigID string, newAppCfg *app.AppConfig) (map[string]*app.AppConfig, error) {
	ids := map[string]struct{}{}
	add := func(id string) {
		if id != "" && id != newAppConfigID {
			ids[id] = struct{}{}
		}
	}
	add(baselines.Fallback)
	add(baselines.Stack)
	add(baselines.Sandbox)
	for _, id := range baselines.Components {
		add(id)
	}

	loaded := map[string]*app.AppConfig{newAppConfigID: newAppCfg}
	for id := range ids {
		var cfg app.AppConfig
		err := preload(db.WithContext(ctx)).First(&cfg, "id = ?", id).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("unable to get app config %s: %w", id, err)
		}
		copied := cfg
		loaded[id] = &copied
	}
	return loaded, nil
}

func fillStackSlice(diff *app.InstallConfigDiff, oldCfg, newCfg *app.AppConfig, graphDiff *pkgdiff.Diff) error {
	if oldCfg == nil {
		if newCfg.StackConfig.ID != "" {
			diff.StackChanged = true
			diff.StackNewID = newCfg.StackConfig.ID
			diff.StackImpacts = stackImpactChanges(nil, newCfg)
		}
		return nil
	}
	if graphDiff != nil {
		stack := graphDiff.FindResource(config.StackResourceID)
		if stack != nil && stack.Summary().HasChanged {
			diff.StackImpacts = graphStackImpacts(stack, oldCfg, newCfg)
			diff.StackImpactReasons = append([]pkgdiff.ImpactReason(nil), stack.ImpactReasons...)
		}
	} else {
		diff.StackImpacts = stackImpactChanges(oldCfg, newCfg)
	}
	if len(diff.StackImpacts) > 0 {
		diff.StackChanged = true
		diff.StackOldID = oldCfg.StackConfig.ID
		diff.StackNewID = newCfg.StackConfig.ID
	}
	return nil
}

func fillSandboxSlice(ctx context.Context, db *gorm.DB, diff *app.InstallConfigDiff, oldCfg, newCfg *app.AppConfig, oldID, newID string) error {
	if oldCfg == nil {
		if newCfg.SandboxConfig.ID != "" {
			diff.SandboxChanged = true
			diff.SandboxNewID = newCfg.SandboxConfig.ID
		}
		buildID, err := latestActiveSandboxBuildID(ctx, db, newID)
		if err != nil {
			return err
		}
		if buildID != "" {
			diff.SandboxBuildChanged = true
			diff.SandboxBuildNewID = buildID
		}
		return nil
	}
	if oldCfg.SandboxConfig.ID != newCfg.SandboxConfig.ID && !sandboxConfigEqual(oldCfg.SandboxConfig, newCfg.SandboxConfig) {
		diff.SandboxChanged = true
		diff.SandboxOldID = oldCfg.SandboxConfig.ID
		diff.SandboxNewID = newCfg.SandboxConfig.ID
	}
	oldBuildID, err := latestActiveSandboxBuildID(ctx, db, oldID)
	if err != nil {
		return err
	}
	newBuildID, err := latestActiveSandboxBuildID(ctx, db, newID)
	if err != nil {
		return err
	}
	if oldBuildID != newBuildID {
		diff.SandboxBuildChanged = true
		diff.SandboxBuildOldID = oldBuildID
		diff.SandboxBuildNewID = newBuildID
	}
	return nil
}

func fillComponentSlices(
	diff *app.InstallConfigDiff,
	baselines CompositeBaselines,
	loaded map[string]*app.AppConfig,
	newCfg *app.AppConfig,
	graphBetween func(oldID string) (*pkgdiff.Diff, error),
) error {
	newByID := connsByComponentID(newCfg)
	ids := map[string]struct{}{}
	for id := range newByID {
		ids[id] = struct{}{}
	}
	collectComponentIDs(ids, loaded[baselines.Fallback])
	seenBaseline := map[string]struct{}{}
	for _, baselineID := range baselines.Components {
		if _, ok := seenBaseline[baselineID]; ok {
			continue
		}
		seenBaseline[baselineID] = struct{}{}
		collectComponentIDs(ids, loaded[baselineID])
	}

	ordered := make([]string, 0, len(ids))
	for id := range ids {
		ordered = append(ordered, id)
	}
	sort.Strings(ordered)

	for _, id := range ordered {
		baselineID := baselines.Components[id]
		if baselineID == "" {
			baselineID = baselines.Fallback
		}
		oldConn := connOn(loaded[baselineID], id)
		if oldConn == nil && baselines.Components[id] == "" {
			oldConn = connOnAnyBaseline(loaded, baselines, id)
		}
		newConn := newByID[id]
		if oldConn == nil && newConn == nil {
			continue
		}
		if newConn == nil {
			diff.Removed = append(diff.Removed, app.ComponentDiffEntry{
				ComponentID:   id,
				ComponentName: oldConn.ComponentName,
				ComponentType: string(oldConn.Type),
				OldChecksum:   oldConn.Checksum,
			})
			continue
		}
		if oldConn == nil {
			diff.Added = append(diff.Added, componentDiffEntry(nil, newConn))
			continue
		}
		graph, err := graphBetween(baselineID)
		if err != nil {
			return err
		}
		entry := componentDiffEntry(oldConn, newConn)
		if graph != nil {
			if component := graph.FindResource(config.ComponentResourceID(entry.ComponentName)); component != nil && component.Impacted {
				entry.ImpactReasons = append([]pkgdiff.ImpactReason(nil), component.ImpactReasons...)
			}
		}
		if checksumsEqual(oldConn, newConn) && !entry.BuildChanged && len(entry.ImpactReasons) == 0 {
			diff.Unchanged = append(diff.Unchanged, entry)
			continue
		}
		diff.Changed = append(diff.Changed, entry)
	}
	return nil
}

func collectComponentIDs(ids map[string]struct{}, cfg *app.AppConfig) {
	if cfg == nil {
		return
	}
	for i := range cfg.ComponentConfigConnections {
		ids[cfg.ComponentConfigConnections[i].ComponentID] = struct{}{}
	}
}

func connsByComponentID(cfg *app.AppConfig) map[string]*app.ComponentConfigConnection {
	if cfg == nil {
		return nil
	}
	out := make(map[string]*app.ComponentConfigConnection, len(cfg.ComponentConfigConnections))
	for i := range cfg.ComponentConfigConnections {
		conn := &cfg.ComponentConfigConnections[i]
		out[conn.ComponentID] = conn
	}
	return out
}

func connOn(cfg *app.AppConfig, componentID string) *app.ComponentConfigConnection {
	if cfg == nil {
		return nil
	}
	for i := range cfg.ComponentConfigConnections {
		if cfg.ComponentConfigConnections[i].ComponentID == componentID {
			return &cfg.ComponentConfigConnections[i]
		}
	}
	return nil
}

func connOnAnyBaseline(loaded map[string]*app.AppConfig, baselines CompositeBaselines, componentID string) *app.ComponentConfigConnection {
	if conn := connOn(loaded[baselines.Fallback], componentID); conn != nil {
		return conn
	}
	for _, baselineID := range baselines.Components {
		if conn := connOn(loaded[baselineID], componentID); conn != nil {
			return conn
		}
	}
	return nil
}

func sortComponentEntries(diff *app.InstallConfigDiff) {
	less := func(entries []app.ComponentDiffEntry) {
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].ComponentID < entries[j].ComponentID
		})
	}
	less(diff.Added)
	less(diff.Removed)
	less(diff.Changed)
	less(diff.Unchanged)
}

func intermediateConfigDiff(ctx context.Context, oldCfg, newCfg *app.AppConfig) (*pkgdiff.Diff, error) {
	oldIntermediate, oldOK, err := loadIntermediateConfig(ctx, oldCfg)
	if err != nil {
		return nil, fmt.Errorf("unable to load old intermediate config: %w", err)
	}
	newIntermediate, newOK, err := loadIntermediateConfig(ctx, newCfg)
	if err != nil {
		return nil, fmt.Errorf("unable to load new intermediate config: %w", err)
	}
	if !oldOK || !newOK {
		return nil, nil
	}
	return newIntermediate.Diff(oldIntermediate), nil
}

func loadIntermediateConfig(ctx context.Context, appCfg *app.AppConfig) (*config.AppConfig, bool, error) {
	if appCfg == nil || appCfg.IntermediateConfig == nil || !appCfg.IntermediateConfig.IsSet() {
		return nil, false, nil
	}
	raw, err := appCfg.IntermediateConfig.Get(ctx)
	if err != nil {
		return nil, false, err
	}
	var cfg config.AppConfig
	decoder := json.NewDecoder(strings.NewReader(raw))
	decoder.UseNumber()
	if err := decoder.Decode(&cfg); err != nil {
		return nil, false, err
	}
	return &cfg, true, nil
}

func graphStackImpacts(stack *pkgdiff.Diff, oldCfg, newCfg *app.AppConfig) []app.InstallConfigImpact {
	var impacts []app.InstallConfigImpact
	if stack.DirectSummary().HasChanged {
		impacts = append(impacts, app.InstallConfigImpactStackConfig)
	}
	for _, reason := range stack.ImpactReasons {
		from := string(reason.From)
		switch {
		case reason.From == config.RunnerResourceID:
			impacts = appendUniqueImpact(impacts, app.InstallConfigImpactRunnerConfig)
		case strings.HasPrefix(from, "secret."):
			impacts = appendUniqueImpact(impacts, app.InstallConfigImpactSecrets)
		case strings.HasPrefix(from, "named_policy."):
			impacts = appendUniqueImpact(impacts, app.InstallConfigImpactPermissions)
		case strings.HasPrefix(from, "role."):
			impact := app.InstallConfigImpactPermissions
			if appConfigHasBreakGlassRole(oldCfg, strings.TrimPrefix(from, "role.")) ||
				appConfigHasBreakGlassRole(newCfg, strings.TrimPrefix(from, "role.")) {
				impact = app.InstallConfigImpactBreakGlass
			}
			impacts = appendUniqueImpact(impacts, impact)
		}
	}
	return impacts
}

func appConfigHasBreakGlassRole(cfg *app.AppConfig, normalizedName string) bool {
	if cfg == nil {
		return false
	}
	for _, role := range cfg.BreakGlassConfig.Roles {
		if strings.ReplaceAll(strings.TrimSpace(role.Name), " ", "") == normalizedName {
			return true
		}
	}
	return false
}

func appendUniqueImpact(impacts []app.InstallConfigImpact, impact app.InstallConfigImpact) []app.InstallConfigImpact {
	for _, existing := range impacts {
		if existing == impact {
			return impacts
		}
	}
	return append(impacts, impact)
}

func checksumsEqual(oldConn, newConn *app.ComponentConfigConnection) bool {
	return oldConn.Checksum != "" && newConn.Checksum != "" && oldConn.Checksum == newConn.Checksum
}

func cccBuildID(ccc *app.ComponentConfigConnection) string {
	if ccc == nil || !ccc.LatestBuildID.Valid {
		return ""
	}
	return ccc.LatestBuildID.String
}

func componentDiffEntry(oldConn, newConn *app.ComponentConfigConnection) app.ComponentDiffEntry {
	entry := app.ComponentDiffEntry{
		ComponentID:   newConn.ComponentID,
		ComponentName: newConn.ComponentName,
		ComponentType: string(newConn.Type),
		NewChecksum:   newConn.Checksum,
		NewBuildID:    cccBuildID(newConn),
	}
	if oldConn != nil {
		entry.OldChecksum = oldConn.Checksum
		entry.OldBuildID = cccBuildID(oldConn)
		entry.BuildChanged = checksumsEqual(oldConn, newConn) && entry.OldBuildID != entry.NewBuildID &&
			(entry.OldBuildID != "" || entry.NewBuildID != "")
	}
	return entry
}

func latestActiveSandboxBuildID(ctx context.Context, db *gorm.DB, appConfigID string) (string, error) {
	if appConfigID == "" {
		return "", nil
	}
	var build app.AppSandboxBuild
	err := db.WithContext(ctx).
		Where(app.AppSandboxBuild{
			AppConfigID: appConfigID,
			Status:      app.AppSandboxBuildStatusActive,
		}).
		Order("created_at DESC").
		First(&build).Error
	if err == nil {
		return build.ID, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", fmt.Errorf("unable to get active sandbox build for app config %s: %w", appConfigID, err)
	}

	var current app.AppSandboxConfig
	err = db.WithContext(ctx).
		Preload("ConnectedGithubVCSConfig").
		Preload("PublicGitVCSConfig").
		Where(app.AppSandboxConfig{AppConfigID: appConfigID}).
		First(&current).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("unable to load sandbox config for app config %s: %w", appConfigID, err)
	}

	var candidate app.AppSandboxBuild
	err = db.WithContext(ctx).
		Preload("AppSandboxConfig").
		Preload("AppSandboxConfig.ConnectedGithubVCSConfig").
		Preload("AppSandboxConfig.PublicGitVCSConfig").
		Where(app.AppSandboxBuild{
			AppID:  current.AppID,
			Status: app.AppSandboxBuildStatusActive,
		}).
		Order("created_at DESC").
		First(&candidate).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("unable to get latest active sandbox build for app %s: %w", current.AppID, err)
	}
	if sandboxConfigEqual(candidate.AppSandboxConfig, current) {
		return candidate.ID, nil
	}
	return "", nil
}

// sandboxContent is everything about a sandbox config that decides what gets
// deployed. Orchestration knobs (max_auto_retries, skip_noops,
// auto_approve_on_policies_passing) are deliberately excluded: changing a retry
// count should not force a reprovision. Anything that selects or renders the
// sandbox belongs here — omitting a field means a real change is read as a
// no-op and never reaches installs.
type sandboxContent struct {
	Source         sandboxSource `json:"source"`
	Variables      any           `json:"variables"`
	EnvVars        any           `json:"env_vars"`
	VariablesFiles any           `json:"variables_files"`
	References     any           `json:"references"`
	Type           string        `json:"type"`
	TerraformVer   string        `json:"terraform_version"`
	DriftSchedule  string        `json:"drift_schedule"`
	Runtime        string        `json:"runtime"`
	PulumiVersion  string        `json:"pulumi_version"`
	PulumiConfig   any           `json:"pulumi_config"`
	OperationRoles any           `json:"operation_roles"`
	AWSRegionType  string        `json:"aws_region_type"`
}

// sandboxSource identifies which code the sandbox runs. A ref, directory or
// repo bump changes the deployed infrastructure while leaving every other
// field identical.
type sandboxSource struct {
	Kind       string `json:"kind"`
	Repo       string `json:"repo"`
	Directory  string `json:"directory"`
	Branch     string `json:"branch"`
	PathFilter string `json:"path_filter"`
	Connection string `json:"connection,omitempty"`
}

func sandboxSourceOf(c app.AppSandboxConfig) sandboxSource {
	switch {
	case c.ConnectedGithubVCSConfig != nil:
		v := c.ConnectedGithubVCSConfig
		return sandboxSource{
			Kind:       "connected-github",
			Repo:       v.Repo,
			Directory:  v.Directory,
			Branch:     v.Branch,
			PathFilter: v.PathFilter,
			Connection: v.VCSConnectionID,
		}
	case c.PublicGitVCSConfig != nil:
		v := c.PublicGitVCSConfig
		return sandboxSource{
			Kind:       "public-git",
			Repo:       v.Repo,
			Directory:  v.Directory,
			Branch:     v.Branch,
			PathFilter: v.PathFilter,
		}
	default:
		return sandboxSource{Kind: "builtin"}
	}
}

func sandboxContentOf(c app.AppSandboxConfig) sandboxContent {
	return sandboxContent{
		Source:         sandboxSourceOf(c),
		Variables:      c.Variables,
		EnvVars:        c.EnvVars,
		VariablesFiles: c.VariablesFiles,
		References:     c.References,
		Type:           c.Type,
		TerraformVer:   c.TerraformVersion,
		DriftSchedule:  c.DriftSchedule,
		Runtime:        c.Runtime,
		PulumiVersion:  c.PulumiVersion,
		PulumiConfig:   c.PulumiConfig,
		OperationRoles: c.OperationRoles,
		AWSRegionType:  c.AWSRegionType.String,
	}
}

func sandboxConfigEqual(a, b app.AppSandboxConfig) bool {
	return contentHashEqual(sandboxContentOf(a), sandboxContentOf(b))
}

func SandboxConfigsEqual(a, b app.AppSandboxConfig) bool {
	return sandboxConfigEqual(a, b)
}

func stackConfigContent(c app.AppStackConfig) any {
	type content struct {
		Type                    string `json:"type"`
		Name                    string `json:"name"`
		Description             string `json:"description"`
		RunnerNestedTemplateURL string `json:"runner_nested_template_url"`
		VPCNestedTemplateURL    string `json:"vpc_nested_template_url"`
		DeploymentScope         string `json:"deployment_scope"`
		CustomNestedStacks      any    `json:"custom_nested_stacks"`
	}
	return content{
		Type:                    string(c.Type),
		Name:                    c.Name,
		Description:             c.Description,
		RunnerNestedTemplateURL: c.RunnerNestedTemplateURL,
		VPCNestedTemplateURL:    c.VPCNestedTemplateURL,
		DeploymentScope:         string(c.DeploymentScope),
		CustomNestedStacks:      c.CustomNestedStacks,
	}
}

func stackConfigEqual(a, b app.AppStackConfig) bool {
	return contentHashEqual(stackConfigContent(a), stackConfigContent(b))
}

type iamPolicyContent struct {
	ManagedPolicyName string   `json:"managed_policy_name"`
	Name              string   `json:"name"`
	Contents          []byte   `json:"contents"`
	GCPPermissions    []string `json:"gcp_permissions"`
	GCPPredefinedRole string   `json:"gcp_predefined_role"`
	AzureActions      []string `json:"azure_actions"`
	AzureBuiltInRoles []string `json:"azure_built_in_roles"`
}

type iamRoleContent struct {
	CloudPlatform           string             `json:"cloud_platform"`
	Type                    app.AWSIAMRoleType `json:"type"`
	Name                    string             `json:"name"`
	Description             string             `json:"description"`
	DisplayName             string             `json:"display_name"`
	EnabledInStack          *bool              `json:"enabled_in_stack,omitempty"`
	PermissionsBoundaryJSON []byte             `json:"permissions_boundary"`
	Policies                []iamPolicyContent `json:"policies"`
}

func roleContents(roles []app.AppAWSIAMRoleConfig) []iamRoleContent {
	out := make([]iamRoleContent, 0, len(roles))
	for _, role := range roles {
		var enabled *bool
		if role.EnabledInStack.Valid {
			value := role.EnabledInStack.Bool
			enabled = &value
		}
		policies := make([]iamPolicyContent, 0, len(role.Policies))
		for _, policy := range role.Policies {
			policies = append(policies, iamPolicyContent{
				ManagedPolicyName: policy.ManagedPolicyName,
				Name:              policy.Name,
				Contents:          policy.Contents,
				GCPPermissions:    policy.GCPPermissions,
				GCPPredefinedRole: policy.GCPPredefinedRole,
				AzureActions:      policy.AzureActions,
				AzureBuiltInRoles: policy.AzureBuiltInRoles,
			})
		}
		sort.Slice(policies, func(i, j int) bool {
			return policies[i].Name+policies[i].ManagedPolicyName < policies[j].Name+policies[j].ManagedPolicyName
		})
		out = append(out, iamRoleContent{
			CloudPlatform:           role.CloudPlatform,
			Type:                    role.Type,
			Name:                    role.Name,
			Description:             role.Description,
			DisplayName:             role.DisplayName,
			EnabledInStack:          enabled,
			PermissionsBoundaryJSON: role.PermissionsBoundaryJSON,
			Policies:                policies,
		})
	}
	sort.Slice(out, func(i, j int) bool {
		return string(out[i].Type)+out[i].Name < string(out[j].Type)+out[j].Name
	})
	return out
}

type namedPolicyContent struct {
	Name        string `json:"name"`
	PolicyName  string `json:"policy_name"`
	Description string `json:"description"`
	Contents    []byte `json:"contents"`
}

func namedPolicyContents(policies []app.AppNamedIAMPolicyConfig) []namedPolicyContent {
	out := make([]namedPolicyContent, 0, len(policies))
	for _, policy := range policies {
		out = append(out, namedPolicyContent{
			Name:        policy.Name,
			PolicyName:  policy.PolicyName,
			Description: policy.Description,
			Contents:    policy.Contents,
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

type secretSyncTargetContent struct {
	Namespaces []string `json:"namespaces"`
	Name       string   `json:"name"`
	Key        string   `json:"key"`
}

type secretContent struct {
	Name                      string                    `json:"name"`
	DisplayName               string                    `json:"display_name"`
	Description               string                    `json:"description"`
	Required                  bool                      `json:"required"`
	AutoGenerate              bool                      `json:"auto_generate"`
	Format                    app.AppSecretConfigFmt    `json:"format"`
	Default                   string                    `json:"default"`
	KubernetesSync            bool                      `json:"kubernetes_sync"`
	KubernetesSecretNamespace string                    `json:"kubernetes_secret_namespace"`
	KubernetesSecretName      string                    `json:"kubernetes_secret_name"`
	KubernetesSyncTargets     []secretSyncTargetContent `json:"kubernetes_sync_targets"`
}

func secretContents(secrets []app.AppSecretConfig) []secretContent {
	out := make([]secretContent, 0, len(secrets))
	for _, secret := range secrets {
		targets := make([]secretSyncTargetContent, 0, len(secret.KubernetesSyncTargets))
		for _, target := range secret.KubernetesSyncTargets {
			namespaces := append([]string(nil), target.Namespaces...)
			sort.Strings(namespaces)
			targets = append(targets, secretSyncTargetContent{
				Namespaces: namespaces,
				Name:       target.Name,
				Key:        target.Key,
			})
		}
		sort.Slice(targets, func(i, j int) bool {
			return targets[i].Name+targets[i].Key < targets[j].Name+targets[j].Key
		})
		out = append(out, secretContent{
			Name:                      secret.Name,
			DisplayName:               secret.DisplayName,
			Description:               secret.Description,
			Required:                  secret.Required,
			AutoGenerate:              secret.AutoGenerate,
			Format:                    secret.Format,
			Default:                   secret.Default,
			KubernetesSync:            secret.KubernetesSync,
			KubernetesSecretNamespace: secret.KubernetesSecretNamespace,
			KubernetesSecretName:      secret.KubernetesSecretName,
			KubernetesSyncTargets:     targets,
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func runnerConfigContent(c app.AppRunnerConfig) any {
	return struct {
		EnvVars            any    `json:"env_vars"`
		Type               string `json:"type"`
		InitScriptURL      string `json:"init_script_url"`
		PhoneHomeScriptURL string `json:"phone_home_script_url"`
		InstanceType       string `json:"instance_type"`
		RunnerAPIURL       string `json:"runner_api_url"`
		PublicAPIURL       string `json:"public_api_url"`
	}{
		EnvVars:            c.EnvVars,
		Type:               string(c.Type),
		InitScriptURL:      c.InitScriptURL,
		PhoneHomeScriptURL: c.PhoneHomeScriptURL,
		InstanceType:       c.InstanceType,
		RunnerAPIURL:       c.RunnerAPIURL,
		PublicAPIURL:       c.PublicAPIURL,
	}
}

func stackImpactChanges(oldCfg, newCfg *app.AppConfig) []app.InstallConfigImpact {
	if newCfg == nil {
		return nil
	}
	var old app.AppConfig
	if oldCfg != nil {
		old = *oldCfg
	}

	checks := []struct {
		impact app.InstallConfigImpact
		old    any
		new    any
	}{
		{app.InstallConfigImpactStackConfig, stackConfigContent(old.StackConfig), stackConfigContent(newCfg.StackConfig)},
		{app.InstallConfigImpactPermissions, roleContents(old.PermissionsConfig.Roles), roleContents(newCfg.PermissionsConfig.Roles)},
		{app.InstallConfigImpactPermissions, namedPolicyContents(old.PermissionsConfig.NamedPolicies), namedPolicyContents(newCfg.PermissionsConfig.NamedPolicies)},
		{app.InstallConfigImpactBreakGlass, roleContents(old.BreakGlassConfig.Roles), roleContents(newCfg.BreakGlassConfig.Roles)},
		{app.InstallConfigImpactSecrets, secretContents(old.SecretsConfig.Secrets), secretContents(newCfg.SecretsConfig.Secrets)},
		{app.InstallConfigImpactRunnerConfig, runnerConfigContent(old.RunnerConfig), runnerConfigContent(newCfg.RunnerConfig)},
	}

	impacts := make([]app.InstallConfigImpact, 0, len(checks))
	for _, check := range checks {
		if !contentHashEqual(check.old, check.new) {
			impacts = appendUniqueImpact(impacts, check.impact)
		}
	}
	return impacts
}

func contentHashEqual(a, b any) bool {
	aJSON, err := json.Marshal(a)
	if err != nil {
		return false
	}
	bJSON, err := json.Marshal(b)
	if err != nil {
		return false
	}
	return sha256.Sum256(aJSON) == sha256.Sum256(bJSON)
}

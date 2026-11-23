package activities

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers/registryprune"
)

type GetRegistryPruneTagsRequest struct {
	InstallID      string `validate:"required"`
	SharedRepo     bool
	GracePeriod    time.Duration
	IncludeWhenOff bool
}

type GetRegistryPruneTagsResponse struct {
	Mode             string
	OwnedTags        []string
	OwnedTagPatterns []string
	KeepTags         []string
}

var inFlightDeployStatuses = map[app.InstallDeployStatus]bool{
	"":                                     true,
	app.InstallDeployStatusUnknown:         true,
	app.InstallDeployStatusPending:         true,
	app.InstallDeployStatusQueued:          true,
	app.InstallDeployStatusPlanning:        true,
	app.InstallDeployStatusSyncing:         true,
	app.InstallDeployStatusExecuting:       true,
	app.InstallDeployStatusPendingApproval: true,
	app.InstallDeployStatusRetried:         true,
}

// @temporal-gen-v2 activity
func (a *Activities) GetRegistryPruneTags(ctx context.Context, req *GetRegistryPruneTagsRequest) (*GetRegistryPruneTagsResponse, error) {
	var install app.Install
	if err := a.db.WithContext(ctx).
		Preload("InstallConfig").
		Where(app.Install{ID: req.InstallID}).
		First(&install).Error; err != nil {
		return nil, fmt.Errorf("unable to get install: %w", err)
	}

	appCfg, err := a.appsHelpers.GetFullAppConfig(ctx, install.AppConfigID, false)
	if err != nil {
		return nil, fmt.Errorf("unable to get app config: %w", err)
	}
	var retentionCfg *app.AppRegistryRetentionConfig
	if appCfg.RegistryRetentionConfig.ID != "" {
		retentionCfg = &appCfg.RegistryRetentionConfig
	}

	mode, defaultCount := app.ResolveRegistryRetention(install.InstallConfig, retentionCfg, nil)
	resp := &GetRegistryPruneTagsResponse{Mode: mode}
	if mode == config.RegistryRetentionModeOff && !req.IncludeWhenOff {
		return resp, nil
	}

	retainCounts := map[string]int{}
	for i := range appCfg.ComponentConfigConnections {
		ccc := &appCfg.ComponentConfigConnections[i]
		_, k := app.ResolveRegistryRetention(install.InstallConfig, retentionCfg, ccc)
		retainCounts[ccc.ComponentID] = k
	}

	var installComponents []app.InstallComponent
	if err := a.db.WithContext(ctx).
		Preload("Component").
		Where(app.InstallComponent{InstallID: install.ID}).
		Find(&installComponents).Error; err != nil {
		return nil, fmt.Errorf("unable to get install components: %w", err)
	}
	if len(installComponents) == 0 {
		return resp, nil
	}

	componentByIC := map[string]app.InstallComponent{}
	icIDs := make([]string, 0, len(installComponents))
	for _, ic := range installComponents {
		componentByIC[ic.ID] = ic
		icIDs = append(icIDs, ic.ID)
	}

	var deploys []app.InstallDeploy
	if err := a.db.WithContext(ctx).
		Select("id", "install_component_id", "component_build_id", "status", "created_at").
		Where("install_component_id IN ?", icIDs).
		Find(&deploys).Error; err != nil {
		return nil, fmt.Errorf("unable to get install deploys: %w", err)
	}

	buildIDs := map[string]bool{}
	for _, d := range deploys {
		buildIDs[d.ComponentBuildID] = true
	}
	ids := make([]string, 0, len(buildIDs))
	for id := range buildIDs {
		ids = append(ids, id)
	}
	var builds []app.ComponentBuild
	if len(ids) > 0 {
		if err := a.db.WithContext(ctx).
			Select("id", "source_digest", "resolved_tag").
			Where("id IN ?", ids).
			Find(&builds).Error; err != nil {
			return nil, fmt.Errorf("unable to get component builds: %w", err)
		}
	}
	buildByID := map[string]app.ComponentBuild{}
	for _, b := range builds {
		buildByID[b.ID] = b
	}

	latestActive := map[string]app.InstallDeploy{}
	for _, d := range deploys {
		if d.Status != app.InstallDeployStatusActive {
			continue
		}
		if cur, ok := latestActive[d.InstallComponentID]; !ok || d.CreatedAt.After(cur.CreatedAt) {
			latestActive[d.InstallComponentID] = d
		}
	}
	liveDeployIDs := map[string]bool{}
	for _, d := range latestActive {
		liveDeployIDs[d.ID] = true
	}

	pruneDeploys := make([]registryprune.Deploy, 0, len(deploys))
	for _, d := range deploys {
		ic := componentByIC[d.InstallComponentID]
		tag := d.ComponentBuildID
		if b, ok := buildByID[d.ComponentBuildID]; ok {
			tag = registryprune.InstallRegistryTag(ic.Component.Name, d.ComponentBuildID, b.SourceDigest, b.ResolvedTag, req.SharedRepo)
		}
		pruneDeploys = append(pruneDeploys, registryprune.Deploy{
			ID:          d.ID,
			ComponentID: ic.ComponentID,
			Tag:         tag,
			Status:      string(d.Status),
			CreatedAt:   d.CreatedAt,
			Live:        liveDeployIDs[d.ID],
			InFlight:    inFlightDeployStatuses[d.Status],
		})
	}

	livePlans, err := a.getLiveDeployPlans(ctx, liveDeployIDs)
	if err != nil {
		return nil, err
	}

	sets := registryprune.Compute(registryprune.ComputeInput{
		Deploys:      pruneDeploys,
		RetainCounts: retainCounts,
		DefaultCount: defaultCount,
		LivePlans:    livePlans,
		GracePeriod:  req.GracePeriod,
		Now:          time.Now(),
	})
	resp.OwnedTags = sets.Owned
	resp.OwnedTagPatterns = []string{registryprune.LegacyDeployTagPattern}
	resp.KeepTags = sets.Keep
	return resp, nil
}

func (a *Activities) getLiveDeployPlans(ctx context.Context, deployIDs map[string]bool) ([]string, error) {
	if len(deployIDs) == 0 {
		return nil, nil
	}
	ids := make([]string, 0, len(deployIDs))
	for id := range deployIDs {
		ids = append(ids, id)
	}

	var jobs []app.RunnerJob
	if err := a.db.WithContext(ctx).
		Select("id", "owner_id").
		Where("owner_type = ? AND owner_id IN ?", "install_deploys", ids).
		Find(&jobs).Error; err != nil {
		return nil, fmt.Errorf("unable to get deploy runner jobs: %w", err)
	}
	jobIDs := make([]string, 0, len(jobs))
	deployByJob := make(map[string]string, len(jobs))
	for _, j := range jobs {
		jobIDs = append(jobIDs, j.ID)
		deployByJob[j.ID] = j.OwnerID
	}

	var plans []app.RunnerJobPlan
	if len(jobIDs) > 0 {
		if err := a.db.WithContext(ctx).
			Where("runner_job_id IN ?", jobIDs).
			Find(&plans).Error; err != nil {
			return nil, fmt.Errorf("unable to get deploy runner job plans: %w", err)
		}
	}

	withPlan := map[string]bool{}
	for _, p := range plans {
		withPlan[deployByJob[p.RunnerJobID]] = true
	}
	for id := range deployIDs {
		if !withPlan[id] {
			return nil, fmt.Errorf("live deploy %s has no saved plan, refusing to compute keep set", id)
		}
	}

	out := make([]string, 0, len(plans)*2)
	for _, p := range plans {
		if p.PlanJSON != "" {
			out = append(out, p.PlanJSON)
		}
		if !p.CompositePlan.IsEmpty() {
			byts, err := json.Marshal(p.CompositePlan)
			if err != nil {
				return nil, fmt.Errorf("unable to marshal composite plan: %w", err)
			}
			out = append(out, string(byts))
		}
	}
	return out, nil
}

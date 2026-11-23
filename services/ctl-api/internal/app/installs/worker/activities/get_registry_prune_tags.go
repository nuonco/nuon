package activities

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers/registryprune"
)

type GetRegistryPruneTagsRequest struct {
	InstallID   string `validate:"required"`
	SharedRepo  bool
	GracePeriod time.Duration
}

type GetRegistryPruneTagsResponse struct {
	Mode      string
	OwnedTags []string
	KeepTags  []string
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

	var retention app.AppRegistryRetentionConfig
	res := a.db.WithContext(ctx).
		Where(app.AppRegistryRetentionConfig{AppConfigID: install.AppConfigID}).
		Limit(1).
		Find(&retention)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get registry retention config: %w", res.Error)
	}
	retentionCfg := &retention
	if res.RowsAffected == 0 {
		retentionCfg = nil
	}

	mode, defaultCount := app.ResolveRegistryRetention(install.InstallConfig, retentionCfg, nil)
	resp := &GetRegistryPruneTagsResponse{Mode: mode}

	var cccs []app.ComponentConfigConnection
	if err := a.db.WithContext(ctx).
		Select("id", "component_id", "registry_retain_count").
		Where(app.ComponentConfigConnection{AppConfigID: install.AppConfigID}).
		Find(&cccs).Error; err != nil {
		return nil, fmt.Errorf("unable to get component configs: %w", err)
	}
	retainCounts := map[string]int{}
	for i := range cccs {
		_, k := app.ResolveRegistryRetention(install.InstallConfig, retentionCfg, &cccs[i])
		retainCounts[cccs[i].ComponentID] = k
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
		Select("id").
		Where("owner_type = ? AND owner_id IN ?", "install_deploys", ids).
		Find(&jobs).Error; err != nil {
		return nil, fmt.Errorf("unable to get deploy runner jobs: %w", err)
	}
	if len(jobs) == 0 {
		return nil, nil
	}
	jobIDs := make([]string, 0, len(jobs))
	for _, j := range jobs {
		jobIDs = append(jobIDs, j.ID)
	}

	var plans []app.RunnerJobPlan
	if err := a.db.WithContext(ctx).
		Where("runner_job_id IN ?", jobIDs).
		Find(&plans).Error; err != nil {
		return nil, fmt.Errorf("unable to get deploy runner job plans: %w", err)
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

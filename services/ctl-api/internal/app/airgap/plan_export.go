package airgap

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"

	runnerairgap "github.com/nuonco/nuon/pkg/runner/airgap"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	appshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/views"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

const planEnvelopeSource = "ctl-api airgap-bundle publish"

// ExportPlanEnvelope renders an offline plan from the newest usable install and forces late-bound customer cloud authentication.
func ExportPlanEnvelope(ctx context.Context, db *gorm.DB, blobRead bool, orgID, appID, appConfigID string, runbooks []runnerairgap.RunbookTemplate, report *QualificationReport) (*runnerairgap.Envelope, error) {
	var installs []app.Install
	err := db.WithContext(ctx).
		Where(app.Install{OrgID: orgID, AppID: appID, AppConfigID: appConfigID}).
		Order("created_at DESC").
		Find(&installs).Error
	if err != nil {
		return nil, fmt.Errorf("list installs for app config %s: %w", appConfigID, err)
	}
	if len(installs) == 0 {
		return nil, fmt.Errorf("airgap bundle requires an install of app config %s with executed plans; none found", appConfigID)
	}

	var steps []runnerairgap.Step
	var reference app.Install
	for _, install := range installs {
		steps, err = exportInstallSteps(ctx, db, blobRead, install.ID)
		if err != nil {
			return nil, err
		}
		if len(steps) > 0 {
			reference = install
			break
		}
	}
	if len(steps) == 0 {
		return nil, fmt.Errorf("no install of app config %s has runner jobs with plans in sandbox, deploy, or sync groups", appConfigID)
	}

	appConfig, err := exportAppConfigJSON(ctx, db, appConfigID)
	if err != nil {
		return nil, err
	}
	inputs, err := exportInputSpecs(ctx, db, appConfigID)
	if err != nil {
		return nil, err
	}
	components, err := exportComponents(ctx, db, reference.ID, appConfigID)
	if err != nil {
		return nil, err
	}
	actions, err := exportActionTemplates(ctx, db, blobRead, reference.ID, report)
	if err != nil {
		return nil, err
	}
	drift, err := exportDriftTemplates(ctx, db, steps, components)
	if err != nil {
		return nil, err
	}

	envelope := &runnerairgap.Envelope{
		Version:               "v0",
		OrgID:                 orgID,
		AppID:                 appID,
		InstallID:             reference.ID,
		CreatedAt:             time.Now().UTC(),
		Source:                planEnvelopeSource,
		AppConfig:             appConfig,
		Inputs:                inputs,
		ForceDefaultCloudAuth: true,
		Components:            components,
		Steps:                 steps,
		Actions:               actions,
		Drift:                 drift,
		Runbooks:              runbooks,
	}
	if err := rewriteInputPlaceholders(ctx, db, envelope); err != nil {
		return nil, fmt.Errorf("rewrite install inputs in rendered plans: %w", err)
	}
	if err := envelope.Validate(); err != nil {
		return nil, fmt.Errorf("rendered plan envelope is invalid: %w", err)
	}
	return envelope, nil
}

type exportJob struct {
	job  app.RunnerJob
	plan json.RawMessage
}

// Retries and reruns collapse to the newest logical owner before jobs are linked into one plan chain.
func exportInstallSteps(ctx context.Context, db *gorm.DB, blobRead bool, installID string) ([]runnerairgap.Step, error) {
	var jobs []app.RunnerJob
	jobsTable := views.TableOrViewName(db, &app.RunnerJob{}, "")
	err := db.WithContext(ctx).
		Joins(fmt.Sprintf("JOIN runners ON runners.id = %s.runner_id AND runners.deleted_at = 0", jobsTable)).
		Joins("JOIN runner_groups ON runner_groups.id = runners.runner_group_id AND runner_groups.deleted_at = 0").
		Where("runner_groups.owner_id = ? AND runner_groups.owner_type = ?", installID, "installs").
		Where(fmt.Sprintf(`%s."group" IN ?`, jobsTable), []string{
			string(app.RunnerJobGroupSandbox),
			string(app.RunnerJobGroupDeploy),
			string(app.RunnerJobGroupSync),
		}).
		Order(fmt.Sprintf("%s.created_at, %s.id", jobsTable, jobsTable)).
		Find(&jobs).Error
	if err != nil {
		return nil, fmt.Errorf("list runner jobs for install %s: %w", installID, err)
	}
	if len(jobs) == 0 {
		return nil, nil
	}

	withPlans := make([]exportJob, 0, len(jobs))
	plansByOwner := make(map[string]json.RawMessage)
	for _, job := range jobs {
		plan, err := loadJobPlan(ctx, db, blobRead, job)
		if err != nil {
			return nil, err
		}
		planKey := strings.Join([]string{job.OwnerType, job.OwnerID, string(job.Group), string(job.Type)}, ":")
		if plan == nil {
			if job.Operation != app.RunnerJobOperationTypeApplyPlan {
				continue
			}
			plan = plansByOwner[planKey]
			if plan == nil {
				continue
			}
		} else {
			plansByOwner[planKey] = plan
		}
		withPlans = append(withPlans, exportJob{job: job, plan: plan})
	}
	if len(withPlans) == 0 {
		return nil, nil
	}

	selected, err := dedupeByOwner(ctx, db, withPlans)
	if err != nil {
		return nil, err
	}

	steps := make([]runnerairgap.Step, 0, len(selected))
	for i, item := range selected {
		step := runnerairgap.Step{
			ID:            item.job.ID,
			Name:          strings.TrimSpace(string(item.job.Type) + " " + string(item.job.Operation)),
			JobType:       string(item.job.Type),
			JobOperation:  string(item.job.Operation),
			JobGroup:      string(item.job.Group),
			CompositePlan: item.plan,
		}
		if i > 0 {
			previous := selected[i-1]
			step.DependsOn = []string{previous.job.ID}
			if string(item.job.Operation) == "apply-plan" &&
				item.job.Type == previous.job.Type &&
				strings.HasPrefix(string(previous.job.Operation), "create-") {
				step.PlanFromStep = previous.job.ID
			}
		}
		steps = append(steps, step)
	}
	return steps, nil
}

func loadJobPlan(ctx context.Context, db *gorm.DB, blobRead bool, job app.RunnerJob) (json.RawMessage, error) {
	var plan app.RunnerJobPlan
	err := db.WithContext(ctx).Where(app.RunnerJobPlan{RunnerJobID: job.ID}).First(&plan).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("load plan for runner job %s: %w", job.ID, err)
	}
	cp, _ := plan.GetCompositePlan(ctx, blobRead)
	if cp.IsEmpty() {
		derived, err := plan.DeriveCompositePlan(&job)
		if err != nil || derived.IsEmpty() {
			return nil, nil
		}
		cp = derived
	}
	contents, err := json.Marshal(cp)
	if err != nil {
		return nil, fmt.Errorf("marshal composite plan for runner job %s: %w", job.ID, err)
	}
	return contents, nil
}

// Sandbox runs share one identity, deploys are keyed by component, and other jobs remain distinct by owner.
func dedupeByOwner(ctx context.Context, db *gorm.DB, jobs []exportJob) ([]exportJob, error) {
	type ownerGroup struct {
		identity string
		earliest time.Time
		latest   time.Time
		jobs     []exportJob
	}
	groups := make(map[string]*ownerGroup)
	order := make([]string, 0)
	for _, item := range jobs {
		ownerKey := item.job.OwnerType + ":" + item.job.OwnerID
		group, ok := groups[ownerKey]
		if !ok {
			identity, err := ownerIdentity(ctx, db, item.job)
			if err != nil {
				return nil, err
			}
			group = &ownerGroup{identity: identity, earliest: item.job.CreatedAt}
			groups[ownerKey] = group
			order = append(order, ownerKey)
		}
		if item.job.CreatedAt.After(group.latest) {
			group.latest = item.job.CreatedAt
		}
		group.jobs = append(group.jobs, item)
	}

	newestByIdentity := make(map[string]*ownerGroup)
	for _, key := range order {
		group := groups[key]
		current, ok := newestByIdentity[group.identity]
		if !ok || group.latest.After(current.latest) {
			newestByIdentity[group.identity] = group
		}
	}

	kept := make([]*ownerGroup, 0, len(newestByIdentity))
	for _, group := range newestByIdentity {
		kept = append(kept, group)
	}
	sort.Slice(kept, func(i, j int) bool {
		iSandbox := kept[i].jobs[0].job.Group == app.RunnerJobGroupSandbox
		jSandbox := kept[j].jobs[0].job.Group == app.RunnerJobGroupSandbox
		if iSandbox != jSandbox {
			return iSandbox
		}
		return kept[i].earliest.Before(kept[j].earliest)
	})

	selected := make([]exportJob, 0, len(jobs))
	for _, group := range kept {
		selected = append(selected, group.jobs...)
	}
	return selected, nil
}

func ownerIdentity(ctx context.Context, db *gorm.DB, job app.RunnerJob) (string, error) {
	switch job.OwnerType {
	case "install_sandbox_runs":
		return "sandbox", nil
	case "install_deploys":
		var deploy app.InstallDeploy
		err := db.WithContext(ctx).Where(app.InstallDeploy{ID: job.OwnerID}).First(&deploy).Error
		if err != nil {
			return "", fmt.Errorf("load install deploy %s for runner job %s: %w", job.OwnerID, job.ID, err)
		}
		return "deploy:" + deploy.InstallComponentID, nil
	default:
		return job.OwnerType + ":" + job.OwnerID, nil
	}
}

// The offline runner requires raw app-config rows with the latest sandbox config embedded as "sandbox".
func exportAppConfigJSON(ctx context.Context, db *gorm.DB, appConfigID string) (json.RawMessage, error) {
	var raw string
	err := db.WithContext(ctx).Raw(`
SELECT (row_to_json(app_config)::jsonb || jsonb_build_object(
	'sandbox',
	(
		SELECT row_to_json(sandbox)::jsonb
		FROM app_sandbox_configs AS sandbox
		WHERE sandbox.app_config_id = app_config.id AND sandbox.deleted_at = 0
		ORDER BY sandbox.created_at DESC
		LIMIT 1
	)
))::text
FROM app_configs AS app_config
WHERE id = ? AND deleted_at = 0`, appConfigID).Scan(&raw).Error
	if err != nil {
		return nil, fmt.Errorf("serialize app config %s: %w", appConfigID, err)
	}
	if raw == "" {
		return nil, fmt.Errorf("app config %s not found", appConfigID)
	}
	var probe struct {
		Sandbox json.RawMessage `json:"sandbox"`
	}
	if err := json.Unmarshal([]byte(raw), &probe); err != nil {
		return nil, fmt.Errorf("parse app config %s: %w", appConfigID, err)
	}
	if len(probe.Sandbox) == 0 || string(probe.Sandbox) == "null" {
		return nil, fmt.Errorf("app config %s has no sandbox config; airgap replay of sandbox jobs requires one", appConfigID)
	}

	connections, err := exportComponentConfigConnections(ctx, db, appConfigID)
	if err != nil {
		return nil, err
	}
	var merged map[string]json.RawMessage
	if err := json.Unmarshal([]byte(raw), &merged); err != nil {
		return nil, fmt.Errorf("parse app config %s: %w", appConfigID, err)
	}
	connectionsJSON, err := json.Marshal(connections)
	if err != nil {
		return nil, fmt.Errorf("serialize component config connections for %s: %w", appConfigID, err)
	}
	merged["component_config_connections"] = connectionsJSON
	out, err := json.Marshal(merged)
	if err != nil {
		return nil, fmt.Errorf("serialize app config %s: %w", appConfigID, err)
	}
	return out, nil
}

// Unchanged components must fall back to their latest earlier config, matching GetFullAppConfig semantics.
func exportComponentConfigConnections(ctx context.Context, db *gorm.DB, appConfigID string) ([]app.ComponentConfigConnection, error) {
	var appCfg app.AppConfig
	err := db.WithContext(ctx).
		Where(app.AppConfig{ID: appConfigID}).
		Scopes(appshelpers.PreloadAppConfigComponentConfigConnections).
		First(&appCfg).Error
	if err != nil {
		return nil, fmt.Errorf("load component configs for app config %s: %w", appConfigID, err)
	}

	seen := map[string]bool{}
	for _, connection := range appCfg.ComponentConfigConnections {
		seen[connection.ComponentID] = true
	}
	var missing []string
	for _, componentID := range appCfg.ComponentIDs {
		if !seen[componentID] {
			missing = append(missing, componentID)
		}
	}
	connections := appCfg.ComponentConfigConnections
	if len(missing) > 0 {
		var latest []app.ComponentConfigConnection
		err = db.WithContext(ctx).
			Scopes(
				scopes.WithDisableViews,
				scopes.WithOverrideTable("component_config_connections_latest_configs_view"),
			).
			Preload("Component").
			Preload("TerraformModuleComponentConfig").
			Preload("TerraformModuleComponentConfig.PublicGitVCSConfig").
			Preload("TerraformModuleComponentConfig.ConnectedGithubVCSConfig").
			Preload("HelmComponentConfig").
			Preload("HelmComponentConfig.PublicGitVCSConfig").
			Preload("HelmComponentConfig.ConnectedGithubVCSConfig").
			Preload("DockerBuildComponentConfig").
			Preload("DockerBuildComponentConfig.PublicGitVCSConfig").
			Preload("DockerBuildComponentConfig.ConnectedGithubVCSConfig").
			Preload("ExternalImageComponentConfig").
			Preload("JobComponentConfig").
			Preload("KubernetesManifestComponentConfig").
			Preload("PulumiComponentConfig").
			Preload("PulumiComponentConfig.PublicGitVCSConfig").
			Preload("PulumiComponentConfig.ConnectedGithubVCSConfig").
			Where("component_id IN ?", missing).
			Find(&latest).Error
		if err != nil {
			return nil, fmt.Errorf("load latest component configs for app config %s: %w", appConfigID, err)
		}
		connections = append(connections, latest...)
	}
	if len(connections) != len(appCfg.ComponentIDs) {
		return nil, fmt.Errorf("app config %s: found %d component configs, expected %d", appConfigID, len(connections), len(appCfg.ComponentIDs))
	}
	return connections, nil
}

// exportComponents freezes the component-health ownership metadata the
// connected runner would fetch live (GetRunnerInstallComponents) into the
// envelope. Probes are deliberately excluded: their targets are rendered from
// install state, which belongs to the reference install, not the customer's.
func exportComponents(ctx context.Context, db *gorm.DB, installID, appConfigID string) ([]runnerairgap.ComponentSpec, error) {
	var installComponents []app.InstallComponent
	if err := db.WithContext(ctx).
		Preload("Component").
		Where(app.InstallComponent{InstallID: installID}).
		Find(&installComponents).Error; err != nil {
		return nil, fmt.Errorf("list install components for install %s: %w", installID, err)
	}
	if len(installComponents) == 0 {
		return nil, nil
	}
	connections, err := exportComponentConfigConnections(ctx, db, appConfigID)
	if err != nil {
		return nil, err
	}
	return componentSpecs(installComponents, connections), nil
}

func componentSpecs(installComponents []app.InstallComponent, connections []app.ComponentConfigConnection) []runnerairgap.ComponentSpec {
	helmByComponent := map[string]*app.HelmComponentConfig{}
	for i := range connections {
		if cfg := connections[i].HelmComponentConfig; cfg != nil {
			helmByComponent[connections[i].ComponentID] = cfg
		}
	}
	specs := make([]runnerairgap.ComponentSpec, 0, len(installComponents))
	for _, ic := range installComponents {
		spec := runnerairgap.ComponentSpec{
			InstallComponentID: ic.ID,
			ComponentID:        ic.ComponentID,
			ComponentName:      ic.Component.Name,
			ComponentType:      string(ic.Component.Type),
		}
		if cfg, ok := helmByComponent[ic.ComponentID]; ok {
			spec.HelmReleaseName = cfg.ChartName
			spec.HelmNamespace = cfg.Namespace.ValueString()
		}
		specs = append(specs, spec)
	}
	return specs
}

func exportInputSpecs(ctx context.Context, db *gorm.DB, appConfigID string) ([]runnerairgap.InputSpec, error) {
	var config app.AppInputConfig
	err := db.WithContext(ctx).
		Preload("AppInputs", func(tx *gorm.DB) *gorm.DB {
			return tx.Order(`app_inputs."index", app_inputs.created_at, app_inputs.id`)
		}).
		Where(app.AppInputConfig{AppConfigID: appConfigID}).
		First(&config).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("load inputs for app config %s: %w", appConfigID, err)
	}
	specs := make([]runnerairgap.InputSpec, 0, len(config.AppInputs))
	for _, in := range config.AppInputs {
		specs = append(specs, runnerairgap.InputSpec{
			Name:        in.Name,
			Type:        string(in.Type),
			Description: in.Description,
			Secret:      in.Sensitive,
			Required:    in.Required,
			Default:     in.Default,
		})
	}
	return specs, nil
}

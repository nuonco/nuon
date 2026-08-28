// Package customer_managed renders artifacts for offline customer-managed installations.
package customermanaged

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"gorm.io/gorm"

	customermanaged "github.com/nuonco/nuon/pkg/customer_managed"
	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/views"
)

type actionExport struct {
	job      app.RunnerJob
	plan     json.RawMessage
	workflow app.ActionWorkflow
	config   app.ActionWorkflowConfig
}

func exportActionTemplates(ctx context.Context, db *gorm.DB, blobRead bool, installID string, report *QualificationReport) ([]customermanaged.ActionTemplate, error) {
	var jobs []app.RunnerJob
	jobsTable := views.TableOrViewName(db, &app.RunnerJob{}, "")
	err := db.WithContext(ctx).
		Joins(fmt.Sprintf("JOIN runners ON runners.id = %s.runner_id AND runners.deleted_at = 0", jobsTable)).
		Joins("JOIN runner_groups ON runner_groups.id = runners.runner_group_id AND runner_groups.deleted_at = 0").
		Where("runner_groups.owner_id = ? AND runner_groups.owner_type = ?", installID, "installs").
		Where(app.RunnerJob{Group: app.RunnerJobGroupActions, Type: app.RunnerJobTypeActionsWorkflowRun, Operation: app.RunnerJobOperationTypeExec}).
		Order(fmt.Sprintf("%s.created_at DESC, %s.id DESC", jobsTable, jobsTable)).
		Find(&jobs).Error
	if err != nil {
		return nil, fmt.Errorf("list action runner jobs for install %s: %w", installID, err)
	}

	selected := map[string]actionExport{}
	for _, job := range jobs {
		plan, err := loadJobPlan(ctx, db, blobRead, job)
		if err != nil {
			return nil, err
		}
		if plan == nil {
			continue
		}
		var composite plantypes.CompositePlan
		if err := json.Unmarshal(plan, &composite); err != nil {
			return nil, fmt.Errorf("decode action plan for runner job %s: %w", job.ID, err)
		}
		if composite.ActionWorkflowRunPlan == nil {
			continue
		}
		gitStep := ""
		for _, step := range composite.ActionWorkflowRunPlan.Steps {
			if step != nil && step.GitSource != nil && (step.GitSource.URL != "" || step.GitSource.Ref != "" || step.GitSource.Path != "" || step.GitSource.RecurseSubmodules) {
				gitStep = step.ID
				break
			}
		}
		if gitStep != "" {
			if report != nil {
				report.Warnings = append(report.Warnings, Finding{Code: "action.git_source_excluded", Member: "runner_job:" + job.ID, Message: fmt.Sprintf("action was excluded from the bundle: step %s sources its contents from Git, and portable bundle actions must be inline-only", gitStep)})
			}
			continue
		}
		resolved, ok, err := resolveActionOwner(ctx, db, job)
		if err != nil {
			return nil, err
		}
		if !ok {
			if report != nil {
				report.Warnings = append(report.Warnings, Finding{Code: "action.plan_owner_unresolved", Member: "runner_job:" + job.ID, Message: "action plan was excluded because its workflow config could not be resolved"})
			}
			continue
		}
		if _, exists := selected[resolved.workflow.ID]; !exists {
			resolved.job = job
			resolved.plan = plan
			selected[resolved.workflow.ID] = resolved
		}
	}
	if report != nil {
		finish(report)
	}

	ids := make([]string, 0, len(selected))
	for id := range selected {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	templates := make([]customermanaged.ActionTemplate, 0, len(ids))
	for _, id := range ids {
		item := selected[id]
		cronSchedule := ""
		for _, trigger := range item.config.Triggers {
			if trigger.Type == app.ActionWorkflowTriggerTypeCron {
				cronSchedule = trigger.CronSchedule
				break
			}
		}
		templates = append(templates, customermanaged.ActionTemplate{ID: id, Name: item.workflow.Name, CronSchedule: cronSchedule, JobType: string(item.job.Type), JobGroup: string(item.job.Group), JobOperation: string(item.job.Operation), CompositePlan: item.plan})
	}
	return templates, nil
}

func resolveActionOwner(ctx context.Context, db *gorm.DB, job app.RunnerJob) (actionExport, bool, error) {
	if job.OwnerType != "install_action_workflow_runs" {
		return actionExport{}, false, nil
	}
	var run app.InstallActionWorkflowRun
	if err := db.WithContext(ctx).Where(app.InstallActionWorkflowRun{ID: job.OwnerID}).First(&run).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return actionExport{}, false, nil
		}
		return actionExport{}, false, fmt.Errorf("load action run %s for runner job %s: %w", job.OwnerID, job.ID, err)
	}
	if !run.ActionWorkflowConfigID.Valid || !run.InstallActionWorkflowID.Valid {
		return actionExport{}, false, nil
	}
	var config app.ActionWorkflowConfig
	if err := db.WithContext(ctx).Preload("Triggers").Where(app.ActionWorkflowConfig{ID: run.ActionWorkflowConfigID.String}).First(&config).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return actionExport{}, false, nil
		}
		return actionExport{}, false, fmt.Errorf("load action config %s for runner job %s: %w", run.ActionWorkflowConfigID.String, job.ID, err)
	}
	var installWorkflow app.InstallActionWorkflow
	if err := db.WithContext(ctx).Preload("ActionWorkflow").Where(app.InstallActionWorkflow{ID: run.InstallActionWorkflowID.String}).First(&installWorkflow).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return actionExport{}, false, nil
		}
		return actionExport{}, false, fmt.Errorf("load install action workflow %s for runner job %s: %w", run.InstallActionWorkflowID.String, job.ID, err)
	}
	return actionExport{workflow: installWorkflow.ActionWorkflow, config: config}, installWorkflow.ActionWorkflow.ID != "", nil
}

func exportDriftTemplates(ctx context.Context, db *gorm.DB, steps []customermanaged.Step, components []customermanaged.ComponentSpec) ([]customermanaged.DriftTemplate, error) {
	componentByInstallID := make(map[string]customermanaged.ComponentSpec, len(components))
	for _, component := range components {
		componentByInstallID[component.InstallComponentID] = component
	}
	byComponent := map[string]customermanaged.DriftTemplate{}
	for _, step := range steps {
		if step.JobType != string(app.RunnerJobTypeTerraformDeploy) || step.JobGroup != string(app.RunnerJobGroupDeploy) || (step.JobOperation != string(app.RunnerJobOperationTypeCreateApplyPlan) && step.JobOperation != string(app.RunnerJobOperationTypeApplyPlan)) {
			continue
		}
		var job app.RunnerJob
		if err := db.WithContext(ctx).Where(app.RunnerJob{ID: step.ID}).First(&job).Error; err != nil {
			return nil, fmt.Errorf("load terraform runner job %s: %w", step.ID, err)
		}
		if job.OwnerType != "install_deploys" {
			continue
		}
		var deploy app.InstallDeploy
		if err := db.WithContext(ctx).Where(app.InstallDeploy{ID: job.OwnerID}).First(&deploy).Error; err != nil {
			return nil, fmt.Errorf("load install deploy %s for drift template: %w", job.OwnerID, err)
		}
		component, ok := componentByInstallID[deploy.InstallComponentID]
		if !ok {
			return nil, fmt.Errorf("resolve component for terraform deploy %s", deploy.ID)
		}
		plan, err := clearTerraformPlanArtifacts(step.CompositePlan)
		if err != nil {
			return nil, fmt.Errorf("derive drift template for step %s: %w", step.ID, err)
		}
		byComponent[component.ComponentID] = customermanaged.DriftTemplate{ID: "drift-" + component.ComponentID, ComponentID: component.ComponentID, ComponentName: component.ComponentName, JobType: step.JobType, JobGroup: step.JobGroup, JobOperation: string(app.RunnerJobOperationTypeCreateApplyPlan), CompositePlan: plan}
	}
	ids := make([]string, 0, len(byComponent))
	for id := range byComponent {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	result := make([]customermanaged.DriftTemplate, 0, len(ids))
	for _, id := range ids {
		result = append(result, byComponent[id])
	}
	return result, nil
}

func clearTerraformPlanArtifacts(plan json.RawMessage) (json.RawMessage, error) {
	var composite map[string]any
	if err := json.Unmarshal(plan, &composite); err != nil {
		return nil, err
	}
	deploy, ok := composite["deploy_plan"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("composite plan has no deploy_plan")
	}
	deploy["apply_plan_contents"] = ""
	deploy["apply_plan_display"] = ""
	terraformPlan, ok := deploy["terraform"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("deploy plan has no terraform plan")
	}
	terraformPlan["plan_json"] = nil
	return json.Marshal(composite)
}

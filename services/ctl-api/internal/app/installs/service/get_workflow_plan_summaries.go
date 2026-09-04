package service

import (
	"bytes"
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

var helmPlanSummaryPattern = regexp.MustCompile(`Plan:\s*(\d+)\s+to add,\s*(\d+)\s+to change,\s*(\d+)\s+to destroy`)

type terraformPlanSummary struct {
	ResourceChanges []struct {
		Change struct {
			Actions []string `json:"actions"`
		} `json:"change"`
	} `json:"resource_changes"`
}

type pulumiPlanSummary struct {
	ChangeSummary   map[string]int `json:"change_summary"`
	ResourceChanges []struct {
		Action string `json:"action"`
	} `json:"resource_changes"`
}

type helmPlanSummary struct {
	Plan        string `json:"plan"`
	ContentDiff []struct {
		Before json.RawMessage `json:"before"`
		After  json.RawMessage `json:"after"`
	} `json:"helm_content_diff"`
}

type kubernetesPlanSummary struct {
	ContentDiff []struct {
		Op    string `json:"op"`
		Type  int    `json:"type"`
		Error string `json:"error"`
	} `json:"k8s_content_diff"`
}

// @ID						GetWorkflowPlanSummaries
// @Summary				get consolidated plan summaries for a workflow
// @Param					workflow_id	path	string	true	"workflow id"
// @Tags					installs
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{array}	app.StepChangeSummary
// @Router					/v1/workflows/{workflow_id}/plan-summaries [GET]
func (s *service) GetWorkflowPlanSummaries(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to get org from context"))
		return
	}

	workflowID := ctx.Param("workflow_id")
	summaries, err := s.getWorkflowPlanSummaries(ctx, org.ID, workflowID)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to get workflow plan summaries"))
		return
	}

	ctx.JSON(http.StatusOK, summaries)
}

func (s *service) getWorkflowPlanSummaries(ctx *gin.Context, orgID, workflowID string) ([]app.StepChangeSummary, error) {
	var workflow app.Workflow
	if err := s.db.WithContext(ctx).
		Select("id").
		Where(app.Workflow{ID: workflowID, OrgID: orgID}).
		First(&workflow).Error; err != nil {
		return nil, errors.Wrap(err, "unable to get workflow")
	}

	var steps []app.WorkflowStep
	if err := s.db.WithContext(ctx).
		Where(app.WorkflowStep{InstallWorkflowID: workflowID, OrgID: orgID}).
		Preload("Approval").
		Order("group_idx, group_retry_idx, idx, created_at asc").
		Find(&steps).Error; err != nil {
		return nil, errors.Wrap(err, "unable to get workflow steps")
	}

	componentNames, err := s.getStepComponentNames(ctx, orgID, steps)
	if err != nil {
		return nil, err
	}

	summaries := make([]app.StepChangeSummary, 0, len(steps))
	for _, step := range steps {
		if step.Approval == nil || !isSummaryApprovalType(step.Approval.Type) {
			continue
		}

		contents, _ := step.Approval.GetContents(ctx, s.cfg.BlobReadEnabled)
		summary, summaryErr := buildStepChangeSummary(step, componentNames[step.StepTargetID], contents)
		if summaryErr != nil {
			summary.Status = app.StepChangeStatusError
		}
		summaries = append(summaries, summary)
	}

	return summaries, nil
}

func (s *service) getStepComponentNames(ctx *gin.Context, orgID string, steps []app.WorkflowStep) (map[string]string, error) {
	deployIDs := make([]string, 0, len(steps))
	for _, step := range steps {
		targetType := app.WorkflowStepTargetType(step.StepTargetType)
		if targetType == app.WorkflowStepTargetTypeInstallDeploy ||
			targetType == app.WorkflowStepTargetTypeInstallDeploys {
			deployIDs = append(deployIDs, step.StepTargetID)
		}
	}
	if len(deployIDs) == 0 {
		return map[string]string{}, nil
	}

	var deploys []app.InstallDeploy
	if err := s.db.WithContext(ctx).
		Where(app.InstallDeploy{OrgID: orgID}).
		Preload("InstallComponent.Component").
		Find(&deploys, deployIDs).Error; err != nil {
		return nil, errors.Wrap(err, "unable to get workflow deploy components")
	}

	names := make(map[string]string, len(deploys))
	for _, deploy := range deploys {
		names[deploy.ID] = deploy.InstallComponent.Component.Name
	}
	return names, nil
}

func buildStepChangeSummary(step app.WorkflowStep, componentName, contents string) (app.StepChangeSummary, error) {
	approval := step.Approval
	summary := app.StepChangeSummary{
		StepID:        step.ID,
		StepName:      step.Name,
		ApprovalID:    approval.ID,
		ComponentName: componentName,
		PlanType:      app.StepChangePlanType(approval.Type),
		Status:        stepChangeStatus(step.Status.Status),
		HasDetail:     approval.Type != app.InstallCreationApprovalType,
	}

	if approval.Type == app.AppBranchPlanApprovalType ||
		approval.Type == app.InstallCreationApprovalType {
		return summary, nil
	}
	if contents == "" {
		if summary.Status == app.StepChangeStatusGenerating {
			return summary, nil
		}
		return summary, errors.New("approval contents are empty")
	}

	counts, err := planChangeCounts(approval.Type, []byte(contents))
	if err != nil {
		return summary, err
	}
	summary.Counts = counts
	return summary, nil
}

func isSummaryApprovalType(approvalType app.WorkflowStepApprovalType) bool {
	switch approvalType {
	case app.TerraformPlanApprovalType,
		app.PulumiApprovalType,
		app.HelmApprovalApprovalType,
		app.KubernetesManifestApprovalType,
		app.AppBranchPlanApprovalType,
		app.InstallCreationApprovalType:
		return true
	default:
		return false
	}
}

func stepChangeStatus(status app.Status) app.StepChangeStatus {
	switch status {
	case app.AwaitingApproval:
		return app.StepChangeStatusPendingApproval
	case app.WorkflowStepApprovalStatusApproved:
		return app.StepChangeStatusApproved
	case app.WorkflowStepApprovalStatusApprovalDenied:
		return app.StepChangeStatusDenied
	case app.StatusSuccess, app.StatusUserSkipped, app.StatusAutoSkipped, app.StatusDiscarded:
		return app.StepChangeStatusApplied
	case app.StatusError, app.StatusCancelled, app.StatusFailedPendingRetry:
		return app.StepChangeStatusError
	default:
		return app.StepChangeStatusGenerating
	}
}

func planChangeCounts(approvalType app.WorkflowStepApprovalType, contents []byte) (app.StepChangeCounts, error) {
	switch approvalType {
	case app.TerraformPlanApprovalType:
		return terraformChangeCounts(contents)
	case app.PulumiApprovalType:
		return pulumiChangeCounts(contents)
	case app.HelmApprovalApprovalType:
		return helmChangeCounts(contents)
	case app.KubernetesManifestApprovalType:
		return kubernetesChangeCounts(contents)
	default:
		return app.StepChangeCounts{}, nil
	}
}

func terraformChangeCounts(contents []byte) (app.StepChangeCounts, error) {
	var plan terraformPlanSummary
	if err := json.Unmarshal(contents, &plan); err != nil {
		return app.StepChangeCounts{}, errors.Wrap(err, "unable to parse terraform plan")
	}

	var counts app.StepChangeCounts
	for _, resource := range plan.ResourceChanges {
		actions := resource.Change.Actions
		if containsAction(actions, "create") && containsAction(actions, "delete") {
			counts.Replace++
			continue
		}
		for _, action := range actions {
			switch action {
			case "create":
				counts.Create++
			case "update":
				counts.Update++
			case "delete":
				counts.Delete++
			case "no-op":
				counts.Noop++
			}
		}
	}
	return counts, nil
}

func pulumiChangeCounts(contents []byte) (app.StepChangeCounts, error) {
	var plan pulumiPlanSummary
	if err := json.Unmarshal(contents, &plan); err != nil {
		return app.StepChangeCounts{}, errors.Wrap(err, "unable to parse pulumi plan")
	}

	if len(plan.ChangeSummary) > 0 {
		return app.StepChangeCounts{
			Create:  plan.ChangeSummary["create"],
			Update:  plan.ChangeSummary["update"],
			Delete:  plan.ChangeSummary["delete"],
			Replace: plan.ChangeSummary["replace"],
			Noop:    plan.ChangeSummary["same"] + plan.ChangeSummary["no-op"],
		}, nil
	}

	var counts app.StepChangeCounts
	for _, resource := range plan.ResourceChanges {
		switch resource.Action {
		case "create":
			counts.Create++
		case "update":
			counts.Update++
		case "delete":
			counts.Delete++
		case "replace":
			counts.Replace++
		case "same", "no-op":
			counts.Noop++
		}
	}
	return counts, nil
}

func helmChangeCounts(contents []byte) (app.StepChangeCounts, error) {
	var plan helmPlanSummary
	if err := json.Unmarshal(contents, &plan); err != nil {
		return app.StepChangeCounts{}, errors.Wrap(err, "unable to parse helm plan")
	}

	if matches := helmPlanSummaryPattern.FindStringSubmatch(plan.Plan); len(matches) == 4 {
		create, _ := strconv.Atoi(matches[1])
		update, _ := strconv.Atoi(matches[2])
		deleteCount, _ := strconv.Atoi(matches[3])
		return app.StepChangeCounts{Create: create, Update: update, Delete: deleteCount}, nil
	}

	var counts app.StepChangeCounts
	for _, resource := range plan.ContentDiff {
		hasBefore := hasJSONValue(resource.Before)
		hasAfter := hasJSONValue(resource.After)
		switch {
		case !hasBefore && hasAfter:
			counts.Create++
		case hasBefore && !hasAfter:
			counts.Delete++
		default:
			counts.Update++
		}
	}
	return counts, nil
}

func kubernetesChangeCounts(contents []byte) (app.StepChangeCounts, error) {
	var plan kubernetesPlanSummary
	if err := json.Unmarshal(contents, &plan); err != nil {
		return app.StepChangeCounts{}, errors.Wrap(err, "unable to parse kubernetes plan")
	}

	var counts app.StepChangeCounts
	for _, resource := range plan.ContentDiff {
		if resource.Error != "" {
			continue
		}
		switch {
		case resource.Op == "delete" || resource.Type == 1:
			counts.Delete++
		case resource.Type == 2:
			counts.Create++
		case resource.Type == 3:
			counts.Update++
		case resource.Type == 0:
			counts.Noop++
		}
	}
	return counts, nil
}

func containsAction(actions []string, target string) bool {
	for _, action := range actions {
		if action == target {
			return true
		}
	}
	return false
}

func hasJSONValue(value json.RawMessage) bool {
	trimmed := bytes.TrimSpace(value)
	return len(trimmed) > 0 && !bytes.Equal(trimmed, []byte("null"))
}

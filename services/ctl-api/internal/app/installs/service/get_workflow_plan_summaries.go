package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

const planSummaryReadConcurrency = 8

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
	Plan        string          `json:"plan"`
	ContentDiff json.RawMessage `json:"helm_content_diff"`
}

type helmContentDiffEntry struct {
	Before json.RawMessage `json:"before"`
	After  json.RawMessage `json:"after"`
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

	enabled, err := s.featuresClient.FeatureEnabled(ctx, app.OrgFeaturePlanSummaries)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to check plan-summaries feature"))
		return
	}
	if !enabled {
		ctx.Error(stderr.ErrAuthorization{
			Err:         fmt.Errorf("plan summaries are not enabled for org %s", org.ID),
			Description: "The plan summaries feature is not enabled for this organization.",
		})
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
		Preload("Approval", func(db *gorm.DB) *gorm.DB {
			return db.Omit("contents")
		}).
		Order("group_idx, group_retry_idx, idx, created_at asc").
		Find(&steps).Error; err != nil {
		return nil, errors.Wrap(err, "unable to get workflow steps")
	}

	componentNames, err := s.getStepComponentNames(ctx, orgID, steps)
	if err != nil {
		return nil, err
	}

	summaries := make([]app.StepChangeSummary, 0, len(steps))
	countable := make([]countableApproval, 0, len(steps))
	for _, step := range steps {
		approval := step.Approval
		if approval == nil || !isSummaryApprovalType(approval.Type) {
			continue
		}

		summary := newStepChangeSummary(step, componentNames[step.StepTargetID])
		if planTypeHasCounts(approval.Type) && summary.Status != app.StepChangeStatusGenerating {
			countable = append(countable, countableApproval{approval: approval, idx: len(summaries)})
		}
		summaries = append(summaries, summary)
	}

	s.applyPlanChangeCounts(ctx, orgID, summaries, countable)

	return summaries, nil
}

func (s *service) getStepComponentNames(ctx *gin.Context, orgID string, steps []app.WorkflowStep) (map[string]string, error) {
	seen := make(map[string]struct{}, len(steps))
	deployIDs := make([]string, 0, len(steps))
	for _, step := range steps {
		targetType := app.WorkflowStepTargetType(step.StepTargetType)
		if targetType != app.WorkflowStepTargetTypeInstallDeploy &&
			targetType != app.WorkflowStepTargetTypeInstallDeploys {
			continue
		}
		if _, ok := seen[step.StepTargetID]; ok {
			continue
		}
		seen[step.StepTargetID] = struct{}{}
		deployIDs = append(deployIDs, step.StepTargetID)
	}
	if len(deployIDs) == 0 {
		return map[string]string{}, nil
	}

	var deploys []app.InstallDeploy
	if err := s.db.WithContext(ctx).
		Where(app.InstallDeploy{OrgID: orgID}).
		Where("id IN ?", deployIDs).
		Preload("InstallComponent.Component").
		Find(&deploys).Error; err != nil {
		return nil, errors.Wrap(err, "unable to get workflow deploy components")
	}

	names := make(map[string]string, len(deploys))
	for _, deploy := range deploys {
		names[deploy.ID] = deploy.InstallComponent.Component.Name
	}
	return names, nil
}

type countableApproval struct {
	approval *app.WorkflowStepApproval
	idx      int
}

func (s *service) applyPlanChangeCounts(
	ctx *gin.Context, orgID string, summaries []app.StepChangeSummary, countable []countableApproval,
) {
	if len(countable) == 0 {
		return
	}

	blobCtx := ctx.Request.Context()
	counted := make([]bool, len(countable))

	g := new(errgroup.Group)
	g.SetLimit(planSummaryReadConcurrency)
	for i, item := range countable {
		blob := item.approval.ContentsBlob
		if !s.cfg.BlobReadEnabled || blob == nil || !blob.IsSet() {
			continue
		}
		g.Go(func() error {
			counts, err := s.blobPlanChangeCounts(blobCtx, item.approval)
			if err != nil {
				s.l.Warn("unable to read approval plan from blob, falling back to contents column",
					zap.String("approval_id", item.approval.ID),
					zap.Error(err))
				return nil
			}
			summaries[item.idx].Counts = counts
			counted[i] = true
			return nil
		})
	}
	_ = g.Wait()

	pending := make([]countableApproval, 0, len(countable))
	for i, item := range countable {
		if !counted[i] {
			pending = append(pending, item)
		}
	}
	if len(pending) == 0 {
		return
	}

	contents, err := s.getApprovalContentsByID(ctx, orgID, pending)
	if err != nil {
		s.l.Warn("unable to read approval plan contents", zap.Error(err))
		for _, item := range pending {
			summaries[item.idx].Status = app.StepChangeStatusError
		}
		return
	}

	for _, item := range pending {
		body := contents[item.approval.ID]
		if body == "" {
			summaries[item.idx].Status = app.StepChangeStatusError
			continue
		}
		counts, err := planChangeCounts(item.approval.Type, strings.NewReader(body))
		if err != nil {
			s.l.Warn("unable to parse approval plan",
				zap.String("approval_id", item.approval.ID),
				zap.Error(err))
			summaries[item.idx].Status = app.StepChangeStatusError
			continue
		}
		summaries[item.idx].Counts = counts
	}
}

func (s *service) blobPlanChangeCounts(
	ctx context.Context, approval *app.WorkflowStepApproval,
) (app.StepChangeCounts, error) {
	if approval.ContentsBlob == nil {
		return app.StepChangeCounts{}, errors.New("approval has no contents blob")
	}

	s3Key := approval.ContentsBlob.Metadata().S3Key
	if s3Key == "" {
		return app.StepChangeCounts{}, errors.New("approval blob has no s3 key")
	}

	reader, err := s.blobSvc.DownloadStream(ctx, s3Key)
	if err != nil {
		return app.StepChangeCounts{}, errors.Wrap(err, "unable to download approval contents")
	}
	defer reader.Close()

	return planChangeCounts(approval.Type, reader)
}

func (s *service) getApprovalContentsByID(
	ctx *gin.Context, orgID string, pending []countableApproval,
) (map[string]string, error) {
	ids := make([]string, 0, len(pending))
	for _, item := range pending {
		ids = append(ids, item.approval.ID)
	}

	var rows []app.WorkflowStepApproval
	if err := s.db.WithContext(ctx).
		Select("id", "contents").
		Where(app.WorkflowStepApproval{OrgID: orgID}).
		Where("id IN ?", ids).
		Find(&rows).Error; err != nil {
		return nil, errors.Wrap(err, "unable to get approval contents")
	}

	contents := make(map[string]string, len(rows))
	for _, row := range rows {
		contents[row.ID] = row.Contents
	}
	return contents, nil
}

func newStepChangeSummary(step app.WorkflowStep, componentName string) app.StepChangeSummary {
	approval := step.Approval
	return app.StepChangeSummary{
		StepID:        step.ID,
		StepName:      step.Name,
		ApprovalID:    approval.ID,
		ComponentName: componentName,
		PlanType:      app.StepChangePlanType(approval.Type),
		Status:        stepChangeStatus(step.Status.Status),
		HasDetail:     approval.Type != app.InstallCreationApprovalType,
	}
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

func planTypeHasCounts(approvalType app.WorkflowStepApprovalType) bool {
	switch approvalType {
	case app.TerraformPlanApprovalType,
		app.PulumiApprovalType,
		app.HelmApprovalApprovalType,
		app.KubernetesManifestApprovalType:
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

func planChangeCounts(approvalType app.WorkflowStepApprovalType, r io.Reader) (app.StepChangeCounts, error) {
	switch approvalType {
	case app.TerraformPlanApprovalType:
		return terraformChangeCounts(r)
	case app.PulumiApprovalType:
		return pulumiChangeCounts(r)
	case app.HelmApprovalApprovalType:
		return helmChangeCounts(r)
	case app.KubernetesManifestApprovalType:
		return kubernetesChangeCounts(r)
	default:
		return app.StepChangeCounts{}, nil
	}
}

func terraformChangeCounts(r io.Reader) (app.StepChangeCounts, error) {
	var plan terraformPlanSummary
	if err := json.NewDecoder(r).Decode(&plan); err != nil {
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

func pulumiChangeCounts(r io.Reader) (app.StepChangeCounts, error) {
	var plan pulumiPlanSummary
	if err := json.NewDecoder(r).Decode(&plan); err != nil {
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

func helmChangeCounts(r io.Reader) (app.StepChangeCounts, error) {
	var plan helmPlanSummary
	if err := json.NewDecoder(r).Decode(&plan); err != nil {
		return app.StepChangeCounts{}, errors.Wrap(err, "unable to parse helm plan")
	}

	if matches := helmPlanSummaryPattern.FindStringSubmatch(plan.Plan); len(matches) == 4 {
		create, _ := strconv.Atoi(matches[1])
		update, _ := strconv.Atoi(matches[2])
		deleteCount, _ := strconv.Atoi(matches[3])
		return app.StepChangeCounts{Create: create, Update: update, Delete: deleteCount}, nil
	}

	if len(plan.ContentDiff) == 0 {
		return app.StepChangeCounts{}, nil
	}

	var entries []helmContentDiffEntry
	if err := json.Unmarshal(plan.ContentDiff, &entries); err != nil {
		return app.StepChangeCounts{}, errors.Wrap(err, "unable to parse helm content diff")
	}

	var counts app.StepChangeCounts
	for _, entry := range entries {
		hasBefore := hasJSONValue(entry.Before)
		hasAfter := hasJSONValue(entry.After)
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

func kubernetesChangeCounts(r io.Reader) (app.StepChangeCounts, error) {
	var plan kubernetesPlanSummary
	if err := json.NewDecoder(r).Decode(&plan); err != nil {
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

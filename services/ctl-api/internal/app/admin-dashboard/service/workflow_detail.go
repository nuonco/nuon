package service

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/admin-dashboard/service/views"
)

func (s *service) WorkflowDetail(c *gin.Context) {
	ctx := c.Request.Context()
	workflowID := c.Param("workflow_id")

	var wf app.Workflow
	res := s.db.WithContext(ctx).
		Preload("CreatedBy").
		Preload("Steps", func(db *gorm.DB) *gorm.DB {
			return db.Order("group_idx ASC, group_retry_idx ASC, idx ASC, created_at ASC")
		}).
		Preload("Steps.Approval").
		Preload("Steps.Approval.Response").
		Where("id = ?", workflowID).
		First(&wf)

	if res.Error != nil {
		s.l.Error("failed to fetch workflow", zap.Error(res.Error))
		c.JSON(http.StatusNotFound, gin.H{"error": "Workflow not found"})
		return
	}

	// Build enriched step details
	stepDetails := make([]views.StepDetailData, len(wf.Steps))
	for i, step := range wf.Steps {
		stepDetails[i] = views.StepDetailData{
			Step: &wf.Steps[i],
		}

		// Format queue signal if present
		if step.QueueSignal != nil {
			data, err := json.MarshalIndent(step.QueueSignal, "", "  ")
			if err == nil {
				stepDetails[i].QueueSignalJSON = string(data)
			}
		}

		// Load the actual QueueSignal record for this step (for linking)
		var stepSignal app.QueueSignal
		if err := s.db.WithContext(ctx).
			Where(app.QueueSignal{OwnerID: step.ID, OwnerType: (&app.WorkflowStep{}).TableName()}).
			First(&stepSignal).Error; err == nil {
			stepDetails[i].StepSignalID = stepSignal.ID
			stepDetails[i].StepSignalQueueID = stepSignal.QueueID
		}

		// Load step target
		if step.StepTargetID != "" {
			stepDetails[i].StepTarget = s.loadStepTarget(c, step.StepTargetID, step.StepTargetType)
		}
	}

	// Load step groups for the workflow (with their queue signals)
	var stepGroups []app.WorkflowStepGroup
	s.db.WithContext(ctx).
		Preload("QueueSignal").
		Where(app.WorkflowStepGroup{WorkflowID: workflowID}).
		Order("group_idx ASC").
		Find(&stepGroups)

	// Organize steps into groups
	groupDetails := s.buildGroupDetails(stepGroups, stepDetails, wf.Steps)

	// Load the generate-steps queue signal if the workflow has one
	var generateStepsSignal *app.QueueSignal
	if wf.GenerateStepsSignal != nil {
		var qs app.QueueSignal
		if err := s.db.WithContext(ctx).
			Preload("Queue").
			Where(app.QueueSignal{
				OwnerID:   wf.ID,
				OwnerType: "install_workflows",
			}).
			First(&qs).Error; err == nil {
			generateStepsSignal = &qs
		}
	}

	// Load the execute-workflow signal for this workflow (the signal that triggers execution)
	var workflowSignal *app.QueueSignal
	var ws app.QueueSignal
	if err := s.db.WithContext(ctx).
		Where("owner_id = ? AND type = ?", wf.ID, "execute-workflow").
		First(&ws).Error; err == nil {
		workflowSignal = &ws
	}

	c.JSON(http.StatusOK, gin.H{
		"workflow":              &wf,
		"group_details":         groupDetails,
		"generate_steps_signal": generateStepsSignal,
		"workflow_signal":       workflowSignal,
	})
}

func (s *service) loadStepTarget(c *gin.Context, targetID, targetType string) *views.StepTargetData {
	ctx := c.Request.Context()
	target := &views.StepTargetData{
		ID:   targetID,
		Type: targetType,
	}

	switch app.WorkflowStepTargetType(targetType) {
	case app.WorkflowStepTargetTypeInstallDeploy:
		var deploy app.InstallDeploy
		if err := s.db.WithContext(ctx).Preload("LogStream").Where("id = ?", targetID).First(&deploy).Error; err == nil {
			target.Status = string(deploy.Status)
			if deploy.LogStream.ID != "" {
				target.LogStreamID = deploy.LogStream.ID
			}
		}
	case app.WorkflowStepTargetTypeInstallSandboxRun:
		var run app.InstallSandboxRun
		if err := s.db.WithContext(ctx).Preload("LogStream").Where("id = ?", targetID).First(&run).Error; err == nil {
			target.Status = string(run.Status)
			if run.LogStream.ID != "" {
				target.LogStreamID = run.LogStream.ID
			}
		}
	case app.WorkflowStepTargetTypeInstallActionWorkflowRun:
		var run app.InstallActionWorkflowRun
		if err := s.db.WithContext(ctx).Preload("LogStream").Where("id = ?", targetID).First(&run).Error; err == nil {
			target.Status = string(run.Status)
			if run.LogStream.ID != "" {
				target.LogStreamID = run.LogStream.ID
			}
		}
	default:
		// For other types, just show the ID
	}

	return target
}

// buildGroupDetails organizes step details into groups. When step group records
// exist, steps are matched by WorkflowStepGroupID. Otherwise, synthetic groups
// are created from step GroupIdx values for backward compatibility.
func (s *service) buildGroupDetails(stepGroups []app.WorkflowStepGroup, stepDetails []views.StepDetailData, steps []app.WorkflowStep) []views.GroupDetailData {
	if len(stepGroups) > 0 {
		return s.buildGroupDetailsFromRecords(stepGroups, stepDetails)
	}
	return s.buildSyntheticGroupDetails(stepDetails, steps)
}

func (s *service) buildGroupDetailsFromRecords(stepGroups []app.WorkflowStepGroup, stepDetails []views.StepDetailData) []views.GroupDetailData {
	// Index step details by step group ID
	groupSteps := make(map[string][]views.StepDetailData)
	var ungrouped []views.StepDetailData

	for _, sd := range stepDetails {
		gID := sd.Step.WorkflowStepGroupID
		if gID != "" {
			groupSteps[gID] = append(groupSteps[gID], sd)
		} else {
			ungrouped = append(ungrouped, sd)
		}
	}

	var result []views.GroupDetailData
	for i := range stepGroups {
		result = append(result, views.GroupDetailData{
			Group: &stepGroups[i],
			Steps: groupSteps[stepGroups[i].ID],
		})
	}

	// Append any ungrouped steps as a synthetic group
	if len(ungrouped) > 0 {
		result = append(result, views.GroupDetailData{
			Group: &app.WorkflowStepGroup{Name: "Ungrouped"},
			Steps: ungrouped,
		})
	}

	return result
}

func (s *service) buildSyntheticGroupDetails(stepDetails []views.StepDetailData, steps []app.WorkflowStep) []views.GroupDetailData {
	// Collect unique group indices in order
	type groupInfo struct {
		idx      int
		parallel bool
	}
	seen := make(map[int]bool)
	var groups []groupInfo

	for _, step := range steps {
		if !seen[step.GroupIdx] {
			seen[step.GroupIdx] = true
			groups = append(groups, groupInfo{idx: step.GroupIdx, parallel: step.GroupParallel})
		}
	}

	// Build step detail map by group index
	groupStepDetails := make(map[int][]views.StepDetailData)
	for _, sd := range stepDetails {
		groupStepDetails[sd.Step.GroupIdx] = append(groupStepDetails[sd.Step.GroupIdx], sd)
	}

	var result []views.GroupDetailData
	for _, g := range groups {
		result = append(result, views.GroupDetailData{
			Group: &app.WorkflowStepGroup{
				GroupIdx: g.idx,
				Parallel: g.parallel,
			},
			Steps: groupStepDetails[g.idx],
		})
	}

	return result
}

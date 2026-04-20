package service

import (
	"encoding/json"
	"net/http"

	"github.com/a-h/templ"
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

	// Load step targets and their log streams
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

		// Load step target
		if step.StepTargetID != "" {
			stepDetails[i].StepTarget = s.loadStepTarget(c, step.StepTargetID, step.StepTargetType)
		}
	}

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

	component := views.WorkflowDetail(&wf, stepDetails, generateStepsSignal)
	templ.Handler(component).ServeHTTP(c.Writer, c.Request)
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

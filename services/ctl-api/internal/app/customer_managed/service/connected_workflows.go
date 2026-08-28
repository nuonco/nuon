package service

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	installshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/workflowstepapprovalresponse"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
	flowclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/client"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

func (s *service) AgentListWorkflows(ctx *gin.Context) {
	agent, ok := s.connectedAgent(ctx, "workflows")
	if !ok {
		return
	}
	var workflows []app.Workflow
	if err := s.db.WithContext(ctx).Where(app.Workflow{OwnerID: agent.Install.ID, OwnerType: "installs"}).Order("created_at DESC").Limit(100).Find(&workflows).Error; err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, workflows)
}

func (s *service) AgentGetWorkflow(ctx *gin.Context) {
	agent, ok := s.connectedAgent(ctx, "workflows")
	if !ok {
		return
	}
	workflow, err := s.agentWorkflow(ctx, agent, ctx.Param("workflow_id"))
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, workflow)
}

func (s *service) AgentGetWorkflowStepLogs(ctx *gin.Context) {
	agent, ok := s.connectedAgent(ctx, "workflow-logs")
	if !ok {
		return
	}
	workflow, err := s.agentWorkflow(ctx, agent, ctx.Param("workflow_id"))
	if err != nil {
		ctx.Error(err)
		return
	}
	var step *app.WorkflowStep
	for i := range workflow.Steps {
		if workflow.Steps[i].ID == ctx.Param("step_id") {
			step = &workflow.Steps[i]
			break
		}
	}
	if step == nil {
		ctx.Status(http.StatusNotFound)
		return
	}
	if step.LogStream == nil || step.LogStream.ID == "" {
		ctx.JSON(http.StatusOK, []app.OtelLogRecord{})
		return
	}
	var logs []app.OtelLogRecord
	if err := s.chDB.WithContext(ctx).
		Where(app.OtelLogRecord{OrgID: agent.Install.OrgID, LogStreamID: step.LogStream.ID}).
		Order("timestamp ASC").
		Limit(500).
		Find(&logs).Error; err != nil {
		ctx.Error(fmt.Errorf("read workflow step logs: %w", err))
		return
	}
	ctx.JSON(http.StatusOK, logs)
}

func (s *service) AgentRetryWorkflowStep(ctx *gin.Context) {
	agent, ok := s.connectedAgent(ctx, "release-deployments")
	if !ok {
		return
	}
	if agent.Policy.ApprovalAuthority != app.InstallAuthorityCustomer {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "management policy does not grant customer approval authority"})
		return
	}
	workflow, err := s.agentWorkflow(ctx, agent, ctx.Param("workflow_id"))
	if err != nil {
		ctx.Error(err)
		return
	}
	stepID := ctx.Param("step_id")
	var step *app.WorkflowStep
	for i := range workflow.Steps {
		if workflow.Steps[i].ID == stepID {
			step = &workflow.Steps[i]
			break
		}
	}
	if step == nil {
		ctx.Status(http.StatusNotFound)
		return
	}
	response, err := s.flowsClient.RetryStep(ctx, &flowclient.RetryStepRequest{
		InstallWorkflowID: workflow.ID,
		StepID:            step.ID,
	})
	if err != nil {
		ctx.Error(fmt.Errorf("retry workflow step: %w", err))
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"workflow_id": workflow.ID, "retryable": response.Retryable})
}

func (s *service) AgentGetApprovalContents(ctx *gin.Context) {
	agent, ok := s.connectedAgent(ctx, "workflow-approvals")
	if !ok {
		return
	}
	_, step, approval, err := s.agentApproval(ctx, agent)
	if err != nil {
		ctx.Error(err)
		return
	}
	if step.ID != ctx.Param("step_id") || approval.ID != ctx.Param("approval_id") {
		ctx.Status(http.StatusNotFound)
		return
	}
	blobReadEnabled := s.cfg != nil && s.cfg.BlobReadEnabled
	contents, _ := approval.GetContents(blobstore.WithBlobService(ctx, s.blobSvc), blobReadEnabled)
	var compressed bytes.Buffer
	writer := gzip.NewWriter(&compressed)
	if _, err := writer.Write([]byte(contents)); err != nil {
		ctx.Error(err)
		return
	}
	if err := writer.Close(); err != nil {
		ctx.Error(err)
		return
	}
	ctx.Header("Content-Encoding", "gzip")
	ctx.Data(http.StatusOK, "application/json", compressed.Bytes())
}

type agentApprovalResponseRequest struct {
	ResponseType app.WorkflowStepResponseType `json:"response_type"`
	Note         string                       `json:"note"`
}

func (s *service) AgentCreateApprovalResponse(ctx *gin.Context) {
	agent, ok := s.connectedAgent(ctx, "workflow-approvals")
	if !ok {
		return
	}
	if agent.Policy.ApprovalAuthority != app.InstallAuthorityCustomer {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "management policy does not grant customer approval authority"})
		return
	}
	var req agentApprovalResponseRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.ResponseType != app.WorkflowStepApprovalResponseTypeApprove && req.ResponseType != app.WorkflowStepApprovalResponseTypeDeny && req.ResponseType != app.WorkflowStepApprovalResponseTypeSkipCurrent && req.ResponseType != app.WorkflowStepApprovalResponseTypeSkipCurrentAndDependents {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "unsupported approval response"})
		return
	}
	workflow, step, approval, err := s.agentApproval(ctx, agent)
	if err != nil {
		ctx.Error(err)
		return
	}
	if approval.Response != nil {
		ctx.JSON(http.StatusConflict, gin.H{"error": "workflow step approval already has a response"})
		return
	}
	requestCtx, err := s.connectedWorkflowContext(ctx, agent)
	if err != nil {
		ctx.Error(err)
		return
	}
	response := app.WorkflowStepApprovalResponse{InstallWorkflowStepApprovalID: approval.ID, Type: req.ResponseType, Note: req.Note}
	if err := s.db.WithContext(requestCtx).Create(&response).Error; err != nil {
		ctx.Error(err)
		return
	}
	signal := &workflowstepapprovalresponse.Signal{
		InstallID: agent.Install.ID, InstallWorkflowID: workflow.ID, WorkflowStepID: step.ID,
		ApprovalID: approval.ID, ApprovalResponseID: response.ID, ResponseType: response.Type,
	}
	queue, err := s.queueClient.GetQueueByOwnerAndName(requestCtx, agent.Install.ID, "installs", installshelpers.InstallSignalsQueueName)
	if err != nil {
		ctx.Error(err)
		return
	}
	if _, err := s.queueClient.EnqueueSignal(requestCtx, &queueclient.EnqueueSignalRequest{
		QueueID: queue.ID, Signal: signal, OwnerID: response.ID, OwnerType: "workflow_step_approval_responses",
	}); err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusCreated, response)
}

func (s *service) agentWorkflow(ctx context.Context, agent *agentContext, workflowID string) (*app.Workflow, error) {
	var workflow app.Workflow
	err := s.db.WithContext(ctx).
		Preload("CreatedBy").
		Preload("Steps", func(db *gorm.DB) *gorm.DB { return db.Order("group_idx, group_retry_idx, idx, created_at ASC") }).
		Preload("Steps.CreatedBy").
		Preload("Steps.Approval", func(db *gorm.DB) *gorm.DB { return db.Omit("contents") }).
		Preload("Steps.Approval.Response").
		Preload("StepGroups", func(db *gorm.DB) *gorm.DB { return db.Order("group_idx ASC") }).
		Where(app.Workflow{ID: workflowID, OrgID: agent.Install.OrgID, OwnerID: agent.Install.ID, OwnerType: "installs"}).First(&workflow).Error
	if err != nil {
		return &workflow, err
	}
	if err := s.loadAgentStepLogStreams(ctx, agent.Install.OrgID, workflow.Steps); err != nil {
		return &workflow, err
	}
	return &workflow, nil
}

func (s *service) loadAgentStepLogStreams(ctx context.Context, orgID string, steps []app.WorkflowStep) error {
	ownerIDs := make([]string, 0, len(steps)*2)
	for i := range steps {
		ownerIDs = append(ownerIDs, steps[i].ID)
		if steps[i].StepTargetID != "" {
			ownerIDs = append(ownerIDs, steps[i].StepTargetID)
		}
	}
	if len(ownerIDs) == 0 {
		return nil
	}
	streams := make([]app.LogStream, 0, len(ownerIDs))
	for _, ownerID := range ownerIDs {
		var stream app.LogStream
		err := s.db.WithContext(ctx).
			Where(app.LogStream{OrgID: orgID, OwnerID: ownerID}).
			Order("created_at DESC").
			First(&stream).Error
		if err == nil {
			streams = append(streams, stream)
		} else if err != gorm.ErrRecordNotFound {
			return fmt.Errorf("load log stream for workflow step owner %s: %w", ownerID, err)
		}
	}
	byOwner := make(map[string]*app.LogStream, len(streams))
	for i := range streams {
		if byOwner[streams[i].OwnerID] == nil {
			byOwner[streams[i].OwnerID] = &streams[i]
		}
	}
	for i := range steps {
		if stream := byOwner[steps[i].StepTargetID]; stream != nil {
			steps[i].LogStream = stream
		} else if stream := byOwner[steps[i].ID]; stream != nil {
			steps[i].LogStream = stream
		}
	}
	return nil
}

func (s *service) agentApproval(ctx *gin.Context, agent *agentContext) (*app.Workflow, *app.WorkflowStep, *app.WorkflowStepApproval, error) {
	workflow, err := s.agentWorkflow(ctx, agent, ctx.Param("workflow_id"))
	if err != nil {
		return nil, nil, nil, err
	}
	for i := range workflow.Steps {
		step := &workflow.Steps[i]
		if step.ID == ctx.Param("step_id") && step.Approval != nil && step.Approval.ID == ctx.Param("approval_id") {
			return workflow, step, step.Approval, nil
		}
	}
	return nil, nil, nil, gorm.ErrRecordNotFound
}

func (s *service) connectedWorkflowContext(ctx context.Context, agent *agentContext) (context.Context, error) {
	return ctx, nil
}

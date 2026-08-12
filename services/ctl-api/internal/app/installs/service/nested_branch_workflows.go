package service

import (
	"github.com/gin-gonic/gin"
)

// The handlers below delegate to their bare-route counterparts. They exist so
// each ancestor-scoped workflow path gets its own swagger operation, and so a
// generated client method, rather than being reachable only by hand-built
// requests. Workflow ownership is enforced by each group's guard, not here.

// @ID						GetWorkflowByAppBranch
// @Summary					get a workflow
// @Description			Ancestor-scoped alternative to the bare route: authorizes against the app branch that owns the workflow, so an identity scoped to that app can reach it.
// @Param					app_id	path	string	true	"app ID"
// @Param					app_branch_id	path	string	true	"app branch ID"
// @Param					workflow_id path	string	true	"workflow ID"
// @Tags					installs
// @Accept					json
// @Produce					json
// @Security				APIKey
// @Security				OrgID
// @Failure					400	{object}	stderr.ErrResponse
// @Failure					401	{object}	stderr.ErrResponse
// @Failure					403	{object}	stderr.ErrResponse
// @Failure					404	{object}	stderr.ErrResponse
// @Failure					500	{object}	stderr.ErrResponse
// @Success					200	{object}	app.Workflow
// @Router					/v1/apps/{app_id}/branches/{app_branch_id}/workflows/{workflow_id} [get]
func (s *service) GetWorkflowByAppBranch(ctx *gin.Context) {
	s.GetWorkflow(ctx)
}

// @ID						UpdateWorkflowByAppBranch
// @Summary					update a workflow
// @Description			Ancestor-scoped alternative to the bare route: authorizes against the app branch that owns the workflow, so an identity scoped to that app can reach it.
// @Param					app_id	path	string	true	"app ID"
// @Param					app_branch_id	path	string	true	"app branch ID"
// @Param					workflow_id path	string	true	"workflow ID"
// @Param					req			body	UpdateWorkflowRequest	true	"Input"
// @Tags					installs
// @Accept					json
// @Produce					json
// @Security				APIKey
// @Security				OrgID
// @Failure					400	{object}	stderr.ErrResponse
// @Failure					401	{object}	stderr.ErrResponse
// @Failure					403	{object}	stderr.ErrResponse
// @Failure					404	{object}	stderr.ErrResponse
// @Failure					500	{object}	stderr.ErrResponse
// @Success					200	{object}	app.Workflow
// @Router					/v1/apps/{app_id}/branches/{app_branch_id}/workflows/{workflow_id} [PATCH]
func (s *service) UpdateWorkflowByAppBranch(ctx *gin.Context) {
	s.UpdateWorkflow(ctx)
}

// @ID						CancelWorkflowByAppBranch
// @Summary						cancel an ongoing workflow
// @Description			Ancestor-scoped alternative to the bare route: authorizes against the app branch that owns the workflow, so an identity scoped to that app can reach it.
// @Param					app_id	path	string	true	"app ID"
// @Param					app_branch_id	path	string	true	"app branch ID"
// @Param workflow_id	path	string true "workflow ID"
// @Tags						installs
// @Accept						json
// @Produce						json
// @Security					APIKey
// @Security					OrgID
// @Failure						400	{object}	stderr.ErrResponse
// @Failure						401	{object}	stderr.ErrResponse
// @Failure						403	{object}	stderr.ErrResponse
// @Failure						404	{object}	stderr.ErrResponse
// @Failure						500	{object}	stderr.ErrResponse
// @Success				202	{object}	app.EmptyResponse
// @Router					/v1/apps/{app_id}/branches/{app_branch_id}/workflows/{workflow_id}/cancel [POST]
func (s *service) CancelWorkflowByAppBranch(ctx *gin.Context) {
	s.CancelWorkflow(ctx)
}

// @ID						GetWorkflowQueuePositionByAppBranch
// @Summary					get queue position for a workflow
// @Description			Ancestor-scoped alternative to the bare route: authorizes against the app branch that owns the workflow, so an identity scoped to that app can reach it.
// @Param					app_id	path	string	true	"app ID"
// @Param					app_branch_id	path	string	true	"app branch ID"
// @Param					workflow_id	path	string	true	"workflow ID"
// @Tags					installs
// @Accept					json
// @Produce					json
// @Security				APIKey
// @Security				OrgID
// @Failure					400	{object}	stderr.ErrResponse
// @Failure					401	{object}	stderr.ErrResponse
// @Failure					403	{object}	stderr.ErrResponse
// @Failure					404	{object}	stderr.ErrResponse
// @Failure					500	{object}	stderr.ErrResponse
// @Success					200	{object}	WorkflowQueuePositionResponse
// @Router					/v1/apps/{app_id}/branches/{app_branch_id}/workflows/{workflow_id}/queue-position [get]
func (s *service) GetWorkflowQueuePositionByAppBranch(ctx *gin.Context) {
	s.GetWorkflowQueuePosition(ctx)
}

// @ID						GetWorkflowStepGroupsByAppBranch
// @Summary					get all step groups for a workflow
// @Description			Ancestor-scoped alternative to the bare route: authorizes against the app branch that owns the workflow, so an identity scoped to that app can reach it.
// @Param					app_id	path	string	true	"app ID"
// @Param					app_branch_id	path	string	true	"app branch ID"
// @Param					workflow_id	path	string	true	"workflow ID"
// @Tags					installs
// @Accept					json
// @Produce					json
// @Security				APIKey
// @Security				OrgID
// @Failure					400	{object}	stderr.ErrResponse
// @Failure					401	{object}	stderr.ErrResponse
// @Failure					403	{object}	stderr.ErrResponse
// @Failure					404	{object}	stderr.ErrResponse
// @Failure					500	{object}	stderr.ErrResponse
// @Success					200	{array}		app.WorkflowStepGroup
// @Router					/v1/apps/{app_id}/branches/{app_branch_id}/workflows/{workflow_id}/step-groups [get]
func (s *service) GetWorkflowStepGroupsByAppBranch(ctx *gin.Context) {
	s.GetWorkflowStepGroups(ctx)
}

// @ID						GetWorkflowStepGroupByAppBranch
// @Summary					get a workflow step group
// @Description			Ancestor-scoped alternative to the bare route: authorizes against the app branch that owns the workflow, so an identity scoped to that app can reach it.
// @Param					app_id	path	string	true	"app ID"
// @Param					app_branch_id	path	string	true	"app branch ID"
// @Param					workflow_id		path	string	true	"workflow ID"
// @Param					step_group_id	path	string	true	"step group ID"
// @Tags					installs
// @Accept					json
// @Produce					json
// @Security				APIKey
// @Security				OrgID
// @Failure					400	{object}	stderr.ErrResponse
// @Failure					401	{object}	stderr.ErrResponse
// @Failure					403	{object}	stderr.ErrResponse
// @Failure					404	{object}	stderr.ErrResponse
// @Failure					500	{object}	stderr.ErrResponse
// @Success					200	{object}	app.WorkflowStepGroup
// @Router					/v1/apps/{app_id}/branches/{app_branch_id}/workflows/{workflow_id}/step-groups/{{step_group_id}} [get]
func (s *service) GetWorkflowStepGroupByAppBranch(ctx *gin.Context) {
	s.GetWorkflowStepGroup(ctx)
}

// @ID						GetWorkflowStepsByAppBranch
// @Summary						get all of the steps for a given workflow
// @Description			Ancestor-scoped alternative to the bare route: authorizes against the app branch that owns the workflow, so an identity scoped to that app can reach it.
// @Param					app_id	path	string	true	"app ID"
// @Param					app_branch_id	path	string	true	"app branch ID"
// @Param workflow_id	path	string true "workflow ID"
// @Param				offset	query	int		false	"offset of results to return"	Default(0)
// @Param				limit	query	int		false	"limit of results to return"	Default(10)
// @Param				page	query	int		false	"page number of results to return"	Default(0)
// @Tags						installs
// @Accept						json
// @Produce						json
// @Security					APIKey
// @Security					OrgID
// @Failure						400	{object}	stderr.ErrResponse
// @Failure						401	{object}	stderr.ErrResponse
// @Failure						403	{object}	stderr.ErrResponse
// @Failure						404	{object}	stderr.ErrResponse
// @Failure						500	{object}	stderr.ErrResponse
// @Success						200	{array}		app.WorkflowStep
// @Router					/v1/apps/{app_id}/branches/{app_branch_id}/workflows/{workflow_id}/steps [get]
func (s *service) GetWorkflowStepsByAppBranch(ctx *gin.Context) {
	s.GetWorkflowSteps(ctx)
}

// @ID						GetWorkflowStepByAppBranch
// @Summary					get a workflow step
// @Description			Ancestor-scoped alternative to the bare route: authorizes against the app branch that owns the workflow, so an identity scoped to that app can reach it.
// @Param					app_id	path	string	true	"app ID"
// @Param					app_branch_id	path	string	true	"app branch ID"
// @Param					workflow_id		path	string	true	"workflow id"
// @Param					step_id		path	string	true	"step id"
// @Tags					installs
// @Accept					json
// @Produce					json
// @Security				APIKey
// @Security				OrgID
// @Failure					400	{object}	stderr.ErrResponse
// @Failure					401	{object}	stderr.ErrResponse
// @Failure					403	{object}	stderr.ErrResponse
// @Failure					404	{object}	stderr.ErrResponse
// @Failure					500	{object}	stderr.ErrResponse
// @Success					200	{object}	app.WorkflowStep
// @Router					/v1/apps/{app_id}/branches/{app_branch_id}/workflows/{workflow_id}/steps/{{step_id}} [get]
func (s *service) GetWorkflowStepByAppBranch(ctx *gin.Context) {
	s.GetWorkflowStep(ctx)
}

// @ID						AwaitWorkflowStepByAppBranch
// @Summary					long-poll for workflow step completion
// @Description			Ancestor-scoped alternative to the bare route: authorizes against the app branch that owns the workflow, so an identity scoped to that app can reach it.
// @Param					app_id	path	string	true	"app ID"
// @Param					app_branch_id	path	string	true	"app branch ID"
// @Param					workflow_id	path	string	true	"workflow id"
// @Param					step_id		path	string	true	"step id"
// @Tags					installs
// @Accept					json
// @Produce					json
// @Security				APIKey
// @Security				OrgID
// @Failure					400	{object}	stderr.ErrResponse
// @Failure					401	{object}	stderr.ErrResponse
// @Failure					404	{object}	stderr.ErrResponse
// @Failure					408	{object}	stderr.ErrResponse
// @Failure					500	{object}	stderr.ErrResponse
// @Success					200	{object}	app.WorkflowStep
// @Router					/v1/apps/{app_id}/branches/{app_branch_id}/workflows/{workflow_id}/steps/{{step_id}}/await [get]
func (s *service) AwaitWorkflowStepByAppBranch(ctx *gin.Context) {
	s.AwaitWorkflowStep(ctx)
}

// @ID						RetryWorkflowStepByAppBranch
// @Summary					retry a failed or awaiting-approval workflow step
// @Description			Ancestor-scoped alternative to the bare route: authorizes against the app branch that owns the workflow, so an identity scoped to that app can reach it.
// @Param					app_id	path	string	true	"app ID"
// @Param					app_branch_id	path	string	true	"app branch ID"
// @Param					workflow_id	path	string	true	"workflow ID"
// @Param					step_id		path	string	true	"step ID"
// @Tags					installs
// @Accept					json
// @Produce					json
// @Security				APIKey
// @Security				OrgID
// @Failure					400	{object}	stderr.ErrResponse
// @Failure					401	{object}	stderr.ErrResponse
// @Failure					403	{object}	stderr.ErrResponse
// @Failure					404	{object}	stderr.ErrResponse
// @Failure					500	{object}	stderr.ErrResponse
// @Success					201	{object}	RetryWorkflowStepResponse
// @Router					/v1/apps/{app_id}/branches/{app_branch_id}/workflows/{workflow_id}/steps/{{step_id}}/retry [POST]
func (s *service) RetryWorkflowStepByAppBranch(ctx *gin.Context) {
	s.RetryWorkflowStep(ctx)
}

// @ID						SkipWorkflowStepByAppBranch
// @Summary					skip a failed workflow step and continue the workflow
// @Description			Ancestor-scoped alternative to the bare route: authorizes against the app branch that owns the workflow, so an identity scoped to that app can reach it.
// @Param					app_id	path	string	true	"app ID"
// @Param					app_branch_id	path	string	true	"app branch ID"
// @Param					workflow_id	path	string	true	"workflow ID"
// @Param					step_id		path	string	true	"step ID"
// @Tags					installs
// @Accept					json
// @Produce					json
// @Security				APIKey
// @Security				OrgID
// @Failure					400	{object}	stderr.ErrResponse
// @Failure					401	{object}	stderr.ErrResponse
// @Failure					403	{object}	stderr.ErrResponse
// @Failure					404	{object}	stderr.ErrResponse
// @Failure					500	{object}	stderr.ErrResponse
// @Success					201	{object}	SkipWorkflowStepResponse
// @Router					/v1/apps/{app_id}/branches/{app_branch_id}/workflows/{workflow_id}/steps/{{step_id}}/skip [POST]
func (s *service) SkipWorkflowStepByAppBranch(ctx *gin.Context) {
	s.SkipWorkflowStep(ctx)
}

// @ID						CancelWorkflowStepByAppBranch
// @Summary					cancel an in-progress workflow step
// @Description			Ancestor-scoped alternative to the bare route: authorizes against the app branch that owns the workflow, so an identity scoped to that app can reach it.
// @Param					app_id	path	string	true	"app ID"
// @Param					app_branch_id	path	string	true	"app branch ID"
// @Param					workflow_id	path	string	true	"workflow ID"
// @Param					step_id		path	string	true	"step ID"
// @Tags					installs
// @Accept					json
// @Produce					json
// @Security				APIKey
// @Security				OrgID
// @Failure					400	{object}	stderr.ErrResponse
// @Failure					401	{object}	stderr.ErrResponse
// @Failure					403	{object}	stderr.ErrResponse
// @Failure					404	{object}	stderr.ErrResponse
// @Failure					500	{object}	stderr.ErrResponse
// @Success					202	{object}	CancelWorkflowStepResponse
// @Router					/v1/apps/{app_id}/branches/{app_branch_id}/workflows/{workflow_id}/steps/{{step_id}}/cancel [POST]
func (s *service) CancelWorkflowStepByAppBranch(ctx *gin.Context) {
	s.CancelWorkflowStep(ctx)
}

// @ID						GetWorkflowStepApprovalByAppBranch
// @Summary								get an workflow step approval
// @Description			Ancestor-scoped alternative to the bare route: authorizes against the app branch that owns the workflow, so an identity scoped to that app can reach it.
// @Param					app_id	path	string	true	"app ID"
// @Param					app_branch_id	path	string	true	"app branch ID"
// @Param	workflow_id			path	string	true	"workflow id"
// @Param	step_id	path	string	true	"step id"
// @Param	approval_id					path	string	true	"approval id"
// @Tags								installs
// @Accept								json
// @Produce								json
// @Security							APIKey
// @Security							OrgID
// @Failure								400	{object}	stderr.ErrResponse
// @Failure								401	{object}	stderr.ErrResponse
// @Failure								403	{object}	stderr.ErrResponse
// @Failure								404	{object}	stderr.ErrResponse
// @Failure								500	{object}	stderr.ErrResponse
// @Success								200	{object}		app.WorkflowStepApproval
// @Router					/v1/apps/{app_id}/branches/{app_branch_id}/workflows/{workflow_id}/steps/{{step_id}}/approvals/{{approval_id}} [get]
func (s *service) GetWorkflowStepApprovalByAppBranch(ctx *gin.Context) {
	s.GetWorkflowStepApproval(ctx)
}

// @ID						CreateWorkflowStepApprovalResponseByAppBranch
// @Summary					Create an approval response for a workflow step.
// @Description			Ancestor-scoped alternative to the bare route: authorizes against the app branch that owns the workflow, so an identity scoped to that app can reach it.
// @Param					app_id	path	string	true	"app ID"
// @Param					app_branch_id	path	string	true	"app branch ID"
// @Param					workflow_id			path	string	true	"workflow id"
// @Param					step_id	path	string	true	"step id"
// @Param					approval_id			path	string	true	"approval id"
// @Param					req					body	CreateWorkflowStepApprovalResponseRequest	true	"Input"
// @Tags					installs
// @Accept					json
// @Produce					json
// @Security				APIKey
// @Security				OrgID
// @Failure					400	{object}	stderr.ErrResponse
// @Failure					401	{object}	stderr.ErrResponse
// @Failure					403	{object}	stderr.ErrResponse
// @Failure					404	{object}	stderr.ErrResponse
// @Failure					409	{object}	stderr.ErrResponse
// @Failure					500	{object}	stderr.ErrResponse
// @Success					201	{object}	CreateWorkflowStepApprovalResponseResponse
// @Router					/v1/apps/{app_id}/branches/{app_branch_id}/workflows/{workflow_id}/steps/{{step_id}}/approvals/{{approval_id}}/response [POST]
func (s *service) CreateWorkflowStepApprovalResponseByAppBranch(ctx *gin.Context) {
	s.CreateWorkflowStepApprovalResponse(ctx)
}

// @ID						GetWorkflowStepApprovalContentsByAppBranch
// @Summary				get a workflow step approval contents
// @Description			Ancestor-scoped alternative to the bare route: authorizes against the app branch that owns the workflow, so an identity scoped to that app can reach it.
// @Param					app_id	path	string	true	"app ID"
// @Param					app_branch_id	path	string	true	"app branch ID"
// @Param					workflow_id			path	string	true	"workflow id"
// @Param					step_id	path	string	true	"step id"
// @Param					approval_id			path	string	true	"approval id"
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
// @Success				200	{object}	interface{}
// @Header					200	{string}	Content-Encoding	"gzip"
// @Router					/v1/apps/{app_id}/branches/{app_branch_id}/workflows/{workflow_id}/steps/{{step_id}}/approvals/{{approval_id}}/contents [get]
func (s *service) GetWorkflowStepApprovalContentsByAppBranch(ctx *gin.Context) {
	s.GetWorkflowStepApprovalContents(ctx)
}

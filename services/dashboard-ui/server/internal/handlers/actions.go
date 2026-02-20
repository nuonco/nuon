package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/sdks/nuon-go/models"
	"github.com/nuonco/nuon/services/dashboard-ui/server/internal/pkg/cctx"
)

type ActionsHandler struct {
	l *zap.Logger
}

func NewActionsHandler(l *zap.Logger) *ActionsHandler {
	return &ActionsHandler{l: l}
}

func (h *ActionsHandler) RegisterRoutes(e *gin.Engine) error {
	// App mutations
	e.POST("/api/actions/apps/build-component", h.BuildComponent)
	e.POST("/api/actions/apps/create-app-install", h.CreateAppInstall)

	// Install mutations
	e.POST("/api/actions/installs/deploy-component", h.DeployComponent)
	e.POST("/api/actions/installs/deprovision-install", h.DeprovisionInstall)
	e.POST("/api/actions/installs/reprovision-install", h.ReprovisionInstall)
	e.POST("/api/actions/installs/forget-install", h.ForgetInstall)
	e.POST("/api/actions/installs/teardown-component", h.TeardownComponent)
	e.POST("/api/actions/installs/teardown-components", h.TeardownComponents)
	e.POST("/api/actions/installs/deploy-components", h.DeployComponents)
	e.POST("/api/actions/installs/update-install", h.UpdateInstall)

	// Workflow mutations
	e.POST("/api/actions/workflows/cancel-workflow", h.CancelWorkflow)
	e.POST("/api/actions/workflows/approve-workflow-step", h.ApproveWorkflowStep)
	e.POST("/api/actions/workflows/retry-workflow-step", h.RetryWorkflowStep)

	// Org mutations
	e.POST("/api/actions/orgs/create-org", h.CreateOrg)

	// VCS mutations
	e.POST("/api/actions/vcs/create-connection", h.CreateVCSConnection)
	e.POST("/api/actions/vcs/delete-connection", h.DeleteVCSConnection)

	return nil
}

// helper to bind JSON and set org on client
func (h *ActionsHandler) clientWithOrg(c *gin.Context, orgID string) error {
	client, err := cctx.APIClientFromGinContext(c)
	if err != nil {
		return err
	}
	client.SetOrgID(orgID)
	return nil
}

// --- App mutations ---

type buildComponentReq struct {
	ComponentID string `json:"componentId"`
	OrgID       string `json:"orgId"`
}

func (h *ActionsHandler) BuildComponent(c *gin.Context) {
	var req buildComponentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	client, err := cctx.APIClientFromGinContext(c)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}
	client.SetOrgID(req.OrgID)

	build, err := client.CreateComponentBuild(c.Request.Context(), req.ComponentID, &models.ServiceCreateComponentBuildRequest{})
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}

	respondJSON(c, http.StatusOK, build)
}

type createAppInstallReq struct {
	AppID string `json:"appId"`
	OrgID string `json:"orgId"`
	Name  string `json:"name"`
}

func (h *ActionsHandler) CreateAppInstall(c *gin.Context) {
	var req createAppInstallReq
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	client, err := cctx.APIClientFromGinContext(c)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}
	client.SetOrgID(req.OrgID)

	name := req.Name
	install, _, err := client.CreateInstall(c.Request.Context(), req.AppID, &models.ServiceCreateInstallRequest{
		Name: &name,
	})
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}

	respondJSON(c, http.StatusOK, install)
}

// --- Install mutations ---

type deployComponentReq struct {
	InstallID string `json:"installId"`
	OrgID     string `json:"orgId"`
	BuildID   string `json:"buildId"`
	PlanOnly  bool   `json:"planOnly"`
}

func (h *ActionsHandler) DeployComponent(c *gin.Context) {
	var req deployComponentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	client, err := cctx.APIClientFromGinContext(c)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}
	client.SetOrgID(req.OrgID)

	deploy, err := client.CreateInstallDeploy(c.Request.Context(), req.InstallID, &models.ServiceCreateInstallDeployRequest{
		BuildID: req.BuildID,
	})
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}

	respondJSON(c, http.StatusOK, deploy)
}

type installIDOrgReq struct {
	InstallID string `json:"installId"`
	OrgID     string `json:"orgId"`
}

func (h *ActionsHandler) DeprovisionInstall(c *gin.Context) {
	var req installIDOrgReq
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	client, err := cctx.APIClientFromGinContext(c)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}
	client.SetOrgID(req.OrgID)

	if err := client.DeprovisionInstall(c.Request.Context(), req.InstallID); err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}

	respondJSON(c, http.StatusOK, nil)
}

func (h *ActionsHandler) ReprovisionInstall(c *gin.Context) {
	var req installIDOrgReq
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	client, err := cctx.APIClientFromGinContext(c)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}
	client.SetOrgID(req.OrgID)

	if err := client.ReprovisionInstall(c.Request.Context(), req.InstallID); err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}

	respondJSON(c, http.StatusOK, nil)
}

func (h *ActionsHandler) ForgetInstall(c *gin.Context) {
	var req installIDOrgReq
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	client, err := cctx.APIClientFromGinContext(c)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}
	client.SetOrgID(req.OrgID)

	if _, err := client.ForgetInstall(c.Request.Context(), req.InstallID); err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}

	respondJSON(c, http.StatusOK, nil)
}

type teardownComponentReq struct {
	InstallID   string `json:"installId"`
	ComponentID string `json:"componentId"`
	OrgID       string `json:"orgId"`
}

func (h *ActionsHandler) TeardownComponent(c *gin.Context) {
	var req teardownComponentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	client, err := cctx.APIClientFromGinContext(c)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}
	client.SetOrgID(req.OrgID)

	if err := client.TeardownInstallComponent(c.Request.Context(), req.InstallID, req.ComponentID); err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}

	respondJSON(c, http.StatusOK, nil)
}

func (h *ActionsHandler) TeardownComponents(c *gin.Context) {
	var req installIDOrgReq
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	client, err := cctx.APIClientFromGinContext(c)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}
	client.SetOrgID(req.OrgID)

	if err := client.TeardownInstallComponents(c.Request.Context(), req.InstallID); err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}

	respondJSON(c, http.StatusOK, nil)
}

func (h *ActionsHandler) DeployComponents(c *gin.Context) {
	var req struct {
		InstallID string `json:"installId"`
		OrgID     string `json:"orgId"`
		PlanOnly  bool   `json:"planOnly"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	client, err := cctx.APIClientFromGinContext(c)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}
	client.SetOrgID(req.OrgID)

	if err := client.DeployInstallComponents(c.Request.Context(), req.InstallID, req.PlanOnly); err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}

	respondJSON(c, http.StatusOK, nil)
}

func (h *ActionsHandler) UpdateInstall(c *gin.Context) {
	var req struct {
		InstallID string                             `json:"installId"`
		OrgID     string                             `json:"orgId"`
		Body      models.ServiceUpdateInstallRequest `json:"body"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	client, err := cctx.APIClientFromGinContext(c)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}
	client.SetOrgID(req.OrgID)

	install, err := client.UpdateInstall(c.Request.Context(), req.InstallID, &req.Body)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}

	respondJSON(c, http.StatusOK, install)
}

// --- Workflow mutations ---

type cancelWorkflowReq struct {
	WorkflowID string `json:"workflowId"`
	OrgID      string `json:"orgId"`
}

func (h *ActionsHandler) CancelWorkflow(c *gin.Context) {
	var req cancelWorkflowReq
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	client, err := cctx.APIClientFromGinContext(c)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}
	client.SetOrgID(req.OrgID)

	if _, err := client.CancelWorkflow(c.Request.Context(), req.WorkflowID); err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}

	respondJSON(c, http.StatusOK, nil)
}

func (h *ActionsHandler) ApproveWorkflowStep(c *gin.Context) {
	var req struct {
		WorkflowID     string `json:"workflowId"`
		WorkflowStepID string `json:"workflowStepId"`
		ApprovalID     string `json:"approvalId"`
		OrgID          string `json:"orgId"`
		ResponseType   string `json:"responseType"`
		Note           string `json:"note"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	client, err := cctx.APIClientFromGinContext(c)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}
	client.SetOrgID(req.OrgID)

	resp, err := client.CreateWorkflowStepApprovalResponse(c.Request.Context(), req.WorkflowID, req.WorkflowStepID, req.ApprovalID, &models.ServiceCreateWorkflowStepApprovalResponseRequest{
		ResponseType: models.AppWorkflowStepResponseType(req.ResponseType),
		Note:         req.Note,
	})
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}

	respondJSON(c, http.StatusOK, resp)
}

func (h *ActionsHandler) RetryWorkflowStep(c *gin.Context) {
	var req struct {
		WorkflowID string `json:"workflowId"`
		StepID     string `json:"stepId"`
		OrgID      string `json:"orgId"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	client, err := cctx.APIClientFromGinContext(c)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}
	client.SetOrgID(req.OrgID)

	if err := client.RetryWorkflowStep(c.Request.Context(), req.WorkflowID, req.StepID, &models.ServiceRetryWorkflowStepRequest{}); err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}

	respondJSON(c, http.StatusOK, nil)
}

// --- Org mutations ---

func (h *ActionsHandler) CreateOrg(c *gin.Context) {
	var req struct {
		Name string `json:"name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	client, err := cctx.APIClientFromGinContext(c)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}

	org, err := client.CreateOrg(c.Request.Context(), &models.ServiceCreateOrgRequest{
		Name: &req.Name,
	})
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}

	respondJSON(c, http.StatusOK, org)
}

// --- VCS mutations ---

func (h *ActionsHandler) CreateVCSConnection(c *gin.Context) {
	var req struct {
		OrgID           string `json:"orgId"`
		GithubInstallID string `json:"githubInstallId"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	client, err := cctx.APIClientFromGinContext(c)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}
	client.SetOrgID(req.OrgID)

	conn, err := client.CreateVCSConnection(c.Request.Context(), &models.ServiceCreateConnectionRequest{
		GithubInstallID: &req.GithubInstallID,
	})
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}

	respondJSON(c, http.StatusOK, conn)
}

func (h *ActionsHandler) DeleteVCSConnection(c *gin.Context) {
	var req struct {
		OrgID        string `json:"orgId"`
		ConnectionID string `json:"connectionId"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	client, err := cctx.APIClientFromGinContext(c)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}
	client.SetOrgID(req.OrgID)

	if err := client.DeleteVCSConnection(c.Request.Context(), req.ConnectionID); err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}

	respondJSON(c, http.StatusOK, nil)
}

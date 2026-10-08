package service

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// InstallDeploymentType is a normalized category for a deployment record.
type InstallDeploymentType string

const (
	InstallDeploymentTypeProvision           InstallDeploymentType = "provision"
	InstallDeploymentTypeReprovision         InstallDeploymentType = "reprovision"
	InstallDeploymentTypeSandboxReprovision  InstallDeploymentType = "sandbox_reprovision"
	InstallDeploymentTypeAppBranchUpdate     InstallDeploymentType = "app_branch_update"
	InstallDeploymentTypeComponentDeploy     InstallDeploymentType = "component_deploy"
	InstallDeploymentTypeImageUpdate         InstallDeploymentType = "image_update"
	InstallDeploymentTypeStackUpdate         InstallDeploymentType = "stack_update"
	InstallDeploymentTypeInstallConfigUpdate InstallDeploymentType = "install_config_update"
)

func workflowTypeToDeploymentType(t app.WorkflowType) InstallDeploymentType {
	switch t {
	case app.WorkflowTypeProvision:
		return InstallDeploymentTypeProvision
	case app.WorkflowTypeManualDeploy,
		app.WorkflowTypeDriftRun,
		app.WorkflowTypeDeployComponents,
		app.WorkflowTypeTeardownComponent,
		app.WorkflowTypeTeardownComponents,
		app.WorkflowTypeComponentEnabled,
		app.WorkflowTypeComponentDisabled,
		app.WorkflowTypeRecoverHelmRelease:
		return InstallDeploymentTypeComponentDeploy
	case app.WorkflowTypeReprovisionSandbox, app.WorkflowTypeDriftRunReprovisionSandbox:
		return InstallDeploymentTypeSandboxReprovision
	case app.WorkflowTypeReprovision:
		return InstallDeploymentTypeReprovision
	case app.WorkflowTypeReprovisionStack:
		return InstallDeploymentTypeStackUpdate
	case app.WorkflowTypeAppBranchConfigUpdate,
		app.WorkflowTypeAppBranchesRun,
		app.WorkflowTypeAppBranchesConfigRepoUpdate,
		app.WorkflowTypeAppBranchesComponentRepoUpdate,
		app.WorkflowTypeAppInstallSync:
		return InstallDeploymentTypeAppBranchUpdate
	case app.WorkflowTypeInputUpdate, app.WorkflowTypeSyncSecrets:
		return InstallDeploymentTypeInstallConfigUpdate
	default:
		return ""
	}
}

// InstallDeploymentWorkflowRef is a lightweight reference to the backing workflow.
type InstallDeploymentWorkflowRef struct {
	ID   string           `json:"id"`
	Type app.WorkflowType `json:"type"`
	Name string           `json:"name"`
}

// InstallDeploymentAppBranchRef captures the app branch that originated this change, when applicable.
type InstallDeploymentAppBranchRef struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	RunID     string `json:"run_id,omitempty"`
	GitRef    string `json:"git_ref,omitempty"`
	CommitSHA string `json:"sha,omitempty"`
}

// InstallDeploymentComponentRef describes the primary component affected by a single-component operation.
type InstallDeploymentConfigChange struct {
	Path          string `json:"path"`
	Operation     string `json:"operation"`
	PreviousValue string `json:"previous_value,omitempty"`
	NextValue     string `json:"next_value,omitempty"`
	IsRedacted    bool   `json:"is_redacted,omitempty"`
}

type InstallDeploymentChangeGroup struct {
	ID           string                          `json:"id"`
	Scope        string                          `json:"scope"`
	Label        string                          `json:"label"`
	ResourceName string                          `json:"resource_name,omitempty"`
	Summary      string                          `json:"summary"`
	Changes      []InstallDeploymentConfigChange `json:"changes"`
	FileDiff     string                          `json:"file_diff,omitempty"`
	DiffLanguage string                          `json:"diff_language,omitempty"`
}

type InstallDeploymentAffectedResources struct {
	Stack      bool     `json:"stack,omitempty"`
	Sandbox    bool     `json:"sandbox,omitempty"`
	Components []string `json:"components"`
	Images     []string `json:"images"`
}

type InstallDeploymentImage struct {
	Repository  string `json:"repository"`
	PreviousTag string `json:"previous_tag,omitempty"`
	NextTag     string `json:"next_tag"`
}

// InstallDeployment is a normalized record representing a single change event on an install.
type InstallDeployment struct {
	ID        string                `json:"id"`
	Type      InstallDeploymentType `json:"type"`
	Status    string                `json:"status"`
	CreatedAt time.Time             `json:"created_at"`
	Title     string                `json:"title"`
	Summary   string                `json:"summary"`

	Workflow      *InstallDeploymentWorkflowRef  `json:"workflow,omitempty"`
	AppBranch     *InstallDeploymentAppBranchRef `json:"app_branch,omitempty"`
	ComponentName string                         `json:"component_name,omitempty"`
	Image         *InstallDeploymentImage        `json:"image,omitempty"`

	AffectedResources InstallDeploymentAffectedResources `json:"affected_resources"`
	ChangeGroups      []InstallDeploymentChangeGroup     `json:"change_groups"`
}

type InstallDeploymentPolicySummary struct {
	DenyCount        int    `json:"deny_count"`
	WarnCount        int    `json:"warn_count"`
	FirstWarnMessage string `json:"first_warn_message,omitempty"`
}

type InstallDeploymentStep struct {
	ID                 string                          `json:"id"`
	Name               string                          `json:"name"`
	Status             string                          `json:"status"`
	Idx                int                             `json:"idx"`
	GroupIdx           int                             `json:"group_idx"`
	GroupRetryIdx      int                             `json:"group_retry_idx"`
	Retried            bool                            `json:"retried,omitempty"`
	ExecutionType      string                          `json:"execution_type"`
	StepTargetType     string                          `json:"step_target_type,omitempty"`
	ComponentName      string                          `json:"component_name,omitempty"`
	ApprovalResponseID string                          `json:"approval_response_id,omitempty"`
	Policy             *InstallDeploymentPolicySummary `json:"policy,omitempty"`
}

type InstallDeploymentSummary struct {
	ID        string                  `json:"id"`
	Type      InstallDeploymentType   `json:"type"`
	Title     string                  `json:"title"`
	CreatedAt time.Time               `json:"created_at"`
	Status    string                  `json:"status"`
	Activity  string                  `json:"activity"`
	Finished  bool                    `json:"finished"`
	Steps     []InstallDeploymentStep `json:"steps"`
}

type GetInstallDeploymentSummariesResponse struct {
	Deployments []InstallDeploymentSummary `json:"deployments"`
	Page        int                        `json:"page"`
	Offset      int                        `json:"offset"`
	Limit       int                        `json:"limit"`
	HasMore     bool                       `json:"has_more"`
	NextCursor  string                     `json:"next_cursor,omitempty"`
	Total       *int64                     `json:"total,omitempty" extensions:"x-nullable"`
}

const (
	installDeploymentsStateActive   = "active"
	installDeploymentsStateFinished = "finished"
	installDeploymentsSortAttention = "attention"

	installDeploymentsStatusExpr = "COALESCE(status->>'status', '')"
	installDeploymentsRankExpr   = "CASE " + installDeploymentsStatusExpr + " WHEN 'approval-awaiting' THEN 0 WHEN 'failed-pending-retry' THEN 1 ELSE 2 END"
)

var installDeploymentActiveStatuses = []string{
	"",
	string(app.StatusPending),
	string(app.StatusQueued),
	string(app.StatusInProgress),
	string(app.StatusRetrying),
	string(app.AwaitingApproval),
	string(app.WorkflowStepApprovalStatusApproved),
	string(app.StatusFailedPendingRetry),
}

func installDeploymentAttentionRank(status app.Status) int {
	switch status {
	case app.AwaitingApproval:
		return 0
	case app.StatusFailedPendingRetry:
		return 1
	default:
		return 2
	}
}

type installDeploymentsCursor struct {
	State     string    `json:"st,omitempty"`
	Sort      string    `json:"so,omitempty"`
	Rank      *int      `json:"r,omitempty"`
	CreatedAt time.Time `json:"t"`
	ID        string    `json:"i"`
}

func encodeInstallDeploymentsCursor(c installDeploymentsCursor) (string, error) {
	byts, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(byts), nil
}

func decodeInstallDeploymentsCursor(raw, state, sortBy string) (*installDeploymentsCursor, error) {
	byts, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return nil, errors.New("cursor is not valid")
	}
	var c installDeploymentsCursor
	if err := json.Unmarshal(byts, &c); err != nil {
		return nil, errors.New("cursor is not valid")
	}
	if c.ID == "" || c.CreatedAt.IsZero() {
		return nil, errors.New("cursor is not valid")
	}
	if c.State != state || c.Sort != sortBy {
		return nil, errors.New("cursor does not match the requested state and sort")
	}
	if sortBy == installDeploymentsSortAttention {
		if c.Rank == nil || *c.Rank < 0 || *c.Rank > 2 {
			return nil, errors.New("cursor is not valid")
		}
	} else if c.Rank != nil {
		return nil, errors.New("cursor is not valid")
	}
	return &c, nil
}

type installDeploymentsQuery struct {
	page, offset, limit         int
	filterTypes, filterStatuses []string
	resource, search            string
	createdAtGte, createdAtLte  *time.Time
	state, sort                 string
	cursor                      *installDeploymentsCursor
}

// @ID                    GetInstallDeploymentSummaries
// @Summary               get lightweight deployment summaries for an install
// @Description.markdown  get_install_deployment_summaries.md
// @Param                 install_id      path   string  true   "install ID"
// @Param                 page            query  int     false  "page number"                                         Default(0)
// @Param                 offset          query  int     false  "offset of results to return"                         Default(0)
// @Param                 limit           query  int     false  "page size"                                           Default(20)
// @Param                 cursor          query  string  false  "opaque cursor from a previous next_cursor; replaces page and offset"
// @Param                 state           query  string  false  "filter by lifecycle state"                           Enums(active, finished)
// @Param                 sort            query  string  false  "sort order; attention puts approvals and failed retries first"  Enums(attention)
// @Param                 type            query  string  false  "filter by deployment type (comma-separated)"
// @Param                 status          query  string  false  "filter by workflow status (comma-separated)"
// @Param                 resource        query  string  false  "filter by affected stack, sandbox, or component name"
// @Param                 search          query  string  false  "case-insensitive substring match on id or title"
// @Param                 created_at_gte  query  string  false  "include deployments created at or after this RFC3339 timestamp"
// @Param                 created_at_lte  query  string  false  "include deployments created at or before this RFC3339 timestamp"
// @Tags                  installs
// @Accept                json
// @Produce               json
// @Security              APIKey
// @Security              OrgID
// @Failure               400  {object}  stderr.ErrResponse
// @Failure               401  {object}  stderr.ErrResponse
// @Failure               403  {object}  stderr.ErrResponse
// @Failure               404  {object}  stderr.ErrResponse
// @Failure               500  {object}  stderr.ErrResponse
// @Success               200  {object}  GetInstallDeploymentSummariesResponse
// @Router                /v1/installs/{install_id}/deployment-summaries [GET]
func (s *service) GetInstallDeploymentSummaries(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	installID := ctx.Param("install_id")

	q := installDeploymentsQuery{
		limit:          queryInt(ctx, "limit", 20, 1, 100),
		filterTypes:    parseCommaSeparated(ctx.Query("type")),
		filterStatuses: parseCommaSeparated(ctx.Query("status")),
		resource:       ctx.Query("resource"),
		search:         ctx.Query("search"),
		state:          ctx.Query("state"),
		sort:           ctx.Query("sort"),
	}
	q.page = queryInt(ctx, "page", 0, 0, 10_000)
	q.offset = queryInt(ctx, "offset", 0, 0, 1_000_000)
	if ctx.Query("page") != "" {
		q.offset = q.page * q.limit
	} else {
		q.page = q.offset / q.limit
	}

	switch q.state {
	case "", installDeploymentsStateActive, installDeploymentsStateFinished:
	default:
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("invalid state %q", q.state),
			Description: "state must be one of: active, finished",
		})
		return
	}

	switch q.sort {
	case "", installDeploymentsSortAttention:
	default:
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("invalid sort %q", q.sort),
			Description: "sort must be: attention",
		})
		return
	}

	if raw := ctx.Query("cursor"); raw != "" {
		if q.offset > 0 {
			ctx.Error(stderr.ErrUser{
				Err:         errors.New("cursor cannot be combined with page or offset"),
				Description: "use either cursor or page/offset",
			})
			return
		}
		cursor, err := decodeInstallDeploymentsCursor(raw, q.state, q.sort)
		if err != nil {
			ctx.Error(stderr.ErrUser{
				Err:         err,
				Description: "invalid cursor",
			})
			return
		}
		q.cursor = cursor
	}

	if raw := ctx.Query("created_at_gte"); raw != "" {
		t, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			ctx.Error(errors.Wrap(err, "invalid created_at_gte, must be RFC3339"))
			return
		}
		q.createdAtGte = &t
	}

	if raw := ctx.Query("created_at_lte"); raw != "" {
		t, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			ctx.Error(errors.Wrap(err, "invalid created_at_lte, must be RFC3339"))
			return
		}
		q.createdAtLte = &t
	}

	resp, err := s.getInstallDeploymentSummaries(ctx, org.ID, installID, q)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install deployments: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

func (s *service) getInstallDeploymentSummaries(
	ctx *gin.Context,
	orgID, installID string,
	q installDeploymentsQuery,
) (*GetInstallDeploymentSummariesResponse, error) {
	resp := &GetInstallDeploymentSummariesResponse{
		Deployments: []InstallDeploymentSummary{},
		Page:        q.page,
		Offset:      q.offset,
		Limit:       q.limit,
	}
	if q.state == installDeploymentsStateActive {
		var zero int64
		resp.Total = &zero
	}

	workflowTypes := deploymentWorkflowTypes(q.filterTypes)
	if len(workflowTypes) == 0 {
		return resp, nil
	}

	base := s.db.WithContext(ctx).
		Model(&app.Workflow{}).
		Where("owner_id = ?", installID).
		Where("org_id = ?", orgID).
		Where("plan_only = ?", false).
		Where("type IN ?", workflowTypes)

	switch q.state {
	case installDeploymentsStateActive:
		base = base.Where(installDeploymentsStatusExpr+" IN ?", installDeploymentActiveStatuses)
	case installDeploymentsStateFinished:
		base = base.Where(installDeploymentsStatusExpr+" NOT IN ?", installDeploymentActiveStatuses)
	}

	if len(q.filterStatuses) > 0 {
		base = base.Where("status->>'status' IN ?", q.filterStatuses)
	}

	const stepExists = `
		EXISTS (
			SELECT 1 FROM install_workflow_steps s
			WHERE s.install_workflow_id = install_workflows.id
			  AND s.deleted_at = 0
			  AND s.execution_type IS DISTINCT FROM 'hidden'
			  AND `
	switch q.resource {
	case "":
	case "stack":
		base = base.Where(stepExists+`(s.step_target_type = 'install_stack_versions' OR s.name ~* ?))`, `\m(install stack|stack policy)\M`)
	case "sandbox":
		base = base.Where(stepExists+`(s.step_target_type = 'install_sandbox_runs' OR s.name ~* ?))`, `\msandbox\M`)
	default:
		base = base.Where(`(`+stepExists+`s.metadata -> 'component_name' = ?)
			OR EXISTS (
				SELECT 1
				FROM install_deploys d
				JOIN install_components ic ON ic.id = d.install_component_id
				JOIN components c ON c.id = ic.component_id
				WHERE d.install_workflow_id = install_workflows.id
				  AND d.deleted_at = 0
				  AND ic.deleted_at = 0
				  AND c.deleted_at = 0
				  AND (
				c.name = ?
				OR EXISTS (
					SELECT 1 FROM component_builds cb
					WHERE cb.id = d.component_build_id
					  AND cb.deleted_at = 0
					  AND (cb.source_image = ? OR cb.source_ref = ?)
				)
			  )
			))`, q.resource, q.resource, q.resource, q.resource)
	}

	for _, token := range strings.Fields(q.search) {
		like := "%" + token + "%"
		base = base.Where("name ILIKE ? OR id ILIKE ?", like, like)
	}

	if q.createdAtGte != nil {
		base = base.Where("created_at >= ?", q.createdAtGte)
	}
	if q.createdAtLte != nil {
		base = base.Where("created_at <= ?", q.createdAtLte)
	}

	if resp.Total != nil {
		if err := base.Session(&gorm.Session{}).Count(resp.Total).Error; err != nil {
			return nil, fmt.Errorf("unable to count workflows: %w", err)
		}
	}

	query := base.Session(&gorm.Session{}).
		Select("id", "name", "type", "status", "created_at", "finished_at")
	if q.sort == installDeploymentsSortAttention {
		query = query.Order(installDeploymentsRankExpr + " ASC")
	}
	query = query.
		Order("created_at DESC").
		Order("id DESC").
		Limit(q.limit + 1)

	if q.cursor != nil {
		if q.sort == installDeploymentsSortAttention {
			query = query.Where(
				"("+installDeploymentsRankExpr+" > ? OR ("+installDeploymentsRankExpr+" = ? AND (created_at, id) < (?, ?)))",
				*q.cursor.Rank, *q.cursor.Rank, q.cursor.CreatedAt, q.cursor.ID,
			)
		} else {
			query = query.Where("(created_at, id) < (?, ?)", q.cursor.CreatedAt, q.cursor.ID)
		}
	} else {
		query = query.Offset(q.offset)
	}

	var workflows []app.Workflow
	if err := query.Find(&workflows).Error; err != nil {
		return nil, fmt.Errorf("unable to query workflows: %w", err)
	}
	resp.HasMore = len(workflows) > q.limit
	if resp.HasMore {
		workflows = workflows[:q.limit]
		last := &workflows[len(workflows)-1]
		next := installDeploymentsCursor{
			State:     q.state,
			Sort:      q.sort,
			CreatedAt: last.CreatedAt,
			ID:        last.ID,
		}
		if q.sort == installDeploymentsSortAttention {
			rank := installDeploymentAttentionRank(last.Status.Status)
			next.Rank = &rank
		}
		nextCursor, err := encodeInstallDeploymentsCursor(next)
		if err != nil {
			return nil, fmt.Errorf("unable to encode cursor: %w", err)
		}
		resp.NextCursor = nextCursor
	}

	workflowIDs := make([]string, 0, len(workflows))
	for i := range workflows {
		workflowIDs = append(workflowIDs, workflows[i].ID)
	}
	stepsByWorkflowID, err := s.installDeploymentSteps(ctx, workflowIDs)
	if err != nil {
		return nil, err
	}

	for i := range workflows {
		resp.Deployments = append(resp.Deployments, buildInstallDeploymentSummary(&workflows[i], stepsByWorkflowID[workflows[i].ID]))
	}

	return resp, nil
}

func buildInstallDeploymentSummary(wf *app.Workflow, steps []InstallDeploymentStep) InstallDeploymentSummary {
	status := string(wf.Status.Status)
	if status == "" {
		status = "pending"
	}
	title := wf.Name
	if title == "" {
		title = wf.Type.Name()
	}
	activity := wf.Status.StatusHumanDescription
	if activity == "" {
		activity = wf.Type.Description()
	}
	if steps == nil {
		steps = []InstallDeploymentStep{}
	}
	return InstallDeploymentSummary{
		ID:        wf.ID,
		Type:      workflowTypeToDeploymentType(wf.Type),
		Title:     title,
		CreatedAt: wf.CreatedAt,
		Status:    status,
		Activity:  activity,
		Finished:  wf.Finished,
		Steps:     steps,
	}
}

func (s *service) installDeploymentSteps(ctx *gin.Context, workflowIDs []string) (map[string][]InstallDeploymentStep, error) {
	stepsByWorkflowID := make(map[string][]InstallDeploymentStep, len(workflowIDs))
	if len(workflowIDs) == 0 {
		return stepsByWorkflowID, nil
	}

	var steps []app.WorkflowStep
	if err := s.db.WithContext(ctx).
		Select("id", "install_workflow_id", "name", "status", "idx", "group_idx", "group_retry_idx", "retried", "execution_type", "step_target_type", "metadata").
		Where("install_workflow_id IN ?", workflowIDs).
		Where("execution_type IS DISTINCT FROM ?", app.WorkflowStepExecutionTypeHidden).
		Order("group_idx, group_retry_idx, idx, created_at asc").
		Find(&steps).Error; err != nil {
		return nil, fmt.Errorf("unable to query workflow steps: %w", err)
	}

	awaitingStepIDs := make([]string, 0)
	for i := range steps {
		if steps[i].Status.Status == app.AwaitingApproval {
			awaitingStepIDs = append(awaitingStepIDs, steps[i].ID)
		}
	}
	responseIDs := make(map[string]string, len(awaitingStepIDs))
	if len(awaitingStepIDs) > 0 {
		var responses []struct {
			StepID     string
			ResponseID string
		}
		if err := s.db.WithContext(ctx).
			Model(&app.WorkflowStepApproval{}).
			Select("install_workflow_step_approvals.install_workflow_step_id AS step_id, r.id AS response_id").
			Joins("JOIN install_workflow_step_approval_responses r ON r.install_workflow_step_approval_id = install_workflow_step_approvals.id AND r.deleted_at = 0").
			Where("install_workflow_step_approvals.install_workflow_step_id IN ?", awaitingStepIDs).
			Scan(&responses).Error; err != nil {
			return nil, fmt.Errorf("unable to query workflow step approval responses: %w", err)
		}
		for _, response := range responses {
			responseIDs[response.StepID] = response.ResponseID
		}
	}

	for i := range steps {
		step := &steps[i]
		componentName := ""
		if name, ok := step.Metadata["component_name"]; ok && name != nil {
			componentName = *name
		}
		stepsByWorkflowID[step.InstallWorkflowID] = append(stepsByWorkflowID[step.InstallWorkflowID], InstallDeploymentStep{
			ID:                 step.ID,
			Name:               step.Name,
			Status:             string(step.Status.Status),
			Idx:                step.Idx,
			GroupIdx:           step.GroupIdx,
			GroupRetryIdx:      step.GroupRetryIdx,
			Retried:            step.Retried,
			ExecutionType:      string(step.ExecutionType),
			StepTargetType:     step.StepTargetType,
			ComponentName:      componentName,
			ApprovalResponseID: responseIDs[step.ID],
			Policy:             installDeploymentPolicySummary(step.Status.Metadata),
		})
	}
	return stepsByWorkflowID, nil
}

func installDeploymentPolicySummary(metadata map[string]any) *InstallDeploymentPolicySummary {
	denials, _ := metadata["deny_violations"].([]any)
	warnings, _ := metadata["warn_violations"].([]any)
	if len(denials)+len(warnings) == 0 {
		return nil
	}
	summary := &InstallDeploymentPolicySummary{
		DenyCount: len(denials),
		WarnCount: len(warnings),
	}
	if len(warnings) > 0 {
		first, _ := warnings[0].(map[string]any)
		summary.FirstWarnMessage, _ = first["message"].(string)
	}
	return summary
}

func (s *service) buildInstallDeployments(ctx *gin.Context, orgID, installID string, workflows []app.Workflow) ([]InstallDeployment, error) {
	workflowIDs := make([]string, 0, len(workflows))
	for i := range workflows {
		workflowIDs = append(workflowIDs, workflows[i].ID)
	}
	versionsByWorkflowID := make(map[string]*app.InstallAppConfigVersion)
	diffsByWorkflowID := make(map[string]*app.InstallConfigDiff)
	if len(workflowIDs) > 0 {
		var versions []app.InstallAppConfigVersion
		if err := s.db.WithContext(ctx).
			Preload("AppBranchRun").
			Preload("AppBranchRun.AppBranch").
			Preload("AppBranchRun.VCSConnectionCommit").
			Where(app.InstallAppConfigVersion{OrgID: orgID, InstallID: installID}).
			Where("workflow_id IN ?", workflowIDs).
			Order("created_at ASC").
			Order("id ASC").
			Find(&versions).Error; err != nil {
			return nil, fmt.Errorf("unable to query install app config versions: %w", err)
		}
		for i := range versions {
			if versions[i].WorkflowID != nil {
				versionsByWorkflowID[*versions[i].WorkflowID] = &versions[i]
			}
		}
		var mu sync.Mutex
		g, gctx := errgroup.WithContext(blobstore.WithBlobService(ctx.Request.Context(), s.blobSvc))
		g.SetLimit(8)
		for workflowID, version := range versionsByWorkflowID {
			g.Go(func() error {
				diff, err := loadInstallConfigDiff(gctx, version)
				if err != nil {
					return fmt.Errorf("unable to load install app config diff: %w", err)
				}
				mu.Lock()
				diffsByWorkflowID[workflowID] = diff
				mu.Unlock()
				return nil
			})
		}
		if err := g.Wait(); err != nil {
			return nil, err
		}
	}

	buildIDs := collectDiffBuildIDs(diffsByWorkflowID)
	buildsByID, err := s.componentBuildsByID(ctx, buildIDs)
	if err != nil {
		return nil, err
	}
	fallbackBranch, err := s.installDeploymentBranch(ctx, orgID, installID)
	if err != nil {
		return nil, err
	}
	pinnedBranch, err := s.pinnedAppConfigBranch(ctx, orgID, installID)
	if err != nil {
		return nil, err
	}

	deployments := make([]InstallDeployment, 0, len(workflows))
	for i := range workflows {
		d := buildInstallDeployment(&workflows[i], versionsByWorkflowID[workflows[i].ID], diffsByWorkflowID[workflows[i].ID], buildsByID, fallbackBranch)
		applyPinnedConfigBranch(&d, pinnedBranch)
		if d.Type == "" {
			continue
		}
		deployments = append(deployments, d)
	}
	return deployments, nil
}

func deploymentWorkflowTypes(filterTypes []string) []app.WorkflowType {
	filters := make(map[string]struct{}, len(filterTypes))
	for _, filterType := range filterTypes {
		filters[filterType] = struct{}{}
	}

	types := make([]app.WorkflowType, 0)
	for _, workflowType := range app.AllWorkflowTypes() {
		deploymentType := workflowTypeToDeploymentType(workflowType)
		if deploymentType == "" {
			continue
		}
		if len(filters) > 0 {
			if _, ok := filters[string(deploymentType)]; !ok {
				continue
			}
		}
		types = append(types, workflowType)
	}
	return types
}

func buildInstallDeployment(wf *app.Workflow, version *app.InstallAppConfigVersion, configDiff *app.InstallConfigDiff, builds map[string]app.ComponentBuild, fallbackBranch *InstallDeploymentAppBranchRef) InstallDeployment {
	status := wf.Status.Status
	if status == "" {
		status = "pending"
	}

	title := wf.Name
	if title == "" {
		title = wf.Type.Name()
	}

	d := InstallDeployment{
		ID:        wf.ID,
		Type:      workflowTypeToDeploymentType(wf.Type),
		Status:    string(status),
		CreatedAt: wf.CreatedAt,
		Title:     title,
		Summary:   wf.Type.Description(),
		Workflow: &InstallDeploymentWorkflowRef{
			ID:   wf.ID,
			Type: wf.Type,
			Name: wf.Name,
		},
		AffectedResources: InstallDeploymentAffectedResources{
			Components: []string{},
			Images:     []string{},
		},
		ChangeGroups: []InstallDeploymentChangeGroup{},
	}

	componentNames := collectComponentNames(wf.InstallDeploys)
	if len(componentNames) > 0 {
		d.AffectedResources.Components = componentNames
		d.ChangeGroups = append(d.ChangeGroups, InstallDeploymentChangeGroup{
			ID:      wf.ID + "-components",
			Scope:   "component",
			Label:   "Components",
			Summary: strings.Join(componentNames, ", "),
			Changes: []InstallDeploymentConfigChange{},
		})
	}

	if len(wf.InstallDeploys) == 1 {
		dep := &wf.InstallDeploys[0]
		d.ComponentName = dep.InstallComponent.Component.Name
	}

	if version != nil && version.AppBranchRunID != nil {
		run := &version.AppBranchRun
		branchRef := &InstallDeploymentAppBranchRef{
			ID:    run.AppBranchID,
			Name:  run.AppBranch.Name,
			RunID: run.ID,
		}
		meta := run.RunMetadata()
		branchRef.GitRef = meta.GitRef
		branchRef.CommitSHA = meta.HeadSHA
		if branchRef.CommitSHA == "" && run.VCSConnectionCommit != nil {
			branchRef.CommitSHA = run.VCSConnectionCommit.SHA
		}
		d.AppBranch = branchRef
	}
	if d.AppBranch == nil && fallbackBranch != nil {
		copied := *fallbackBranch
		d.AppBranch = &copied
	}

	categoryGroups := changeGroupsForWorkflowType(wf.Type)
	for _, cg := range categoryGroups {
		if cg.Scope == "component" && len(componentNames) > 0 {
			continue
		}
		d.ChangeGroups = append(d.ChangeGroups, cg)
		switch cg.Scope {
		case "stack":
			d.AffectedResources.Stack = true
		case "sandbox":
			d.AffectedResources.Sandbox = true
		}
	}
	applyInstallConfigDiff(&d, configDiff, builds)
	applyDeploymentImage(&d, wf)
	sort.Strings(d.AffectedResources.Components)
	sort.Strings(d.AffectedResources.Images)

	return d
}

func collectComponentNames(deploys []app.InstallDeploy) []string {
	seen := make(map[string]struct{})
	names := make([]string, 0)
	for i := range deploys {
		name := deploys[i].InstallComponent.Component.Name
		if name == "" {
			name = deploys[i].InstallComponent.ComponentID
		}
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		names = append(names, name)
	}
	return names
}

func changeGroupsForWorkflowType(t app.WorkflowType) []InstallDeploymentChangeGroup {
	switch t {
	case app.WorkflowTypeInputUpdate:
		return []InstallDeploymentChangeGroup{{ID: "inputs", Scope: "install_config", Label: "Inputs", Summary: "Install inputs changed", Changes: []InstallDeploymentConfigChange{}}}
	case app.WorkflowTypeReprovisionStack, app.WorkflowTypeReprovision:
		return []InstallDeploymentChangeGroup{{ID: "stack", Scope: "stack", Label: "Stack", Summary: "Install stack changed", Changes: []InstallDeploymentConfigChange{}}}
	case app.WorkflowTypeReprovisionSandbox, app.WorkflowTypeDriftRunReprovisionSandbox:
		return []InstallDeploymentChangeGroup{{ID: "sandbox", Scope: "sandbox", Label: "Sandbox", Summary: "Install sandbox changed", Changes: []InstallDeploymentConfigChange{}}}
	case app.WorkflowTypeProvision:
		return []InstallDeploymentChangeGroup{
			{ID: "stack", Scope: "stack", Label: "Stack", Summary: "Install stack provisioned", Changes: []InstallDeploymentConfigChange{}},
			{ID: "sandbox", Scope: "sandbox", Label: "Sandbox", Summary: "Install sandbox provisioned", Changes: []InstallDeploymentConfigChange{}},
			{ID: "components", Scope: "component", Label: "Components", Summary: "Install components deployed", Changes: []InstallDeploymentConfigChange{}},
		}
	default:
		return nil
	}
}

func applyInstallConfigDiff(deployment *InstallDeployment, configDiff *app.InstallConfigDiff, builds map[string]app.ComponentBuild) {
	if configDiff == nil {
		return
	}

	componentEntries := []struct {
		operation string
		entries   []app.ComponentDiffEntry
	}{
		{operation: "add", entries: configDiff.Added},
		{operation: "change", entries: configDiff.Changed},
		{operation: "remove", entries: configDiff.Removed},
	}
	for _, set := range componentEntries {
		for _, entry := range set.entries {
			name := entry.ComponentName
			if name == "" {
				name = entry.ComponentID
			}
			deployment.AffectedResources.Components = appendUniqueString(deployment.AffectedResources.Components, name)
			group := InstallDeploymentChangeGroup{
				ID:           deployment.ID + "-component-" + entry.ComponentID,
				Scope:        "component",
				Label:        "Component",
				ResourceName: name,
				Summary:      name,
				Changes: []InstallDeploymentConfigChange{{
					Path:          "build",
					Operation:     set.operation,
					PreviousValue: entry.OldBuildID,
					NextValue:     entry.NewBuildID,
				}},
			}
			if image := imageFromBuilds(builds[entry.OldBuildID], builds[entry.NewBuildID]); image != nil && app.ComponentType(entry.ComponentType).IsImage() {
				deployment.Image = image
				deployment.AffectedResources.Images = appendUniqueString(deployment.AffectedResources.Images, image.Repository)
				if deployment.Type == InstallDeploymentTypeComponentDeploy && image.PreviousTag != "" && image.PreviousTag != image.NextTag {
					deployment.Type = InstallDeploymentTypeImageUpdate
				}
			}
			deployment.ChangeGroups = append(deployment.ChangeGroups, group)
		}
	}
	if diffText := renderInstallConfigDiff(configDiff); diffText != "" {
		deployment.ChangeGroups = append(deployment.ChangeGroups, InstallDeploymentChangeGroup{
			ID:           deployment.ID + "-config-diff",
			Scope:        "app_branch",
			Label:        "Config diff",
			Summary:      "App config changes",
			Changes:      []InstallDeploymentConfigChange{},
			FileDiff:     diffText,
			DiffLanguage: "diff",
		})
	}

	if configDiff.StackChanged {
		deployment.AffectedResources.Stack = true
		deployment.ChangeGroups = append(deployment.ChangeGroups, InstallDeploymentChangeGroup{
			ID:      deployment.ID + "-stack",
			Scope:   "stack",
			Label:   "Stack",
			Summary: "Stack configuration changed",
			Changes: []InstallDeploymentConfigChange{{
				Path:          "config",
				Operation:     "change",
				PreviousValue: configDiff.StackOldID,
				NextValue:     configDiff.StackNewID,
			}},
		})
	}

	if configDiff.SandboxChanged || configDiff.SandboxBuildChanged {
		deployment.AffectedResources.Sandbox = true
		changes := make([]InstallDeploymentConfigChange, 0, 2)
		if configDiff.SandboxChanged {
			changes = append(changes, InstallDeploymentConfigChange{
				Path:          "config",
				Operation:     "change",
				PreviousValue: configDiff.SandboxOldID,
				NextValue:     configDiff.SandboxNewID,
			})
		}
		if configDiff.SandboxBuildChanged {
			changes = append(changes, InstallDeploymentConfigChange{
				Path:          "build",
				Operation:     "change",
				PreviousValue: configDiff.SandboxBuildOldID,
				NextValue:     configDiff.SandboxBuildNewID,
			})
		}
		deployment.ChangeGroups = append(deployment.ChangeGroups, InstallDeploymentChangeGroup{
			ID:      deployment.ID + "-sandbox",
			Scope:   "sandbox",
			Label:   "Sandbox",
			Summary: "Sandbox configuration changed",
			Changes: changes,
		})
	}
}

func collectDiffBuildIDs(diffs map[string]*app.InstallConfigDiff) []string {
	seen := map[string]struct{}{}
	ids := make([]string, 0)
	add := func(id string) {
		if id == "" {
			return
		}
		if _, ok := seen[id]; ok {
			return
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	for _, diff := range diffs {
		if diff == nil {
			continue
		}
		for _, entries := range [][]app.ComponentDiffEntry{diff.Added, diff.Changed, diff.Removed} {
			for _, entry := range entries {
				add(entry.OldBuildID)
				add(entry.NewBuildID)
			}
		}
	}
	return ids
}

func (s *service) componentBuildsByID(ctx *gin.Context, ids []string) (map[string]app.ComponentBuild, error) {
	builds := map[string]app.ComponentBuild{}
	if len(ids) == 0 {
		return builds, nil
	}
	var rows []app.ComponentBuild
	if err := s.db.WithContext(ctx).Where("id IN ?", ids).Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("unable to query component builds: %w", err)
	}
	for i := range rows {
		builds[rows[i].ID] = rows[i]
	}
	return builds, nil
}

func applyPinnedConfigBranch(d *InstallDeployment, pinned *InstallDeploymentAppBranchRef) {
	if pinned == nil || pinned.RunID == "" {
		return
	}
	if d.Type != InstallDeploymentTypeProvision && d.Type != InstallDeploymentTypeReprovision {
		return
	}
	if d.AppBranch != nil && d.AppBranch.RunID != "" {
		return
	}
	copied := *pinned
	d.AppBranch = &copied
}

func (s *service) pinnedAppConfigBranch(ctx *gin.Context, orgID, installID string) (*InstallDeploymentAppBranchRef, error) {
	var install app.Install
	err := s.db.WithContext(ctx).
		Where(app.Install{ID: installID, OrgID: orgID}).
		First(&install).Error
	if isNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("unable to get install app config: %w", err)
	}
	if install.AppConfigID == "" {
		return nil, nil
	}

	var run app.AppBranchRun
	err = s.db.WithContext(ctx).
		Preload("AppBranch").
		Preload("VCSConnectionCommit").
		Where(app.AppBranchRun{OrgID: orgID, AppConfigID: install.AppConfigID}).
		Order("created_at DESC").
		First(&run).Error
	if isNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("unable to get app config branch run: %w", err)
	}

	ref := &InstallDeploymentAppBranchRef{
		ID:    run.AppBranchID,
		Name:  run.AppBranch.Name,
		RunID: run.ID,
	}
	meta := run.RunMetadata()
	ref.GitRef = meta.GitRef
	ref.CommitSHA = meta.HeadSHA
	if ref.CommitSHA == "" && run.VCSConnectionCommit != nil {
		ref.CommitSHA = run.VCSConnectionCommit.SHA
	}
	return ref, nil
}

func (s *service) installDeploymentBranch(ctx *gin.Context, orgID, installID string) (*InstallDeploymentAppBranchRef, error) {
	var connection app.InstallAppBranchConnection
	err := s.db.WithContext(ctx).
		Preload("AppBranch").
		Where(app.InstallAppBranchConnection{InstallID: installID, OrgID: orgID}).
		First(&connection).Error
	if isNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("unable to get install app branch: %w", err)
	}
	return &InstallDeploymentAppBranchRef{
		ID:   connection.AppBranchID,
		Name: connection.AppBranch.Name,
	}, nil
}

func imageFromBuilds(previous, next app.ComponentBuild) *InstallDeploymentImage {
	if previous.ID == "" && next.ID == "" {
		return nil
	}
	repo := next.SourceImage
	if repo == "" {
		repo = previous.SourceImage
	}
	nextTag := next.ResolvedTag
	if nextTag == "" {
		nextTag = tagFromImageRef(next.SourceRef)
	}
	prevTag := previous.ResolvedTag
	if prevTag == "" {
		prevTag = tagFromImageRef(previous.SourceRef)
	}
	if repo == "" && nextTag == "" {
		return nil
	}
	return &InstallDeploymentImage{
		Repository:  repo,
		PreviousTag: prevTag,
		NextTag:     nextTag,
	}
}

func tagFromImageRef(ref string) string {
	if i := strings.Index(ref, "@"); i >= 0 {
		ref = ref[:i]
	}
	if i := strings.LastIndex(ref, ":"); i >= 0 {
		return ref[i+1:]
	}
	return ""
}

func applyDeploymentImage(deployment *InstallDeployment, wf *app.Workflow) {
	if deployment.Image != nil {
		return
	}
	for i := range wf.InstallDeploys {
		dep := &wf.InstallDeploys[i]
		if !dep.InstallComponent.Component.Type.IsImage() {
			continue
		}
		image := imageFromBuilds(app.ComponentBuild{}, dep.ComponentBuild)
		if image == nil {
			continue
		}
		deployment.Image = image
		deployment.AffectedResources.Images = appendUniqueString(deployment.AffectedResources.Images, image.Repository)
		return
	}
}

func renderInstallConfigDiff(configDiff *app.InstallConfigDiff) string {
	if configDiff == nil {
		return ""
	}
	var b strings.Builder
	writeEntries := func(prefix string, entries []app.ComponentDiffEntry) {
		for _, entry := range entries {
			name := entry.ComponentName
			if name == "" {
				name = entry.ComponentID
			}
			if name == "" {
				continue
			}
			fmt.Fprintf(&b, "%s %s\n", prefix, name)
		}
	}
	writeEntries("+", configDiff.Added)
	writeEntries("~", configDiff.Changed)
	writeEntries("-", configDiff.Removed)
	if configDiff.StackChanged {
		b.WriteString("~ stack\n")
	}
	if configDiff.SandboxChanged || configDiff.SandboxBuildChanged {
		b.WriteString("~ sandbox\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

func containsString(values []string, value string) bool {
	for _, existing := range values {
		if existing == value {
			return true
		}
	}
	return false
}

func appendUniqueString(values []string, value string) []string {
	if value == "" {
		return values
	}
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}

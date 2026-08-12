package service

import (
	"github.com/gin-gonic/gin"
)

// The handlers below delegate to their bare-route counterparts. They exist so
// each ancestor-scoped path gets its own swagger operation, and so a generated
// client method, rather than being reachable only by hand-built requests.
// Ancestor ownership is enforced by each group's guard, not here.

// @ID						GetInstallRunnerJob
// @Summary				get runner job
// @Description			Ancestor-scoped alternative to the bare route: authorizes against the install that owns the resource, so an identity scoped to that install can reach it.
// @Param					install_id	path	string	true	"install ID"
// @Param					runner_job_id	path	string	true	"runner job ID"
// @Tags					runners,runners/runner
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	app.RunnerJob
// @Router					/v1/installs/{install_id}/runner-jobs/{runner_job_id} [get]
func (s *service) GetInstallRunnerJob(ctx *gin.Context) {
	s.GetRunnerJobPublic(ctx)
}

// @ID						GetInstallRunnerJobPlan
// @Summary				get runner job plan
// @Description			Ancestor-scoped alternative to the bare route: authorizes against the install that owns the resource, so an identity scoped to that install can reach it.
// @Param					install_id	path	string	true	"install ID"
// @Param					runner_job_id	path	string	true	"runner job ID"
// @Tags					runners,runners/runner
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	string
// @Router					/v1/installs/{install_id}/runner-jobs/{runner_job_id}/plan [get]
func (s *service) GetInstallRunnerJobPlan(ctx *gin.Context) {
	s.GetRunnerJobPlanPublic(ctx)
}

// @ID						GetInstallRunnerJobCompositePlan
// @Summary				get runner job composite plan
// @Description			Ancestor-scoped alternative to the bare route: authorizes against the install that owns the resource, so an identity scoped to that install can reach it.
// @Param					install_id	path	string	true	"install ID"
// @Param					runner_job_id	path	string	true	"runner job ID"
// @Tags					runners,runners/runner
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	plantypes.CompositePlan
// @Router					/v1/installs/{install_id}/runner-jobs/{runner_job_id}/composite-plan [get]
func (s *service) GetInstallRunnerJobCompositePlan(ctx *gin.Context) {
	s.GetRunnerJobCompositePlan(ctx)
}

// @ID						CancelInstallRunnerJob
// @Summary				cancel runner job
// @Description			Ancestor-scoped alternative to the bare route: authorizes against the install that owns the resource, so an identity scoped to that install can reach it.
// @Param					install_id	path	string	true	"install ID"
// @Param					req				body	CancelRunnerJobRequest	true	"Input"
// @Param					runner_job_id	path	string					true	"runner job ID"
// @Tags					runners
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				202	{object}	app.RunnerJob
// @Router					/v1/installs/{install_id}/runner-jobs/{runner_job_id}/cancel [POST]
func (s *service) CancelInstallRunnerJob(ctx *gin.Context) {
	s.CancelRunnerJob(ctx)
}

// @ID						ReadInstallRunnerJobLogs
// @Summary				read a log stream's logs
// @Description			Ancestor-scoped alternative to the bare route: authorizes against the install that owns the resource, so an identity scoped to that install can reach it.
// @Param					install_id	path	string	true	"install ID"
// @Param					runner_job_id	path	string	true	"runner job ID"
// @Param					X-Nuon-API-Offset	header	string	false	"log stream offset"
// @Param					order				query	string	false	"sort direction"	default(asc)
// @Param					service_name		query	[]string	false	"filter by service_name (repeatable)"
// @Param					scope_name			query	[]string	false	"filter by scope_name (repeatable)"
// @Param					severity_text		query	[]string	false	"filter by severity_text (repeatable)"
// @Param					tool				query	string		false	"filter by log_attributes['nuon.tool']"
// @Param					helm_release_name	query	string		false	"filter by log_attributes['helm.release_name']"
// @Param					helm_operation		query	string		false	"filter by log_attributes['helm.operation']"
// @Param					tf_workspace_id		query	string		false	"filter by log_attributes['tf.workspace_id']"
// @Param					tf_operation		query	string		false	"filter by log_attributes['tf.operation']"
// @Param					k8s_kind			query	string		false	"filter by log_attributes['k8s.kind']"
// @Param					k8s_namespace		query	string		false	"filter by log_attributes['k8s.namespace']"
// @Param					k8s_name			query	string		false	"filter by log_attributes['k8s.name']"
// @Param					trace_id			query	string		false	"filter by exact trace_id (dedicated CH column)"
// @Param					span_id				query	string		false	"filter by exact span_id (dedicated CH column)"
// @Param					q					query	string		false	"case-insensitive substring filter on log body"
// @Tags					runners
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	[]app.OtelLogRecord
// @Router					/v1/installs/{install_id}/runner-jobs/{runner_job_id}/logs [GET]
func (s *service) ReadInstallRunnerJobLogs(ctx *gin.Context) {
	s.LogStreamReadLogs(ctx)
}

// @ID						TailInstallRunnerJobLogs
// @Summary				long-poll tail a log stream
// @Description			Ancestor-scoped alternative to the bare route: authorizes against the install that owns the resource, so an identity scoped to that install can reach it.
// @Param					install_id	path	string	true	"install ID"
// @Param					runner_job_id	path	string	true	"runner job ID"
// @Param					since			query	string	false	"composite cursor in the form `<unix_nano>:<id>`; empty starts from the oldest row"
// @Param					wait			query	string	false	"max wait for new rows (Go duration, capped server-side at 30s)"
// @Tags					runners
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	LogStreamTailLogsResponse
// @Router					/v1/installs/{install_id}/runner-jobs/{runner_job_id}/logs/tail [GET]
func (s *service) TailInstallRunnerJobLogs(ctx *gin.Context) {
	s.LogStreamTailLogs(ctx)
}

// @ID						ReadInstallRunnerJobSpans
// @Summary				read a log stream's trace spans
// @Description			Ancestor-scoped alternative to the bare route: authorizes against the install that owns the resource, so an identity scoped to that install can reach it.
// @Param					install_id	path	string	true	"install ID"
// @Param					runner_job_id	path	string	true	"runner job ID"
// @Tags					runners
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	[]LogStreamSpan
// @Router					/v1/installs/{install_id}/runner-jobs/{runner_job_id}/spans [GET]
func (s *service) ReadInstallRunnerJobSpans(ctx *gin.Context) {
	s.LogStreamReadSpans(ctx)
}

// @ID						GetInstallTerraformWorkspace
// @Summary				get  terraform workspace
// @Description			Ancestor-scoped alternative to the bare route: authorizes against the install that owns the resource, so an identity scoped to that install can reach it.
// @Param					install_id	path	string	true	"install ID"
// @Param					workspace_id	path	string	true	"workspace ID"
// @Tags					runners,runners/runner
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{array}	app.TerraformWorkspace
// @Router					/v1/installs/{install_id}/terraform-workspaces/{workspace_id} [get]
func (s *service) GetInstallTerraformWorkspace(ctx *gin.Context) {
	s.GetTerraformWorkpace(ctx)
}

// @ID						GetInstallTerraformWorkspaceLock
// @Summary				get terraform workspace lock
// @Description			Ancestor-scoped alternative to the bare route: authorizes against the install that owns the resource, so an identity scoped to that install can reach it.
// @Param					install_id	path	string	true	"install ID"
// @Param					workspace_id	path	string	true	"workspace ID"
// @Tags					runners
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	app.TerraformWorkspaceLock
// @Router					/v1/installs/{install_id}/terraform-workspaces/{workspace_id}/lock [get]
func (s *service) GetInstallTerraformWorkspaceLock(ctx *gin.Context) {
	s.GetTerraformWorkspaceLock(ctx)
}

// @ID						LockInstallTerraformWorkspace
// @Summary				lock terraform state
// @Description			Ancestor-scoped alternative to the bare route: authorizes against the install that owns the resource, so an identity scoped to that install can reach it.
// @Param					install_id	path	string	true	"install ID"
// @Param					workspace_id	path	string	true	"workspace ID"
// @Param job_id 				query	string	false	"job ID"
// @Tags					runners,runners/runner
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Param					body body interface{} true "terraform workspace lock "
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	app.TerraformWorkspaceState
// @Router					/v1/installs/{install_id}/terraform-workspaces/{workspace_id}/lock [POST]
func (s *service) LockInstallTerraformWorkspace(ctx *gin.Context) {
	s.LockTerraformWorkspace(ctx)
}

// @ID						UnlockInstallTerraformWorkspace
// @Summary				unlock terraform workspace
// @Description			Ancestor-scoped alternative to the bare route: authorizes against the install that owns the resource, so an identity scoped to that install can reach it.
// @Param					install_id	path	string	true	"install ID"
// @Param					workspace_id	path	string	true	"workspace ID"
// @Tags					runners,runners/runner
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Param					body body interface{} true "terraform workspace unlock "
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	app.TerraformWorkspaceState
// @Router					/v1/installs/{install_id}/terraform-workspaces/{workspace_id}/unlock [POST]
func (s *service) UnlockInstallTerraformWorkspace(ctx *gin.Context) {
	s.UnlockTerraformWorkspace(ctx)
}

// @ID						GetInstallTerraformWorkspaceStates
// @Summary				get terraform states
// @Description			Ancestor-scoped alternative to the bare route: authorizes against the install that owns the resource, so an identity scoped to that install can reach it.
// @Param					install_id	path	string	true	"install ID"
// @Param					workspace_id	path	string	true	"workspace ID"
// @Param					offset						query	int		false	"offset of results to return"	Default(0)
// @Param					limit						query	int		false	"limit of results to return"	Default(10)
// @Param					page						query	int		false	"page number of results to return"	Default(0)
// @Tags					runners
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{array}	app.TerraformWorkspaceState
// @Router					/v1/installs/{install_id}/terraform-workspaces/{workspace_id}/states [get]
func (s *service) GetInstallTerraformWorkspaceStates(ctx *gin.Context) {
	s.GetTerraformWorkspaceStatesV2(ctx)
}

// @ID						GetInstallTerraformWorkspaceState
// @Summary				get terraform state by ID
// @Description			Ancestor-scoped alternative to the bare route: authorizes against the install that owns the resource, so an identity scoped to that install can reach it.
// @Param					install_id	path	string	true	"install ID"
// @Param					workspace_id	path	string	true	"workspace ID"
// @Param					state_id 		path	string	true	"state ID"
// @Tags					runners
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	app.TerraformWorkspaceState
// @Router					/v1/installs/{install_id}/terraform-workspaces/{workspace_id}/states/{state_id} [get]
func (s *service) GetInstallTerraformWorkspaceState(ctx *gin.Context) {
	s.GetTerraformWorkspaceStateByIDV2(ctx)
}

// @ID						GetInstallTerraformWorkspaceStateResources
// @Summary				get terraform state resources. This output is similar to "terraform state list"
// @Description			Ancestor-scoped alternative to the bare route: authorizes against the install that owns the resource, so an identity scoped to that install can reach it.
// @Param					install_id	path	string	true	"install ID"
// @Param					workspace_id	path	string	true	"workspace ID"
// @Param					state_id 		path	string	true	"state ID"
// @Tags					runners
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
// @Router					/v1/installs/{install_id}/terraform-workspaces/{workspace_id}/states/{state_id}/resources [get]
func (s *service) GetInstallTerraformWorkspaceStateResources(ctx *gin.Context) {
	s.GetTerraformWorkspaceStateResourcesV2(ctx)
}

// @ID						GetInstallTerraformWorkspaceStatesJSON
// @Summary				get terraform states json
// @Description			Ancestor-scoped alternative to the bare route: authorizes against the install that owns the resource, so an identity scoped to that install can reach it.
// @Param					install_id	path	string	true	"install ID"
// @Param					workspace_id	path	string	true	"workspace ID"
// @Param					offset						query	int		false	"offset of results to return"	Default(0)
// @Param					limit						query	int		false	"limit of results to return"	Default(10)
// @Param					page						query	int		false	"page number of results to return"	Default(0)
// @Tags					runners
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{array}	app.TerraformWorkspaceStateJSON
// @Router					/v1/installs/{install_id}/terraform-workspaces/{workspace_id}/state-json [get]
func (s *service) GetInstallTerraformWorkspaceStatesJSON(ctx *gin.Context) {
	s.GetTerraformWorkspaceStatesJSONV2(ctx)
}

// @ID						GetInstallTerraformWorkspaceStateJSON
// @Summary				get terraform state json by id. This output is same as "terraform show --json"
// @Description			Ancestor-scoped alternative to the bare route: authorizes against the install that owns the resource, so an identity scoped to that install can reach it.
// @Param					install_id	path	string	true	"install ID"
// @Param					workspace_id	path	string	true	"workspace ID"
// @Param					state_id					path	string	true	"terraform state ID"
// @Tags					runners
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	object{}
// @Router					/v1/installs/{install_id}/terraform-workspaces/{workspace_id}/state-json/{state_id} [get]
func (s *service) GetInstallTerraformWorkspaceStateJSON(ctx *gin.Context) {
	s.GetTerraformWorkspaceStatesJSONByIDV2(ctx)
}

// @ID						GetInstallTerraformWorkspaceStateJSONRaw
// @Summary				get raw workspace state json by id
// @Description			Ancestor-scoped alternative to the bare route: authorizes against the install that owns the resource, so an identity scoped to that install can reach it.
// @Param					install_id	path	string	true	"install ID"
// @Param					workspace_id	path	string	true	"workspace ID"
// @Param					state_id		path	string	true	"state ID"
// @Tags					runners
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	object{}
// @Router					/v1/installs/{install_id}/terraform-workspaces/{workspace_id}/state-json/{state_id}/raw [get]
func (s *service) GetInstallTerraformWorkspaceStateJSONRaw(ctx *gin.Context) {
	s.GetWorkspaceStateJSONRawByID(ctx)
}

// @ID						GetInstallTerraformWorkspaceStateJSONResources
// @Summary				get terraform state resources. This output is similar to "terraform state list"
// @Description			Ancestor-scoped alternative to the bare route: authorizes against the install that owns the resource, so an identity scoped to that install can reach it.
// @Param					install_id	path	string	true	"install ID"
// @Param					workspace_id	path	string	true	"workspace ID"
// @Param					state_id 		path	string	true	"state ID"
// @Tags					runners
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
// @Router					/v1/installs/{install_id}/terraform-workspaces/{workspace_id}/state-json/{state_id}/resources [get]
func (s *service) GetInstallTerraformWorkspaceStateJSONResources(ctx *gin.Context) {
	s.GetTerraformWorkspaceStateResourcesV2(ctx)
}

// @ID						ReadComponentBuildLogs
// @Summary				read a log stream's logs
// @Description			Ancestor-scoped alternative to the bare route: authorizes against the app and component that own the build, which may serve many installs.
// @Param					app_id	path	string	true	"app ID"
// @Param					component_id	path	string	true	"component ID"
// @Param					build_id	path	string	true	"build ID"
// @Param					X-Nuon-API-Offset	header	string	false	"log stream offset"
// @Param					order				query	string	false	"sort direction"	default(asc)
// @Param					service_name		query	[]string	false	"filter by service_name (repeatable)"
// @Param					scope_name			query	[]string	false	"filter by scope_name (repeatable)"
// @Param					severity_text		query	[]string	false	"filter by severity_text (repeatable)"
// @Param					tool				query	string		false	"filter by log_attributes['nuon.tool']"
// @Param					helm_release_name	query	string		false	"filter by log_attributes['helm.release_name']"
// @Param					helm_operation		query	string		false	"filter by log_attributes['helm.operation']"
// @Param					tf_workspace_id		query	string		false	"filter by log_attributes['tf.workspace_id']"
// @Param					tf_operation		query	string		false	"filter by log_attributes['tf.operation']"
// @Param					k8s_kind			query	string		false	"filter by log_attributes['k8s.kind']"
// @Param					k8s_namespace		query	string		false	"filter by log_attributes['k8s.namespace']"
// @Param					k8s_name			query	string		false	"filter by log_attributes['k8s.name']"
// @Param					trace_id			query	string		false	"filter by exact trace_id (dedicated CH column)"
// @Param					span_id				query	string		false	"filter by exact span_id (dedicated CH column)"
// @Param					q					query	string		false	"case-insensitive substring filter on log body"
// @Tags					runners
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	[]app.OtelLogRecord
// @Router					/v1/apps/{app_id}/components/{component_id}/builds/{build_id}/logs [GET]
func (s *service) ReadComponentBuildLogs(ctx *gin.Context) {
	s.LogStreamReadLogs(ctx)
}

// @ID						TailComponentBuildLogs
// @Summary				long-poll tail a log stream
// @Description			Ancestor-scoped alternative to the bare route: authorizes against the app and component that own the build, which may serve many installs.
// @Param					app_id	path	string	true	"app ID"
// @Param					component_id	path	string	true	"component ID"
// @Param					build_id	path	string	true	"build ID"
// @Param					since			query	string	false	"composite cursor in the form `<unix_nano>:<id>`; empty starts from the oldest row"
// @Param					wait			query	string	false	"max wait for new rows (Go duration, capped server-side at 30s)"
// @Tags					runners
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	LogStreamTailLogsResponse
// @Router					/v1/apps/{app_id}/components/{component_id}/builds/{build_id}/logs/tail [GET]
func (s *service) TailComponentBuildLogs(ctx *gin.Context) {
	s.LogStreamTailLogs(ctx)
}

// @ID						ReadComponentBuildSpans
// @Summary				read a log stream's trace spans
// @Description			Ancestor-scoped alternative to the bare route: authorizes against the app and component that own the build, which may serve many installs.
// @Param					app_id	path	string	true	"app ID"
// @Param					component_id	path	string	true	"component ID"
// @Param					build_id	path	string	true	"build ID"
// @Tags					runners
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	[]LogStreamSpan
// @Router					/v1/apps/{app_id}/components/{component_id}/builds/{build_id}/spans [GET]
func (s *service) ReadComponentBuildSpans(ctx *gin.Context) {
	s.LogStreamReadSpans(ctx)
}

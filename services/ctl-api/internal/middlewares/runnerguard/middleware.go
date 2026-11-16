package runnerguard

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/shortid"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

const (
	runnerPrefix = "run"
	jobPrefix    = "job"
	logPrefix    = "log"
)

type Middleware struct {
	db *gorm.DB
}

type Params struct {
	fx.In

	DB *gorm.DB `name:"psql"`
}

func New(params Params) *Middleware {
	return &Middleware{db: params.DB}
}

func (m *Middleware) JobHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if cctx.IsGlobal(ctx) || cctx.IsPublic(ctx) {
			ctx.Next()
			return
		}
		acct, err := cctx.AccountFromGinContext(ctx)
		if err != nil {
			ctx.Next()
			return
		}
		prefix, prefixErr := shortid.GetPrefix(acct.Subject)
		jobAccount := prefixErr == nil && prefix == jobPrefix
		method := ctx.Request.Method
		path := ctx.FullPath()
		if isJobTokenExchange(method, path) {
			if jobAccount {
				deny(ctx, "job token cannot claim a job")
				return
			}
			ctx.Next()
			return
		}
		if isJobWrite(method, path) {
			jobID := jobIDFromPath(ctx)
			if !jobAccount || acct.Subject != jobID {
				deny(ctx, "job writes require the job token")
				return
			}
			ctx.Next()
			return
		}
		// Job tokens may call other runner-API routes needed to run the job;
		// InstallHandler scopes them to their job → runner → install.
		ctx.Next()
	}
}

func (m *Middleware) InstallHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if cctx.IsGlobal(ctx) || cctx.IsPublic(ctx) {
			ctx.Next()
			return
		}
		acct, err := cctx.AccountFromGinContext(ctx)
		if err != nil {
			ctx.Next()
			return
		}
		prefix, prefixErr := shortid.GetPrefix(acct.Subject)
		if prefixErr == nil && prefix == logPrefix {
			m.authorizeLogToken(ctx, acct.Subject)
			return
		}
		if prefixErr != nil || (prefix != runnerPrefix && prefix != jobPrefix) {
			ctx.Next()
			return
		}

		runnerID := acct.Subject
		if prefix == jobPrefix {
			var job app.RunnerJob
			if err := m.db.WithContext(ctx.Request.Context()).Select("id", "runner_id").
				Where(app.RunnerJob{ID: acct.Subject}).Take(&job).Error; err != nil {
				notFound(ctx, err)
				return
			}
			runnerID = job.RunnerID
			if id := jobIDFromPath(ctx); id != "" && id != acct.Subject {
				notFound(ctx, fmt.Errorf("job %s", id))
				return
			}
		}

		var runner app.Runner
		if err := m.db.WithContext(ctx.Request.Context()).Preload("RunnerGroup").
			Where(app.Runner{ID: runnerID}).Take(&runner).Error; err != nil {
			notFound(ctx, err)
			return
		}
		installID := ""
		if runner.RunnerGroup.Type == app.RunnerGroupTypeInstall {
			installID = runner.RunnerGroup.OwnerID
		}

		if id := ctx.Param("runner_id"); id != "" && id != runner.ID {
			notFound(ctx, fmt.Errorf("runner %s", id))
			return
		}
		if id := ctx.Param("install_id"); id != "" && (installID == "" || id != installID) {
			notFound(ctx, fmt.Errorf("install %s", id))
			return
		}
		if err := m.authorizeJob(ctx, runner.ID); err != nil {
			notFound(ctx, err)
			return
		}
		if err := m.authorizeProcess(ctx, runner.ID); err != nil {
			notFound(ctx, err)
			return
		}
		if err := m.authorizeOwnedResource(ctx, runner.OrgID, installID); err != nil {
			notFound(ctx, err)
			return
		}
		cctx.SetRunnerGinContext(ctx, &runner)
		ctx.Next()
	}
}

func (m *Middleware) authorizeLogToken(ctx *gin.Context, subject string) {
	if ctx.Request.Method == http.MethodPost &&
		ctx.FullPath() == "/v1/log-streams/:log_stream_id/logs" &&
		ctx.Param("log_stream_id") == subject {
		ctx.Next()
		return
	}
	notFound(ctx, fmt.Errorf("log token %s", subject))
}

func (m *Middleware) authorizeJob(ctx *gin.Context, runnerID string) error {
	jobID := ""
	for _, param := range []string{"job_id", "runner_job_id"} {
		if id := ctx.Param(param); id != "" {
			jobID = id
			break
		}
	}
	if jobID != "" {
		var job app.RunnerJob
		if err := m.db.WithContext(ctx.Request.Context()).Select("id", "runner_id").
			Where(app.RunnerJob{ID: jobID, RunnerID: runnerID}).Take(&job).Error; err != nil {
			return fmt.Errorf("job %s: %w", jobID, err)
		}
	}
	execID := ctx.Param("runner_job_execution_id")
	if execID == "" {
		return nil
	}
	var execution app.RunnerJobExecution
	if err := m.db.WithContext(ctx.Request.Context()).Select("id", "runner_job_id").
		Where(app.RunnerJobExecution{ID: execID}).Take(&execution).Error; err != nil {
		return fmt.Errorf("execution %s: %w", execID, err)
	}
	if jobID != "" && execution.RunnerJobID != jobID {
		return fmt.Errorf("execution %s is not part of job %s", execID, jobID)
	}
	var job app.RunnerJob
	if err := m.db.WithContext(ctx.Request.Context()).Select("id", "runner_id").
		Where(app.RunnerJob{ID: execution.RunnerJobID, RunnerID: runnerID}).Take(&job).Error; err != nil {
		return fmt.Errorf("execution %s: %w", execID, err)
	}
	return nil
}

func (m *Middleware) authorizeProcess(ctx *gin.Context, runnerID string) error {
	processID := ctx.Param("process_id")
	if processID != "" {
		var process app.RunnerProcess
		if err := m.db.WithContext(ctx.Request.Context()).Select("id", "runner_id").
			Where(app.RunnerProcess{ID: processID, RunnerID: runnerID}).Take(&process).Error; err != nil {
			return fmt.Errorf("process %s: %w", processID, err)
		}
	}
	shutdownID := ctx.Param("shutdown_id")
	if shutdownID == "" {
		return nil
	}
	var shutdown app.RunnerProcessShutdown
	if err := m.db.WithContext(ctx.Request.Context()).Select("id", "runner_process_id").
		Where(app.RunnerProcessShutdown{ID: shutdownID}).Take(&shutdown).Error; err != nil {
		return fmt.Errorf("shutdown %s: %w", shutdownID, err)
	}
	if processID != "" && shutdown.RunnerProcessID != processID {
		return fmt.Errorf("shutdown %s is not part of process %s", shutdownID, processID)
	}
	var process app.RunnerProcess
	if err := m.db.WithContext(ctx.Request.Context()).Select("id", "runner_id").
		Where(app.RunnerProcess{ID: shutdown.RunnerProcessID, RunnerID: runnerID}).Take(&process).Error; err != nil {
		return fmt.Errorf("shutdown %s: %w", shutdownID, err)
	}
	return nil
}

func (m *Middleware) authorizeOwnedResource(ctx *gin.Context, orgID, installID string) error {
	if id := firstID(ctx, "workspace_id"); id != "" {
		if err := m.ownerMatchesInstall(ctx, orgID, installID, &app.TerraformWorkspace{}, id); err != nil {
			return err
		}
	}
	if id := ctx.Param("helm_chart_id"); id != "" {
		if err := m.ownerMatchesInstall(ctx, orgID, installID, &app.HelmChart{}, id); err != nil {
			return err
		}
	}
	if id := ctx.Param("log_stream_id"); id != "" {
		return m.authorizeLogStream(ctx, orgID, installID, id)
	}
	return nil
}

// authorizeLogStream allows the runner root token to write OTLP to its own
// process log streams (owner_type=runner_processes). Job / install owned
// streams still require install ownership.
func (m *Middleware) authorizeLogStream(ctx *gin.Context, orgID, installID, logStreamID string) error {
	var stream app.LogStream
	if err := m.db.WithContext(ctx.Request.Context()).
		Select("id", "org_id", "owner_id", "owner_type").
		Where("id = ?", logStreamID).
		Take(&stream).Error; err != nil {
		return err
	}
	if stream.OrgID != "" && stream.OrgID != orgID {
		return fmt.Errorf("resource %s is outside the runner org", logStreamID)
	}
	if stream.OwnerType == "runner_processes" {
		acct, err := cctx.AccountFromGinContext(ctx)
		if err != nil {
			return err
		}
		runnerID := acct.Subject
		if prefix, err := shortid.GetPrefix(acct.Subject); err == nil && prefix == jobPrefix {
			var job app.RunnerJob
			if err := m.db.WithContext(ctx.Request.Context()).Select("id", "runner_id").
				Where(app.RunnerJob{ID: acct.Subject}).Take(&job).Error; err != nil {
				return fmt.Errorf("log stream %s: %w", logStreamID, err)
			}
			runnerID = job.RunnerID
		}
		var process app.RunnerProcess
		q := m.db.WithContext(ctx.Request.Context()).
			Select("id", "runner_id").
			Where("runner_id = ?", runnerID)
		if stream.OwnerID != "" {
			q = q.Where("id = ?", stream.OwnerID)
		} else {
			// legacy streams created before owner_id was bound
			q = q.Where("log_stream_id = ?", logStreamID)
		}
		if err := q.Take(&process).Error; err != nil {
			return fmt.Errorf("log stream %s: %w", logStreamID, err)
		}
		return nil
	}
	return m.ownerMatchesInstall(ctx, orgID, installID, &app.LogStream{}, logStreamID)
}

func firstID(ctx *gin.Context, name string) string {
	if id := ctx.Param(name); id != "" {
		return id
	}
	return ctx.Query(name)
}

func (m *Middleware) ownerMatchesInstall(ctx *gin.Context, orgID, installID string, model interface{}, id string) error {
	var row struct {
		OrgID     string
		OwnerID   string
		OwnerType string
	}
	err := m.db.WithContext(ctx.Request.Context()).Model(model).
		Select("org_id", "owner_id", "owner_type").
		Where("id = ?", id).
		Take(&row).Error
	if err != nil {
		return err
	}
	if row.OrgID != "" && row.OrgID != orgID {
		return fmt.Errorf("resource %s is outside the runner org", id)
	}
	resolved, err := ResolveInstallID(ctx.Request.Context(), m.db, row.OwnerType, row.OwnerID)
	if err != nil {
		return err
	}
	if installID == "" || resolved != installID {
		return fmt.Errorf("resource %s belongs to install %s", id, resolved)
	}
	return nil
}

func deny(ctx *gin.Context, description string) {
	ctx.Error(stderr.ErrAuthorization{
		Err:         errors.New(description),
		Description: description,
	})
	ctx.Abort()
}

func notFound(ctx *gin.Context, err error) {
	cctx.GetLogger(ctx, zap.NewNop()).Info("runner install authorization denied",
		zap.String("route", ctx.FullPath()),
		zap.Error(err),
	)
	ctx.Error(stderr.ErrNotFound{Err: errors.New("not found"), Description: "not found"})
	ctx.Abort()
}

func jobIDFromPath(ctx *gin.Context) string {
	if id := ctx.Param("job_id"); id != "" {
		return id
	}
	return ctx.Param("runner_job_id")
}

func isJobTokenExchange(method, path string) bool {
	return method == http.MethodPost && path == "/v1/runners/:runner_id/jobs/:job_id/token"
}

func isJobWrite(method, path string) bool {
	switch path {
	case "/v1/runners/:runner_id/jobs/:job_id":
		return method == http.MethodPatch
	case "/v1/runner-jobs/:runner_job_id":
		return method == http.MethodPatch
	case "/v1/runner-jobs/:runner_job_id/executions":
		return method == http.MethodPost
	case "/v1/runner-jobs/:runner_job_id/executions/:runner_job_execution_id":
		return method == http.MethodPatch
	case "/v1/runner-jobs/:runner_job_id/executions/:runner_job_execution_id/result",
		"/v1/runner-jobs/:runner_job_id/executions/:runner_job_execution_id/outputs":
		return method == http.MethodPost
	default:
		return false
	}
}

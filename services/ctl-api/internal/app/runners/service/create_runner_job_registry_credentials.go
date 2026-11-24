package service

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/plugins/configs"
	"github.com/nuonco/nuon/pkg/shortid"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/runnerguard"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type CreateRunnerJobRegistryCredentialsResponse struct {
	Username      string `json:"username"`
	Password      string `json:"password"`
	ServerAddress string `json:"server_address"`
	Repository    string `json:"repository"`
}

// @ID CreateRunnerJobRegistryCredentials
// @Summary Create registry credentials for a runner job
// @Description.markdown create_runner_job_registry_credentials.md
// @Tags runners/runner
// @Produce json
// @Security APIKey
// @Security OrgID
// @Param runner_job_id path string true "runner job ID"
// @Success 200 {object} CreateRunnerJobRegistryCredentialsResponse
// @Failure 401 {object} stderr.ErrResponse
// @Failure 403 {object} stderr.ErrResponse
// @Failure 404 {object} stderr.ErrResponse
// @Failure 409 {object} stderr.ErrResponse
// @Failure 500 {object} stderr.ErrResponse
// @Router /v1/runner-jobs/{runner_job_id}/registry-credentials [POST]
func (s *service) CreateRunnerJobRegistryCredentials(ctx *gin.Context) {
	jobID := ctx.Param("runner_job_id")

	acct, err := cctx.AccountFromGinContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	if prefix, prefixErr := shortid.GetPrefix(acct.Subject); prefixErr != nil || prefix != "job" || acct.Subject != jobID {
		ctx.Error(stderr.ErrAuthorization{
			Err:         fmt.Errorf("token subject %s cannot mint registry credentials for job %s", acct.Subject, jobID),
			Description: "registry credentials require the job's own token",
		})
		return
	}
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	var job app.RunnerJob
	if err := s.db.WithContext(ctx.Request.Context()).
		Where(app.RunnerJob{ID: jobID, OrgID: org.ID}).
		First(&job).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.Error(stderr.ErrNotFound{Err: err, Description: "job not found"})
			return
		}
		ctx.Error(fmt.Errorf("unable to get runner job: %w", err))
		return
	}
	if !runnerJobAcceptsCredentials(job.Status) {
		ctx.Error(stderr.ErrConflict{
			Err:         fmt.Errorf("job %s is %s", job.ID, job.Status),
			Description: "job is not running",
		})
		return
	}

	cp, err := s.getOrgRunnerJobCompositePlan(ctx.Request.Context(), job.ID, org.ID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get runner job plan: %w", err))
		return
	}
	registries, push := managedRegistries(cp)
	if len(registries) == 0 {
		ctx.Error(stderr.ErrNotFound{
			Err:         fmt.Errorf("job %s plan has no managed registry", job.ID),
			Description: "job plan has no managed registry",
		})
		return
	}

	installID, err := runnerguard.ResolveInstallID(ctx.Request.Context(), s.db, job.OwnerType, job.OwnerID)
	if err != nil {
		ctx.Error(stderr.ErrNotFound{
			Err:         fmt.Errorf("job %s owner: %w", job.ID, err),
			Description: "job is not owned by an install",
		})
		return
	}
	var install app.Install
	if err := s.db.WithContext(ctx.Request.Context()).
		Where(app.Install{ID: installID, OrgID: org.ID}).
		First(&install).Error; err != nil {
		ctx.Error(fmt.Errorf("unable to get install %s: %w", installID, err))
		return
	}

	for _, registry := range registries {
		ok, err := s.orgRegistry.Matches(ctx.Request.Context(), install.AppID, registry)
		if err != nil {
			ctx.Error(fmt.Errorf("unable to resolve org registry repository: %w", err))
			return
		}
		if !ok {
			ctx.Error(stderr.ErrNotFound{
				Err:         fmt.Errorf("job %s plan registry %s/%s is not app %s repository", job.ID, registry.LoginServer, registry.Repository, install.AppID),
				Description: "job plan registry does not match the app repository",
			})
			return
		}
	}

	cred, err := s.orgRegistry.Mint(ctx.Request.Context(), org.ID, install.AppID, push)
	if err != nil {
		ctx.Error(stderr.ErrSystem{
			Err:         fmt.Errorf("mint registry credentials for job %s: %w", job.ID, err),
			Description: "unable to create registry credentials",
		})
		return
	}

	s.l.Info("minted job registry credentials",
		zap.String("runner_job.id", job.ID),
		zap.String("runner_job.type", string(job.Type)),
		zap.String("install.id", install.ID),
		zap.String("app.id", install.AppID),
		zap.Bool("push", push),
	)

	ctx.Header("Cache-Control", "no-store")
	ctx.Header("Pragma", "no-cache")
	ctx.JSON(http.StatusOK, CreateRunnerJobRegistryCredentialsResponse{
		Username:      cred.Username,
		Password:      cred.Password,
		ServerAddress: cred.ServerAddress,
		Repository:    cred.Repository,
	})
}

func runnerJobAcceptsCredentials(status app.RunnerJobStatus) bool {
	switch status {
	case app.RunnerJobStatusAvailable, app.RunnerJobStatusInProgress:
		return true
	default:
		return false
	}
}

func managedRegistries(cp *plantypes.CompositePlan) ([]*configs.OCIRegistryRepository, bool) {
	if cp == nil {
		return nil, false
	}

	var (
		out  []*configs.OCIRegistryRepository
		push bool
	)
	add := func(cfg *configs.OCIRegistryRepository, writes bool) {
		if cfg == nil || !cfg.ManagedCredential {
			return
		}
		out = append(out, cfg)
		push = push || writes
	}

	if p := cp.SyncOCIPlan; p != nil {
		add(p.Src, false)
		add(p.Dst, true)
	}
	if p := cp.ActionWorkflowRunPlan; p != nil {
		add(p.ImageRegistry, false)
	}
	if p := cp.SandboxRunPlan; p != nil && p.OCISource != nil {
		add(p.OCISource.Registry, false)
	}

	return out, push
}

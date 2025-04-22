package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"gorm.io/gorm/clause"
)

type AdminaBackfillInstallSandboxesRequest struct{}

type AdminaBackfillInstallSandboxesResponse struct {
	InstallSandboxIDs []string `json:"install_sandbox_ids"`
}

// @ID						AdminaBackfillInstallSandboxes
// @Description.markdown	admin_backfill_install_components.md
// @Tags					installs/admin
// @Security				AdminEmail
// @Accept					json
// @Param					req	body	AdminaBackfillInstallSandboxesRequest	true	"Input"
// @Produce				json
// @Success				200	{object}	AdminaBackfillInstallSandboxesResponse
// @Router					/v1/installs/admin-backfill-install-sandboxes [POST]
func (s *service) AdminaBackfillInstallSandboxes(ctx *gin.Context) {
	// get installs without install_sandbox_runs
	installs := []app.Install{}
	res := s.db.WithContext(ctx).
		Raw(`
    	SELECT 
    			"installs".*
			FROM 
    			"installs"
			LEFT JOIN 
    			install_sandboxes 
			ON 
    			install_sandboxes.install_id = installs.id
			WHERE 
    			install_sandboxes.id IS NULL
		LIMIT 100
		`).
		Limit(100).
		Scan(&installs)
	if res.Error != nil {
		ctx.Error(errors.Wrap(res.Error, "unable to get installs"))
		return
	}

	installIDVisited := make(map[string]bool)
	installSandboxes := make([]app.InstallSandbox, 0)
	for _, install := range installs {
		installSandboxRuns := []app.InstallSandboxRun{}
		res = s.db.WithContext(ctx).
			Where("install_id = ?", install.ID).
			// install sandbox is null
			Where("install_sandbox_id IS NULL").
			Order("created_at desc").
			Limit(1).
			Find(&installSandboxRuns)

		if res.Error != nil {
			ctx.Error(errors.Wrap(res.Error, "unable to get sandbox run"))
			return
		}

		status := app.InstallSandboxStatusUnknown
		if len(installSandboxRuns) > 0 {
			status = app.SandboxRunStatusToInstallSandboxStatus(installSandboxRuns[0].Status)
		}

		if _, ok := installIDVisited[install.ID]; ok {
			// already visited this install, skip
			continue
		}

		installIDVisited[install.ID] = true
		installSandboxes = append(installSandboxes, app.InstallSandbox{
			InstallID: install.ID,
			OrgID:     install.OrgID,
			Status:    status,
		})

	}

	if len(installSandboxes) < 1 {
		ctx.JSON(http.StatusOK, AdminaBackfillInstallSandboxesResponse{
			InstallSandboxIDs: []string{},
		})
		return
	}

	res = s.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			DoNothing: true,
		}).
		Create(&installSandboxes)
	if res.Error != nil {
		ctx.Error(errors.Wrap(res.Error, "unable to create install sandbox"))
		return
	}

	installSandboxIDs := make([]string, 0)
	for _, installSandbox := range installSandboxes {
		installSandboxIDs = append(installSandboxIDs, installSandbox.ID)
	}

	// update runs for each install sandbox
	for _, installSandbox := range installSandboxes {
		res = s.db.WithContext(ctx).
			Unscoped().
			Model(&app.InstallSandboxRun{}).
			Where("install_id = ?", installSandbox.InstallID).
			Where("install_sandbox_id IS NULL").
			Updates(app.InstallSandboxRun{
				InstallSandboxID: &installSandbox.ID,
			})
		if res.Error != nil {
			ctx.Error(errors.Wrap(res.Error, "unable to update install sandbox run"))
			return
		}
	}

	ctx.JSON(http.StatusOK, AdminaBackfillInstallSandboxesResponse{
		InstallSandboxIDs: installSandboxIDs,
	})
}

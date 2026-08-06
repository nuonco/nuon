package service

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

type createAirgapInstallRequest struct {
	Name string `json:"name" binding:"required"`
}

type airgapInstallResponse struct {
	ID             string    `json:"id"`
	CreatedAt      time.Time `json:"created_at"`
	Name           string    `json:"name"`
	AppID          string    `json:"app_id"`
	AppConfigID    string    `json:"app_config_id"`
	AirgapBundleID string    `json:"airgap_bundle_id"`
}

// @ID CreateAirgapInstall
// @Summary create a virtual install that tracks an air-gapped delivery of a bundle
// @Tags airgap-bundles
// @Accept json
// @Produce json
// @Security APIKey
// @Security OrgID
// @Param app_id path string true "app ID"
// @Param bundle_id path string true "bundle ID"
// @Param request body createAirgapInstallRequest true "install request"
// @Success 201 {object} airgapInstallResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} stderr.ErrResponse
// @Router /v1/apps/{app_id}/airgap-bundles/{bundle_id}/installs [post]
func (s *service) CreateAirgapInstall(ctx *gin.Context) {
	var req createAirgapInstallRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid request: %v", err)})
		return
	}
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	install, err := s.createAirgapInstall(ctx, org.ID, ctx.Param("app_id"), ctx.Param("bundle_id"), req.Name)
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusCreated, airgapInstallResponse{
		ID: install.ID, CreatedAt: install.CreatedAt, Name: install.Name,
		AppID: install.AppID, AppConfigID: install.AppConfigID, AirgapBundleID: install.AirgapBundleID.String,
	})
}

// createAirgapInstall persists only the install row: no runner group, queues,
// provision workflow, lifecycle updates, or signals — none of those exist for
// an air-gapped customer who never connects back to the control plane.
func (s *service) createAirgapInstall(ctx *gin.Context, orgID, appID, bundleID, name string) (*app.Install, error) {
	var bundle app.AirgapBundle
	if err := s.db.WithContext(ctx).Where(app.AirgapBundle{ID: bundleID, OrgID: orgID, AppID: appID}).First(&bundle).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("air-gap bundle %s: %w", bundleID, err)
		}
		return nil, err
	}
	install := app.Install{
		Name:           name,
		AppID:          bundle.AppID,
		AppConfigID:    bundle.AppConfigID,
		AirgapBundleID: generics.NewNullString(bundle.ID),
	}
	// Insert only the columns a virtual install owns; everything else keeps its
	// database default since no provisioning ever fills it in.
	columns := []string{"ID", "CreatedByID", "CreatedAt", "UpdatedAt", "DeletedAt", "OrgID", "Name", "AppID", "AppConfigID", "AirgapBundleID"}
	if err := s.db.WithContext(ctx).Select(columns).Create(&install).Error; err != nil {
		return nil, fmt.Errorf("unable to create air-gap install: %w", err)
	}
	return &install, nil
}

// @ID ListAirgapInstalls
// @Summary list virtual installs tracking air-gapped deliveries of a bundle
// @Tags airgap-bundles
// @Produce json
// @Security APIKey
// @Security OrgID
// @Param app_id path string true "app ID"
// @Param bundle_id path string true "bundle ID"
// @Param offset query int false "offset of results to return" Default(0)
// @Param limit query int false "limit of results to return" Default(10)
// @Param page query int false "page number of results to return" Default(0)
// @Success 200 {array} airgapInstallResponse
// @Failure 500 {object} stderr.ErrResponse
// @Router /v1/apps/{app_id}/airgap-bundles/{bundle_id}/installs [get]
func (s *service) ListAirgapInstalls(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	installs, err := s.listAirgapInstalls(ctx, org.ID, ctx.Param("app_id"), ctx.Param("bundle_id"))
	if err != nil {
		ctx.Error(fmt.Errorf("unable to list air-gap installs: %w", err))
		return
	}
	ctx.JSON(http.StatusOK, installs)
}

// Scans into the response struct directly so the Install model's AfterQuery
// rollups (runner status, org sandbox mode, links) never run: a virtual
// install has none of that state.
func (s *service) listAirgapInstalls(ctx *gin.Context, orgID, appID, bundleID string) ([]airgapInstallResponse, error) {
	installs := []airgapInstallResponse{}
	result := s.db.WithContext(ctx).Model(&app.Install{}).Scopes(scopes.WithOffsetPagination).
		Where(app.Install{OrgID: orgID, AppID: appID, AirgapBundleID: generics.NewNullString(bundleID)}).
		Order("created_at DESC").Find(&installs)
	if result.Error != nil {
		return nil, result.Error
	}
	return db.HandlePaginatedResponse(ctx, installs)
}

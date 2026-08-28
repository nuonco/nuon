package service

import (
	"context"
	"fmt"
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/runner/customer_managed/operation"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// @ID ListInstallReleaseDeployments
// @Summary list immutable release deployment history for an install
// @Tags installs
// @Produce json
// @Security APIKey
// @Security OrgID
// @Param install_id path string true "Install ID"
// @Success 200 {array} app.InstallReleaseDeployment
// @Failure 404 {object} stderr.ErrResponse
// @Failure 500 {object} stderr.ErrResponse
// @Router /v1/installs/{install_id}/release-deployments [get]
func (s *service) ListReleaseDeployments(ctx *gin.Context) {
	if !s.customerManagedInstallsEnabled(ctx) {
		return
	}
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	var install app.Install
	if err := s.db.WithContext(ctx).Where(app.Install{ID: ctx.Param("install_id"), OrgID: org.ID}).First(&install).Error; err != nil {
		ctx.Error(err)
		return
	}
	var deployments []app.InstallReleaseDeployment
	if err := s.db.WithContext(ctx).Where(app.InstallReleaseDeployment{OrgID: org.ID, InstallID: install.ID}).Order("finished_at DESC, created_at DESC").Find(&deployments).Error; err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, deployments)
}

func (s *service) syncBundleHistory(ctx context.Context, install *app.Install, history []operation.BundleInfo) error {
	if len(history) == 0 {
		return nil
	}
	sorted := append([]operation.BundleInfo(nil), history...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].ActivatedAt.Before(sorted[j].ActivatedAt) })
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var policy app.InstallManagementPolicyVersion
		if err := tx.Where(app.InstallManagementPolicyVersion{OrgID: install.OrgID, InstallID: install.ID}).Order("version DESC").First(&policy).Error; err != nil {
			return fmt.Errorf("load install management policy: %w", err)
		}
		var previous app.InstallReleaseDeployment
		err := tx.Where(app.InstallReleaseDeployment{OrgID: install.OrgID, InstallID: install.ID, Status: app.InstallDeploymentStatusSucceeded}).Order("finished_at DESC, created_at DESC").First(&previous).Error
		if err != nil && err != gorm.ErrRecordNotFound {
			return fmt.Errorf("load release deployment history: %w", err)
		}
		for _, info := range sorted {
			if info.ActivatedAt.IsZero() || info.Release == nil || info.Package == nil {
				continue
			}
			var count int64
			operationID := info.OperationID
			if operationID == "" {
				if err := tx.Model(&app.InstallReleaseDeployment{}).Where(app.InstallReleaseDeployment{
					OrgID: install.OrgID, InstallID: install.ID, PlanDigest: info.BundleDigest,
					Status: app.InstallDeploymentStatusSucceeded,
				}).Count(&count).Error; err != nil {
					return err
				}
				if count > 0 {
					continue
				}
				operationID = "bundle:" + info.BundleDigest
			}
			if err := tx.Model(&app.InstallReleaseDeployment{}).Where(app.InstallReleaseDeployment{OrgID: install.OrgID, InstallID: install.ID, OperationID: operationID}).Count(&count).Error; err != nil {
				return err
			}
			if count > 0 {
				continue
			}
			var release app.AppRelease
			if err := tx.Where(app.AppRelease{ID: info.Release.ID, OrgID: install.OrgID, AppID: install.AppID, SemanticDigest: info.Release.Digest}).First(&release).Error; err != nil {
				return fmt.Errorf("validate activated release %q: %w", info.Release.ID, err)
			}
			var pkg app.ReleasePackage
			if err := tx.Where(app.ReleasePackage{ID: info.Package.ID, OrgID: install.OrgID, ReleaseID: release.ID, PackageDigest: info.Package.Digest, PlanDigest: info.BundleDigest, Status: app.ReleasePackageStatusActive}).First(&pkg).Error; err != nil {
				return fmt.Errorf("validate activated release package %q: %w", info.Package.ID, err)
			}
			finishedAt := info.ActivatedAt.UTC()
			deployment := app.InstallReleaseDeployment{
				OrgID: install.OrgID, InstallID: install.ID, ReleaseID: release.ID, PackageID: &pkg.ID,
				PolicyVersionID: policy.ID, Method: app.InstallDeploymentMethodDisconnectedLocal,
				Actor: "customer", Executor: "customer-local", OperationID: operationID, PlanDigest: info.BundleDigest,
				ResultDirective: "applied", Status: app.InstallDeploymentStatusSucceeded,
				StartedAt: finishedAt, FinishedAt: &finishedAt,
			}
			if previous.ID != "" {
				deployment.PreviousReleaseID = previous.ReleaseID
			}
			if err := tx.Create(&deployment).Error; err != nil {
				return fmt.Errorf("record activated release %q: %w", release.ID, err)
			}
			previous = deployment
		}
		return nil
	})
}

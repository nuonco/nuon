package service

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	customermanaged "github.com/nuonco/nuon/pkg/runner/customer_managed"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type installRegistrationResponse struct {
	Install      app.Install                        `json:"install"`
	Registration app.InstallRegistration            `json:"registration"`
	Policy       app.InstallManagementPolicyVersion `json:"management_policy"`
	Deployment   app.InstallReleaseDeployment       `json:"release_deployment"`
}

// @ID RegisterInstall
// @Summary register a customer-managed installation
// @Tags installs
// @Accept json
// @Produce json
// @Security APIKey
// @Security OrgID
// @Param request body customermanaged.InstallationRegistration true "installation registration exported by the customer portal"
// @Success 200 {object} installRegistrationResponse
// @Success 201 {object} installRegistrationResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} stderr.ErrResponse
// @Router /v1/install-registrations [post]
func (s *service) RegisterInstall(ctx *gin.Context) {
	if !s.customerManagedInstallsEnabled(ctx) {
		return
	}
	var registration customermanaged.InstallationRegistration
	if err := ctx.ShouldBindJSON(&registration); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid installation registration: %v", err)})
		return
	}
	canonical, err := customermanaged.NewInstallationRegistration(registration)
	if err != nil || canonical.RegistrationID != strings.TrimSpace(registration.RegistrationID) {
		if err == nil {
			err = errors.New("installation registration ID does not match its contents")
		}
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	result, created, err := s.registerInstall(ctx, org.ID, canonical)
	if err != nil {
		ctx.Error(err)
		return
	}
	status := http.StatusOK
	if created {
		status = http.StatusCreated
	}
	ctx.JSON(status, result)
}

func (s *service) registerInstall(ctx *gin.Context, orgID string, registration customermanaged.InstallationRegistration) (*installRegistrationResponse, bool, error) {
	existing, found, err := s.installRegistration(ctx, orgID, registration.RegistrationID)
	if err != nil || found {
		return existing, false, err
	}

	var pkg app.ReleasePackage
	err = s.db.WithContext(ctx).
		Preload("Release").
		Where(app.ReleasePackage{ID: registration.PackageID, OrgID: orgID, Status: app.ReleasePackageStatusActive}).
		First(&pkg).Error
	if err != nil {
		return nil, false, fmt.Errorf("find release package: %w", err)
	}
	if err := validateRegistrationPackage(registration, &pkg); err != nil {
		return nil, false, err
	}

	now := time.Now().UTC()
	nameSuffix := strings.TrimPrefix(registration.RegistrationID, "airreg_")
	if len(nameSuffix) > 6 {
		nameSuffix = nameSuffix[len(nameSuffix)-6:]
	}
	install := app.Install{Name: registration.DeploymentID + "-" + nameSuffix, AppID: pkg.Release.AppID, AppConfigID: pkg.Release.AppConfigID}
	policy := app.InstallManagementPolicyVersion{
		Version: 1, Connectivity: app.InstallConnectivityDisconnected,
		ReleaseSelection: app.InstallReleaseSelectionCustomer, CommandAuthority: app.InstallAuthorityCustomer,
		ApprovalAuthority: app.InstallAuthorityCustomer, Telemetry: app.InstallTelemetryManual,
		EffectiveAt: now,
	}
	record := app.InstallRegistration{
		ID: registration.RegistrationID, ReleaseID: pkg.ReleaseID, PackageID: pkg.ID,
		Source: app.InstallRegistrationSourceManual, Registration: registration,
		IntegrityStatus: "verified", AssociationStatus: "verified", ImportedAt: now,
	}
	finishedAt := registration.InstalledAt.UTC()
	operationID := registration.OperationID
	if operationID == "" {
		operationID = registration.RegistrationID
	}
	deployment := app.InstallReleaseDeployment{
		ReleaseID: pkg.ReleaseID, PackageID: &pkg.ID, Method: app.InstallDeploymentMethodDisconnectedLocal,
		Actor: "customer", Executor: "customer-local", OperationID: operationID,
		PlanDigest:      registration.BundleDigest,
		ResultDirective: "applied", Status: app.InstallDeploymentStatusSucceeded,
		StartedAt: finishedAt, FinishedAt: &finishedAt,
	}

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		columns := []string{"ID", "CreatedByID", "CreatedAt", "UpdatedAt", "DeletedAt", "OrgID", "Name", "AppID", "AppConfigID"}
		if err := tx.Select(columns).Create(&install).Error; err != nil {
			return err
		}
		policy.InstallID, policy.OrgID = install.ID, orgID
		if err := tx.Create(&policy).Error; err != nil {
			return err
		}
		record.InstallID, record.OrgID = install.ID, orgID
		if err := tx.Create(&record).Error; err != nil {
			return err
		}
		deployment.InstallID, deployment.OrgID, deployment.PolicyVersionID = install.ID, orgID, policy.ID
		return tx.Create(&deployment).Error
	})
	if err != nil {
		if existing, found, lookupErr := s.installRegistration(ctx, orgID, registration.RegistrationID); lookupErr == nil && found {
			return existing, false, nil
		}
		return nil, false, fmt.Errorf("register customer-managed install: %w", err)
	}
	return &installRegistrationResponse{Install: install, Registration: record, Policy: policy, Deployment: deployment}, true, nil
}

func validateRegistrationPackage(registration customermanaged.InstallationRegistration, pkg *app.ReleasePackage) error {
	if pkg.ReleaseID != registration.ReleaseID || pkg.Release.SemanticDigest != registration.ReleaseDigest {
		return errors.New("registration release identity does not match the published package")
	}
	if pkg.PackageDigest != registration.PackageDigest {
		return errors.New("registration package digest does not match the published package")
	}
	if "sha256:"+strings.TrimPrefix(pkg.ArchiveChecksum, "sha256:") != registration.ArchiveDigest {
		return errors.New("registration archive digest does not match the published package")
	}
	if pkg.PlanDigest != registration.BundleDigest {
		return errors.New("registration deployment plan digest does not match the published package")
	}
	return nil
}

func (s *service) installRegistration(ctx *gin.Context, orgID, registrationID string) (*installRegistrationResponse, bool, error) {
	var record app.InstallRegistration
	err := s.db.WithContext(ctx).Where(app.InstallRegistration{ID: registrationID, OrgID: orgID}).First(&record).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	var install app.Install
	if err := s.db.WithContext(ctx).Where(app.Install{ID: record.InstallID, OrgID: orgID}).First(&install).Error; err != nil {
		return nil, false, err
	}
	var policy app.InstallManagementPolicyVersion
	if err := s.db.WithContext(ctx).Where(app.InstallManagementPolicyVersion{InstallID: install.ID, OrgID: orgID, Version: 1}).First(&policy).Error; err != nil {
		return nil, false, err
	}
	var deployment app.InstallReleaseDeployment
	if err := s.db.WithContext(ctx).Where(app.InstallReleaseDeployment{InstallID: install.ID, OrgID: orgID}).Order("created_at ASC").First(&deployment).Error; err != nil {
		return nil, false, err
	}
	return &installRegistrationResponse{Install: install, Registration: record, Policy: policy, Deployment: deployment}, true, nil
}

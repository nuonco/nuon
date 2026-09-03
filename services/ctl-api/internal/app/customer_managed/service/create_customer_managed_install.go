package service

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	installshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/provisioning"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type createCustomerManagedInstallRequest struct {
	AppID        string             `json:"app_id"`
	IntendedName string             `json:"intended_name"`
	ReleaseID    string             `json:"release_id"`
	Telemetry    string             `json:"telemetry"`
	AWSRegion    string             `json:"aws_region"`
	AWSAccountID string             `json:"aws_account_id"`
	Inputs       map[string]*string `json:"inputs"`
}

type customerManagedInstallResponse struct {
	Install              app.Install               `json:"install"`
	OperatingModel       app.InstallOperatingModel `json:"operating_model"`
	PortalServiceAccount app.Account               `json:"portal_service_account"`
}

// @ID CreateCustomerManagedInstall
// @Summary create an authenticated customer-managed install
// @Tags installs
// @Accept json
// @Produce json
// @Security APIKey
// @Security OrgID
// @Param request body createCustomerManagedInstallRequest true "Customer-managed install"
// @Success 201 {object} customerManagedInstallResponse
// @Failure 400 {object} map[string]string
// @Failure 403 {object} stderr.ErrResponse
// @Router /v1/customer-managed/installs [post]
func (s *service) CreateCustomerManagedInstall(ctx *gin.Context) {
	if !s.customerManagedInstallsEnabled(ctx) {
		return
	}
	var req createCustomerManagedInstallRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.IntendedName = strings.TrimSpace(req.IntendedName)
	if req.AppID == "" || req.IntendedName == "" || req.ReleaseID == "" || req.AWSRegion == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "app_id, intended_name, release_id, and aws_region are required"})
		return
	}
	if req.Telemetry == "" {
		req.Telemetry = app.InstallTelemetryOperational
	}
	if req.Telemetry != app.InstallTelemetryOperational {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "telemetry must be operational"})
		return
	}
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	var release app.AppRelease
	if err := s.db.WithContext(ctx).Where(app.AppRelease{
		ID: req.ReleaseID, OrgID: org.ID, AppID: req.AppID, Status: app.AppReleaseStatusReady,
	}).First(&release).Error; err != nil {
		ctx.Error(fmt.Errorf("find app release: %w", err))
		return
	}
	if err := s.validateReleaseBuilds(ctx, release); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}
	install, workflow, err := provisioning.PrepareFromRelease(ctx, s.installsHelpers, s.db, &release, &installshelpers.CreateInstallParams{
		Name: req.IntendedName, Inputs: req.Inputs,
		AWSAccount:    &installshelpers.CreateInstallAWSAccountParams{Region: req.AWSRegion, AccountID: req.AWSAccountID},
		InstallConfig: &installshelpers.CreateInstallConfigParams{ApprovalOption: app.InstallApprovalOptionPrompt},
		Metadata:      installshelpers.InstallMetadata{ManagedBy: "customer"},
	})
	if err != nil {
		ctx.Error(fmt.Errorf("create customer-managed install: %w", err))
		return
	}
	operatingModel := app.InstallOperatingModel{
		InstallID: install.ID, Connectivity: app.InstallConnectivityConnected,
		ReleaseSelection:  app.InstallReleaseSelectionVendor,
		ApprovalAuthority: app.InstallAuthorityCustomer, Telemetry: req.Telemetry,
	}
	if err := s.db.WithContext(ctx).Create(&operatingModel).Error; err != nil {
		ctx.Error(fmt.Errorf("create install operating model: %w", err))
		return
	}
	portalAccount, err := s.accountClient.EnsureServiceAccount(ctx, install.ID, "Customer portal: "+install.Name)
	if err != nil {
		ctx.Error(fmt.Errorf("ensure customer portal service account: %w", err))
		return
	}
	if err := s.authzClient.EnsureCustomerPortalInstallRole(ctx, org.ID, install.ID, portalAccount.ID); err != nil {
		ctx.Error(fmt.Errorf("ensure customer portal install role: %w", err))
		return
	}
	if err := provisioning.Enqueue(ctx, s.db, s.queueClient, install, workflow); err != nil {
		ctx.Error(fmt.Errorf("start customer-managed install provisioning: %w", err))
		return
	}
	ctx.JSON(http.StatusCreated, customerManagedInstallResponse{
		Install: *install, OperatingModel: operatingModel, PortalServiceAccount: *portalAccount,
	})
}

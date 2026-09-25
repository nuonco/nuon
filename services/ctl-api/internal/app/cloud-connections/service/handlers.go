package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type CreateRequest struct {
	Name             string                          `json:"name"`
	Platform         app.CloudPlatform               `json:"platform"`
	TargetID         string                          `json:"target_id"`
	Principal        string                          `json:"principal"`
	TenantID         string                          `json:"tenant_id,omitempty"`
	IdentityProvider string                          `json:"identity_provider,omitempty"`
	DefaultRegion    string                          `json:"default_region,omitempty"`
	Capabilities     []app.CloudConnectionCapability `json:"capabilities"`
	Repositories     []string                        `json:"repositories,omitempty"`
	Registry         string                          `json:"registry,omitempty"`
}

type ConnectionResponse struct {
	app.CloudConnection
	Setup  SetupResponse   `json:"setup"`
	UsedBy ConnectionUsage `json:"used_by"`
}

type ConnectionUsage struct {
	Installs   int64 `json:"installs"`
	Components int64 `json:"components"`
}

type VerifyRequest struct {
	Repositories []string `json:"repositories,omitempty"`
	Registry     string   `json:"registry,omitempty"`
}

func userError(err error) error {
	return stderr.ErrUser{Err: err, Description: err.Error()}
}

func (s *service) response(ctx context.Context, connection *app.CloudConnection, options SetupOptions) (ConnectionResponse, error) {
	usage, err := s.usage(ctx, connection.ID)
	if err != nil {
		return ConnectionResponse{}, err
	}
	return ConnectionResponse{CloudConnection: *connection, Setup: s.setup(connection, options), UsedBy: usage}, nil
}

func (s *service) usage(ctx context.Context, connectionID string) (ConnectionUsage, error) {
	var usage ConnectionUsage
	if err := s.db.WithContext(ctx).Model(&app.Install{}).Where(app.Install{CloudConnectionID: &connectionID}).Count(&usage.Installs).Error; err != nil {
		return usage, fmt.Errorf("count installs using cloud connection: %w", err)
	}
	if err := s.db.WithContext(ctx).Model(&app.AWSECRImageConfig{}).Where(app.AWSECRImageConfig{CloudConnectionID: connectionID}).Count(&usage.Components).Error; err != nil {
		return usage, fmt.Errorf("count components using cloud connection: %w", err)
	}
	var azureComponents int64
	if err := s.db.WithContext(ctx).Model(&app.AzureACRImageConfig{}).Where(app.AzureACRImageConfig{CloudConnectionID: connectionID}).Count(&azureComponents).Error; err != nil {
		return usage, fmt.Errorf("count Azure components using cloud connection: %w", err)
	}
	usage.Components += azureComponents
	var gcpComponents int64
	if err := s.db.WithContext(ctx).Model(&app.GCPGARImageConfig{}).Where(app.GCPGARImageConfig{CloudConnectionID: connectionID}).Count(&gcpComponents).Error; err != nil {
		return usage, fmt.Errorf("count GCP components using cloud connection: %w", err)
	}
	usage.Components += gcpComponents
	return usage, nil
}

// @ID CreateCloudConnection
// @Summary create a cloud connection
// @Tags cloud-connections
// @Accept json
// @Produce json
// @Security APIKey
// @Security OrgID
// @Param req body CreateRequest true "Input"
// @Success 201 {object} ConnectionResponse
// @Router /v1/cloud-connections [post]
func (s *service) Create(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	var req CreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}
	connection := app.CloudConnection{OrgID: org.ID, Name: req.Name, Platform: req.Platform, TargetID: req.TargetID, Principal: req.Principal, TenantID: req.TenantID, IdentityProvider: req.IdentityProvider, DefaultRegion: req.DefaultRegion, Capabilities: req.Capabilities}
	if err := validateConnection(&connection); err != nil {
		ctx.Error(userError(err))
		return
	}
	if err := s.db.WithContext(ctx).Create(&connection).Error; err != nil {
		ctx.Error(fmt.Errorf("unable to create cloud connection: %w", err))
		return
	}
	response, err := s.response(ctx, &connection, SetupOptions{Repositories: req.Repositories, Registry: req.Registry})
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusCreated, response)
}

// @ID ListCloudConnections
// @Summary list cloud connections
// @Tags cloud-connections
// @Produce json
// @Security APIKey
// @Security OrgID
// @Success 200 {array} ConnectionResponse
// @Router /v1/cloud-connections [get]
func (s *service) List(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	var connections []app.CloudConnection
	if err := s.db.WithContext(ctx).Where(app.CloudConnection{OrgID: org.ID}).Order("created_at DESC").Find(&connections).Error; err != nil {
		ctx.Error(fmt.Errorf("unable to list cloud connections: %w", err))
		return
	}
	responses := make([]ConnectionResponse, 0, len(connections))
	for i := range connections {
		response, err := s.response(ctx, &connections[i], SetupOptions{})
		if err != nil {
			ctx.Error(err)
			return
		}
		responses = append(responses, response)
	}
	ctx.JSON(http.StatusOK, responses)
}

// @ID GetCloudConnection
// @Summary get a cloud connection
// @Tags cloud-connections
// @Produce json
// @Security APIKey
// @Security OrgID
// @Param connection_id path string true "connection ID"
// @Success 200 {object} ConnectionResponse
// @Router /v1/cloud-connections/{connection_id} [get]
func (s *service) Get(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	connection, err := s.get(ctx, org.ID, ctx.Param("connection_id"))
	if err != nil {
		ctx.Error(err)
		return
	}
	response, err := s.response(ctx, connection, SetupOptions{})
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, response)
}

// @ID GetCloudConnectionSetup
// @Summary get cloud connection setup material
// @Tags cloud-connections
// @Produce json
// @Security APIKey
// @Security OrgID
// @Param connection_id path string true "connection ID"
// @Param repository query []string false "ECR repository names"
// @Success 200 {object} SetupResponse
// @Router /v1/cloud-connections/{connection_id}/setup [get]
func (s *service) Setup(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	connection, err := s.get(ctx, org.ID, ctx.Param("connection_id"))
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, s.setup(connection, SetupOptions{Repositories: ctx.QueryArray("repository"), Registry: ctx.Query("registry")}))
}

// @ID DeleteCloudConnection
// @Summary delete a cloud connection
// @Tags cloud-connections
// @Security APIKey
// @Security OrgID
// @Param connection_id path string true "connection ID"
// @Success 204
// @Router /v1/cloud-connections/{connection_id} [delete]
func (s *service) Delete(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	err = s.delete(ctx, org.ID, ctx.Param("connection_id"))
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.Status(http.StatusNoContent)
}

func (s *service) delete(ctx context.Context, orgID, connectionID string) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var connection app.CloudConnection
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(app.CloudConnection{OrgID: orgID, ID: connectionID}).First(&connection).Error; err != nil {
			return fmt.Errorf("cloud connection not found: %w", err)
		}
		var installReferences int64
		if err := tx.Model(&app.Install{}).Where(app.Install{CloudConnectionID: &connection.ID}).Count(&installReferences).Error; err != nil {
			return fmt.Errorf("count install references: %w", err)
		}
		var componentReferences int64
		if err := tx.Model(&app.AWSECRImageConfig{}).Where(app.AWSECRImageConfig{CloudConnectionID: connection.ID}).Count(&componentReferences).Error; err != nil {
			return fmt.Errorf("count component references: %w", err)
		}
		var azureComponentReferences int64
		if err := tx.Model(&app.AzureACRImageConfig{}).Where(app.AzureACRImageConfig{CloudConnectionID: connection.ID}).Count(&azureComponentReferences).Error; err != nil {
			return fmt.Errorf("count Azure component references: %w", err)
		}
		componentReferences += azureComponentReferences
		var gcpComponentReferences int64
		if err := tx.Model(&app.GCPGARImageConfig{}).Where(app.GCPGARImageConfig{CloudConnectionID: connection.ID}).Count(&gcpComponentReferences).Error; err != nil {
			return fmt.Errorf("count GCP component references: %w", err)
		}
		componentReferences += gcpComponentReferences
		if installReferences > 0 || componentReferences > 0 {
			return stderr.ErrConflict{Err: fmt.Errorf("cloud connection %s is in use", connection.ID), Description: "Cloud connection cannot be deleted while it is in use"}
		}
		return tx.Unscoped().Delete(&connection).Error
	})
}

// @ID VerifyCloudConnection
// @Summary verify a cloud connection
// @Tags cloud-connections
// @Accept json
// @Produce json
// @Security APIKey
// @Security OrgID
// @Param connection_id path string true "connection ID"
// @Param req body VerifyRequest false "Capability probe options"
// @Success 200 {object} ConnectionResponse
// @Router /v1/cloud-connections/{connection_id}/verify [post]
func (s *service) Verify(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	var req VerifyRequest
	if ctx.Request.ContentLength > 0 {
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.Error(stderr.NewInvalidRequest(err))
			return
		}
	}
	options := VerifyOptions{Repositories: req.Repositories, Registry: req.Registry}
	connection, err := s.verify(ctx, org.ID, ctx.Param("connection_id"), options)
	if err != nil {
		ctx.Error(err)
		return
	}
	response, err := s.response(ctx, connection, SetupOptions(options))
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, response)
}

func (s *service) verify(ctx context.Context, orgID, connectionID string, options VerifyOptions) (*app.CloudConnection, error) {
	connection, err := s.getContext(ctx, orgID, connectionID)
	if err != nil {
		return nil, err
	}
	result, err := s.verifier.Verify(ctx, connection, options)
	if err != nil {
		return nil, fmt.Errorf("verify cloud connection: %w", err)
	}
	now := time.Now().UTC()
	update := app.CloudConnection{Status: result.Status, StatusMessage: result.Message, LastVerifiedAt: &now, Capabilities: result.Capabilities, Registries: result.Registries}
	selected := []string{"status", "status_message", "last_verified_at", "capabilities", "registries"}
	if result.Status == app.CloudConnectionStatusVerified {
		update.AuthMode = app.CloudConnectionAuthModeOIDC
		selected = append(selected, "auth_mode")
	}
	// struct-based update so the jsonb serializers on capabilities/registries apply
	if err := s.db.WithContext(ctx).Model(&app.CloudConnection{}).Where(app.CloudConnection{OrgID: orgID, ID: connection.ID, Principal: connection.Principal}).Select(selected).Updates(update).Error; err != nil {
		return nil, fmt.Errorf("save cloud connection verification: %w", err)
	}
	connection, err = s.getContext(ctx, orgID, connection.ID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	return connection, err
}

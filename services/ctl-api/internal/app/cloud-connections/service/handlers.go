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
	Name          string                    `json:"name"`
	Platform      app.CloudPlatform         `json:"platform" swaggertype:"string" enums:"aws"`
	TargetID      string                    `json:"target_id"`
	Principal     string                    `json:"principal"`
	DefaultRegion string                    `json:"default_region,omitempty"`
	Preset        app.CloudConnectionPreset `json:"preset"`
}

type ConnectionResponse struct {
	app.CloudConnection
	Setup                  SetupResponse   `json:"setup"`
	UsedBy                 ConnectionUsage `json:"used_by"`
	VerificationInProgress bool            `json:"verification_in_progress"`
}

type ConnectionUsage struct {
	Installs int64 `json:"installs"`
}

type VerifyRequest struct{}

func userError(err error) error {
	return stderr.ErrUser{Err: err, Description: err.Error()}
}

func (s *service) response(ctx context.Context, connection *app.CloudConnection) (ConnectionResponse, error) {
	usage, err := s.usage(ctx, connection.ID)
	if err != nil {
		return ConnectionResponse{}, err
	}
	response := ConnectionResponse{CloudConnection: *connection, Setup: s.setup(connection), UsedBy: usage, VerificationInProgress: verificationInProgress(connection)}
	if connection.VerificationRequestedAt != nil {
		requestedAt := connection.VerificationRequestedAt.UTC()
		response.VerificationRequestedAt = &requestedAt
	}
	return response, nil
}

func verificationInProgress(connection *app.CloudConnection) bool {
	return connection.VerificationRequestedAt != nil && (connection.LastVerifiedAt == nil || connection.VerificationRequestedAt.After(*connection.LastVerifiedAt))
}

func (s *service) usage(ctx context.Context, connectionID string) (ConnectionUsage, error) {
	var usage ConnectionUsage
	if err := s.db.WithContext(ctx).Model(&app.Install{}).Where(app.Install{CloudConnectionID: &connectionID}).Count(&usage.Installs).Error; err != nil {
		return usage, fmt.Errorf("count installs using cloud connection: %w", err)
	}
	return usage, nil
}

// @ID CreateCloudConnection
// @Summary create a cloud connection
// @Description Create an AWS connection using the stacks or custom preset. Custom renders trust only; attach your own permissions policy.
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
	connection := app.CloudConnection{OrgID: org.ID, Name: req.Name, Platform: req.Platform, TargetID: req.TargetID, Principal: req.Principal, DefaultRegion: req.DefaultRegion, Preset: req.Preset}
	if err := validateConnection(&connection); err != nil {
		ctx.Error(userError(err))
		return
	}
	if err := s.db.WithContext(ctx).Create(&connection).Error; err != nil {
		ctx.Error(fmt.Errorf("unable to create cloud connection: %w", err))
		return
	}
	if _, err := s.helpers.EnsureConnectionQueue(ctx, &connection); err != nil {
		ctx.Error(err)
		return
	}
	response, err := s.response(ctx, &connection)
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
		response, err := s.response(ctx, &connections[i])
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
	response, err := s.response(ctx, connection)
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
	ctx.JSON(http.StatusOK, s.setup(connection))
}

// @ID DeleteCloudConnection
// @Summary delete a cloud connection
// @Tags cloud-connections
// @Security APIKey
// @Security OrgID
// @Param connection_id path string true "connection ID"
// @Success 204
// @Failure 409 {object} stderr.ErrResponse
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
		if installReferences > 0 {
			description := fmt.Sprintf("This connection is used by %d installs. Delete those installs first.", installReferences)
			if installReferences == 1 {
				description = "This connection is used by 1 install. Delete that install first."
			}
			return stderr.ErrConflict{Err: errors.New(description), Description: description}
		}
		if err := s.helpers.TerminateConnectionQueue(ctx, connection.ID); err != nil {
			return fmt.Errorf("terminate cloud connection queue: %w", err)
		}
		return tx.Delete(&connection).Error
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
// @Param req body VerifyRequest false "Input"
// @Success 202 {object} ConnectionResponse
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
	connection, err := s.verify(ctx, org.ID, ctx.Param("connection_id"))
	if err != nil {
		ctx.Error(err)
		return
	}
	response, err := s.response(ctx, connection)
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusAccepted, response)
}

func (s *service) verify(ctx context.Context, orgID, connectionID string) (*app.CloudConnection, error) {
	connection, err := s.getContext(ctx, orgID, connectionID)
	if err != nil {
		return nil, err
	}
	now := s.now().UTC().Truncate(time.Microsecond)
	if verificationInProgress(connection) && connection.VerificationRequestedAt.After(now.Add(-2*time.Minute)) {
		return connection, nil
	}
	if err := s.enqueueVerification(ctx, connection); err != nil {
		return nil, err
	}
	if err := s.db.WithContext(ctx).Model(&app.CloudConnection{}).Where(app.CloudConnection{OrgID: orgID, ID: connection.ID}).
		Select("verification_requested_at").Updates(app.CloudConnection{VerificationRequestedAt: &now}).Error; err != nil {
		return nil, fmt.Errorf("save cloud connection verification request: %w", err)
	}
	connection.VerificationRequestedAt = &now
	return connection, nil
}

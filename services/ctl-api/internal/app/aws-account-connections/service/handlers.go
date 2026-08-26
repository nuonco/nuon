package service

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

type CreateRequest struct {
	Name          string `json:"name"`
	AccountID     string `json:"account_id"`
	DefaultRegion string `json:"default_region"`
}

type PatchRequest struct {
	Name          *string `json:"name"`
	DefaultRegion *string `json:"default_region"`
	RoleARN       *string `json:"role_arn"`
}

type ConnectionResponse struct {
	ID                     string                                     `json:"id"`
	CreatedAt              time.Time                                  `json:"created_at"`
	UpdatedAt              time.Time                                  `json:"updated_at"`
	Name                   string                                     `json:"name"`
	AccountID              string                                     `json:"account_id"`
	DefaultRegion          string                                     `json:"default_region"`
	RoleARN                string                                     `json:"role_arn,omitempty"`
	VerificationStatus     app.AWSAccountConnectionVerificationStatus `json:"verification_status"`
	VerificationCode       string                                     `json:"verification_code,omitempty"`
	VerificationMessage    string                                     `json:"verification_message,omitempty"`
	LastCheckedAt          *time.Time                                 `json:"last_checked_at,omitempty"`
	VerifiedAt             *time.Time                                 `json:"verified_at,omitempty"`
	VerifiedPrincipalARN   string                                     `json:"verified_principal_arn,omitempty"`
	ExternalID             string                                     `json:"external_id,omitempty"`
	ManagementPrincipalARN string                                     `json:"management_principal_arn,omitempty"`
	TrustPolicy            *TrustPolicy                               `json:"trust_policy,omitempty"`
}

type TrustPolicy struct {
	Version   string                 `json:"Version"`
	Statement []TrustPolicyStatement `json:"Statement"`
}

type TrustPolicyStatement struct {
	Effect    string                       `json:"Effect"`
	Principal map[string]string            `json:"Principal"`
	Action    string                       `json:"Action"`
	Condition map[string]map[string]string `json:"Condition"`
}

func response(connection *app.AWSAccountConnection, managementRoleARN string, detail bool) ConnectionResponse {
	result := ConnectionResponse{ID: connection.ID, CreatedAt: connection.CreatedAt, UpdatedAt: connection.UpdatedAt, Name: connection.Name, AccountID: connection.AccountID, DefaultRegion: connection.DefaultRegion, RoleARN: connection.RoleARN, VerificationStatus: connection.VerificationStatus, VerificationCode: connection.VerificationCode, VerificationMessage: connection.VerificationMessage, LastCheckedAt: connection.LastCheckedAt, VerifiedAt: connection.VerifiedAt, VerifiedPrincipalARN: connection.VerifiedPrincipalARN}
	if detail {
		result.ExternalID = connection.ExternalID
		result.ManagementPrincipalARN = managementRoleARN
		result.TrustPolicy = &TrustPolicy{Version: "2012-10-17", Statement: []TrustPolicyStatement{{Effect: "Allow", Principal: map[string]string{"AWS": managementRoleARN}, Action: "sts:AssumeRole", Condition: map[string]map[string]string{"StringEquals": {"sts:ExternalId": connection.ExternalID}}}}}
	}
	return result
}

func userError(err error) error {
	return stderr.ErrUser{Err: err, Description: err.Error()}
}

func (s *service) requireManagementRole() error {
	if s.cfg.ManagementIAMRoleARN == "" {
		return stderr.ErrSystem{Err: fmt.Errorf("management IAM role ARN is not configured"), Description: "AWS account connections are unavailable"}
	}
	if err := validateManagementRoleARN(s.cfg.ManagementIAMRoleARN); err != nil {
		return stderr.ErrSystem{Err: err, Description: "AWS account connections are unavailable"}
	}
	return nil
}

// @ID CreateAWSAccountConnection
// @Summary create an AWS account connection
// @Tags aws-account-connections
// @Accept json
// @Produce json
// @Security APIKey && OrgID
// @Param req body CreateRequest true "Input"
// @Success 201 {object} ConnectionResponse
// @Router /v1/aws-account-connections [post]
func (s *service) Create(ctx *gin.Context) {
	org, err := s.gate(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	if err := s.requireManagementRole(); err != nil {
		ctx.Error(err)
		return
	}
	var req CreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}
	if req.Name == "" {
		ctx.Error(userError(fmt.Errorf("name is required")))
		return
	}
	if err := validateAccountID(req.AccountID); err != nil {
		ctx.Error(userError(err))
		return
	}
	if err := validateRegion(req.DefaultRegion); err != nil {
		ctx.Error(userError(err))
		return
	}
	connection := app.AWSAccountConnection{OrgID: org.ID, Name: req.Name, AccountID: req.AccountID, DefaultRegion: req.DefaultRegion}
	if result := s.db.WithContext(ctx).Create(&connection); result.Error != nil {
		ctx.Error(fmt.Errorf("unable to create AWS account connection: %w", result.Error))
		return
	}
	ctx.JSON(http.StatusCreated, response(&connection, s.cfg.ManagementIAMRoleARN, true))
}

// @ID ListAWSAccountConnections
// @Summary list AWS account connections
// @Tags aws-account-connections
// @Produce json
// @Security APIKey && OrgID
// @Success 200 {array} ConnectionResponse
// @Router /v1/aws-account-connections [get]
func (s *service) List(ctx *gin.Context) {
	org, err := s.gate(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	var connections []app.AWSAccountConnection
	if result := s.db.WithContext(ctx).Where(app.AWSAccountConnection{OrgID: org.ID}).Order("created_at DESC").Find(&connections); result.Error != nil {
		ctx.Error(fmt.Errorf("unable to list AWS account connections: %w", result.Error))
		return
	}
	result := make([]ConnectionResponse, 0, len(connections))
	for i := range connections {
		result = append(result, response(&connections[i], "", false))
	}
	ctx.JSON(http.StatusOK, result)
}

// @ID GetAWSAccountConnection
// @Summary get an AWS account connection
// @Tags aws-account-connections
// @Produce json
// @Security APIKey && OrgID
// @Param connection_id path string true "connection ID"
// @Success 200 {object} ConnectionResponse
// @Router /v1/aws-account-connections/{connection_id} [get]
func (s *service) Get(ctx *gin.Context) {
	org, err := s.gate(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	if err := s.requireManagementRole(); err != nil {
		ctx.Error(err)
		return
	}
	connection, err := s.get(ctx, org.ID, ctx.Param("connection_id"))
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, response(connection, s.cfg.ManagementIAMRoleARN, true))
}

// @ID PatchAWSAccountConnection
// @Summary update an AWS account connection
// @Tags aws-account-connections
// @Accept json
// @Produce json
// @Security APIKey && OrgID
// @Param connection_id path string true "connection ID"
// @Param req body PatchRequest true "Input"
// @Success 200 {object} ConnectionResponse
// @Router /v1/aws-account-connections/{connection_id} [patch]
func (s *service) Patch(ctx *gin.Context) {
	org, err := s.gate(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	if err := s.requireManagementRole(); err != nil {
		ctx.Error(err)
		return
	}
	connection, err := s.get(ctx, org.ID, ctx.Param("connection_id"))
	if err != nil {
		ctx.Error(err)
		return
	}
	var req PatchRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}
	updates := map[string]any{}
	if req.Name != nil {
		if *req.Name == "" {
			ctx.Error(userError(fmt.Errorf("name cannot be empty")))
			return
		}
		updates["name"] = *req.Name
	}
	if req.DefaultRegion != nil {
		if err := validateRegion(*req.DefaultRegion); err != nil {
			ctx.Error(userError(err))
			return
		}
		updates["default_region"] = *req.DefaultRegion
	}
	if req.RoleARN != nil {
		if err := validateRoleARN(*req.RoleARN, connection.AccountID); err != nil {
			ctx.Error(userError(err))
			return
		}
		updates["role_arn"] = *req.RoleARN
		if *req.RoleARN != connection.RoleARN {
			updates["verification_status"] = app.AWSAccountConnectionVerificationPending
			updates["verification_code"] = ""
			updates["verification_message"] = ""
			updates["last_checked_at"] = nil
			updates["verified_at"] = nil
			updates["verified_principal_arn"] = ""
		}
	}
	if len(updates) > 0 {
		if result := s.db.WithContext(ctx).Model(&app.AWSAccountConnection{}).Where(app.AWSAccountConnection{OrgID: org.ID, ID: connection.ID}).Updates(updates); result.Error != nil {
			ctx.Error(fmt.Errorf("unable to update AWS account connection: %w", result.Error))
			return
		}
	}
	connection, err = s.get(ctx, org.ID, connection.ID)
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, response(connection, s.cfg.ManagementIAMRoleARN, true))
}

// @ID DeleteAWSAccountConnection
// @Summary delete an AWS account connection
// @Tags aws-account-connections
// @Security APIKey && OrgID
// @Param connection_id path string true "connection ID"
// @Success 204
// @Router /v1/aws-account-connections/{connection_id} [delete]
func (s *service) Delete(ctx *gin.Context) {
	org, err := s.gate(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var connection app.AWSAccountConnection
		if result := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where(app.AWSAccountConnection{OrgID: org.ID, ID: ctx.Param("connection_id")}).
			First(&connection); result.Error != nil {
			return fmt.Errorf("AWS account connection not found: %w", result.Error)
		}
		var references int64
		if result := tx.Model(&app.AWSAccount{}).
			Where("aws_account_connection_id = ?", connection.ID).
			Count(&references); result.Error != nil {
			return fmt.Errorf("count AWS account connection references: %w", result.Error)
		}
		if references > 0 {
			return stderr.ErrConflict{
				Err:         fmt.Errorf("AWS account connection %s is in use", connection.ID),
				Description: "AWS account connection cannot be deleted while it is used by an install",
			}
		}
		if result := tx.Unscoped().Delete(&connection); result.Error != nil {
			return fmt.Errorf("unable to delete AWS account connection: %w", result.Error)
		}
		return nil
	})
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.Status(http.StatusNoContent)
}

// @ID VerifyAWSAccountConnection
// @Summary verify an AWS account connection
// @Tags aws-account-connections
// @Produce json
// @Security APIKey && OrgID
// @Param connection_id path string true "connection ID"
// @Success 200 {object} ConnectionResponse
// @Router /v1/aws-account-connections/{connection_id}/verify [post]
func (s *service) Verify(ctx *gin.Context) {
	org, err := s.gate(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	if err := s.requireManagementRole(); err != nil {
		ctx.Error(err)
		return
	}
	connection, err := s.get(ctx, org.ID, ctx.Param("connection_id"))
	if err != nil {
		ctx.Error(err)
		return
	}
	if connection.RoleARN == "" {
		ctx.Error(stderr.ErrConflict{Err: fmt.Errorf("role ARN is required before verification"), Description: "Attach a customer role ARN before verification"})
		return
	}
	verification, err := s.verifier.Verify(ctx, s.cfg.ManagementIAMRoleARN, connection.RoleARN, connection.ExternalID, connection.AccountID)
	if err != nil {
		s.l.Error("unable to verify AWS account connection", zap.String("connection-id", connection.ID), zap.Error(err))
		ctx.Error(stderr.ErrSystem{Err: errors.New("AWS account connection verification failed"), Description: "Unable to verify AWS account connection"})
		return
	}
	now := time.Now().UTC()
	updates := map[string]any{"verification_status": verification.Status, "verification_code": verification.Code, "verification_message": verification.Message, "last_checked_at": &now, "verified_at": nil, "verified_principal_arn": ""}
	if verification.Status == string(app.AWSAccountConnectionVerificationVerified) {
		updates["verified_at"] = &now
		updates["verified_principal_arn"] = verification.PrincipalARN
	}
	result := s.db.WithContext(ctx).Model(&app.AWSAccountConnection{}).Where(app.AWSAccountConnection{OrgID: org.ID, ID: connection.ID, RoleARN: connection.RoleARN}).Updates(updates)
	if result.Error != nil {
		ctx.Error(fmt.Errorf("unable to save AWS account connection verification: %w", result.Error))
		return
	}
	if result.RowsAffected == 0 {
		ctx.Error(stderr.ErrConflict{Err: errors.New("role ARN changed during verification"), Description: "The role changed; retry verification"})
		return
	}
	connection, err = s.get(ctx, org.ID, connection.ID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		ctx.Error(err)
		return
	}
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, response(connection, s.cfg.ManagementIAMRoleARN, true))
}

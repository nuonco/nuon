package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	cloudconnections "github.com/nuonco/nuon/services/ctl-api/internal/app/cloud-connections"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

type CreateImageRepositoryRequest struct {
	CloudConnectionID string `json:"cloud_connection_id" binding:"required"`
	Repository        string `json:"repository" binding:"required"`
}

// @ID CreateImageRepository
// @Summary register an image repository
// @Description Save an exact ECR ARN, GAR repository resource, or ACR repository path against an existing org-owned image connection. Pending connections are allowed; image use requires verification. Does not change cloud permissions.
// @Tags image-repositories
// @Accept json
// @Produce json
// @Security APIKey
// @Security OrgID
// @Param req body CreateImageRepositoryRequest true "Input"
// @Success 201 {object} app.ImageRepository
// @Failure 400 {object} stderr.ErrResponse
// @Failure 401 {object} stderr.ErrResponse
// @Failure 403 {object} stderr.ErrResponse
// @Failure 404 {object} stderr.ErrResponse
// @Failure 409 {object} stderr.ErrResponse
// @Failure 500 {object} stderr.ErrResponse
// @Router /v1/image-repositories [post]
func (s *service) CreateImageRepository(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	var req CreateImageRepositoryRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}
	var repository app.ImageRepository
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var connection app.CloudConnection
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(app.CloudConnection{OrgID: org.ID, ID: req.CloudConnectionID}).First(&connection).Error; err != nil {
			return fmt.Errorf("cloud connection not found: %w", err)
		}
		key, prefix, err := cloudconnections.ImageRepositoryCoordinates(&connection, req.Repository)
		if err != nil {
			return userError(err)
		}
		repository = app.ImageRepository{
			OrgID: org.ID, Platform: connection.Platform, Repository: req.Repository,
			CanonicalKey: key, ImageURLPrefix: prefix, CloudConnectionID: connection.ID,
		}
		if err := tx.Create(&repository).Error; err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return stderr.ErrConflict{Err: err, Description: "This repository is already assigned to a Cloud Connection in this organization. Remove its existing Image repositories assignment before adding it again."}
			}
			return fmt.Errorf("unable to register image repository: %w", err)
		}
		repository.CloudConnection = connection
		return nil
	})
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusCreated, repository)
}

// @ID ListImageRepositories
// @Summary list image repositories
// @Tags image-repositories
// @Produce json
// @Security APIKey
// @Security OrgID
// @Param offset query int false "offset of results to return" Default(0)
// @Param limit query int false "limit of results to return" Default(10)
// @Param page query int false "page number of results to return" Default(0)
// @Param q query string false "search by repository or image URL prefix"
// @Param platform query string false "cloud platform" Enums(aws,gcp,azure)
// @Success 200 {array} app.ImageRepository
// @Failure 401 {object} stderr.ErrResponse
// @Failure 403 {object} stderr.ErrResponse
// @Failure 500 {object} stderr.ErrResponse
// @Router /v1/image-repositories [get]
func (s *service) ListImageRepositories(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	query := s.db.WithContext(ctx).Scopes(scopes.WithOffsetPagination).
		Where(app.ImageRepository{OrgID: org.ID}).
		Preload("CloudConnection", app.CloudConnection{OrgID: org.ID})
	if q := ctx.Query("q"); q != "" {
		pattern := "%" + q + "%"
		query = query.Where("repository ILIKE ? OR image_url_prefix ILIKE ?", pattern, pattern)
	}
	if platform := ctx.Query("platform"); platform != "" {
		query = query.Where(app.ImageRepository{Platform: app.CloudPlatform(platform)})
	}
	var repositories []app.ImageRepository
	if err := query.Order("created_at DESC, id DESC").Find(&repositories).Error; err != nil {
		ctx.Error(fmt.Errorf("unable to list image repositories: %w", err))
		return
	}
	repositories, err = db.HandlePaginatedResponse(ctx, repositories)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to handle paginated response: %w", err))
		return
	}
	ctx.JSON(http.StatusOK, repositories)
}

// @ID GetImageRepository
// @Summary get an image repository
// @Tags image-repositories
// @Produce json
// @Security APIKey
// @Security OrgID
// @Param repository_id path string true "repository ID"
// @Success 200 {object} app.ImageRepository
// @Failure 401 {object} stderr.ErrResponse
// @Failure 403 {object} stderr.ErrResponse
// @Failure 404 {object} stderr.ErrResponse
// @Failure 500 {object} stderr.ErrResponse
// @Router /v1/image-repositories/{repository_id} [get]
func (s *service) GetImageRepository(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	repository, err := s.getImageRepository(ctx, org.ID, ctx.Param("repository_id"))
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, repository)
}

// @ID DeleteImageRepository
// @Summary remove an image repository assignment
// @Description Does not revoke cloud permissions or change existing image configuration connection pins.
// @Tags image-repositories
// @Security APIKey
// @Security OrgID
// @Param repository_id path string true "repository ID"
// @Success 204
// @Failure 401 {object} stderr.ErrResponse
// @Failure 403 {object} stderr.ErrResponse
// @Failure 404 {object} stderr.ErrResponse
// @Failure 500 {object} stderr.ErrResponse
// @Router /v1/image-repositories/{repository_id} [delete]
func (s *service) DeleteImageRepository(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	repository, err := s.getImageRepository(ctx, org.ID, ctx.Param("repository_id"))
	if err != nil {
		ctx.Error(err)
		return
	}
	if err := s.db.WithContext(ctx).Where(app.ImageRepository{OrgID: org.ID}).Delete(repository).Error; err != nil {
		ctx.Error(fmt.Errorf("unable to remove image repository assignment: %w", err))
		return
	}
	ctx.Status(http.StatusNoContent)
}

func (s *service) getImageRepository(ctx context.Context, orgID, id string) (*app.ImageRepository, error) {
	var repository app.ImageRepository
	if err := s.db.WithContext(ctx).Where(app.ImageRepository{OrgID: orgID, ID: id}).
		Preload("CloudConnection", app.CloudConnection{OrgID: orgID}).First(&repository).Error; err != nil {
		return nil, fmt.Errorf("image repository not found: %w", err)
	}
	return &repository, nil
}

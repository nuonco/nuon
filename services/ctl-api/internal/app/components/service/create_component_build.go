package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	buildsignal "github.com/nuonco/nuon/services/ctl-api/internal/app/components/signals/build"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/request"
	validatorPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/validator"
)

type CreateComponentBuildRequest struct {
	RequestID string  `json:"request_id,omitempty" validate:"omitempty,max=255"`
	GitRef    *string `validate:"required_unless=UseLatest true" json:"git_ref"`
	UseLatest bool    `validate:"required_without=GitRef" json:"use_latest"`
}

func (c *CreateComponentBuildRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return validatorPkg.FormatValidationError(err)
	}
	return nil
}

// @ID						CreateAppComponentBuild
// @Summary				create component build
// @Description.markdown	create_component_build.md
// @Param					app_id			path	string						true	"app ID"
// @Param					component_id	path	string						true	"component ID"
// @Param					req				body	CreateComponentBuildRequest	true	"Input"
// @Tags					components
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				409	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	app.ComponentBuild
// @Router					/v1/apps/{app_id}/components/{component_id}/builds [POST]
func (s *service) CreateAppComponentBuild(ctx *gin.Context) {
	appID := ctx.Param("app_id")
	cmpID := ctx.Param("component_id")
	cmp, err := s.getAppComponent(ctx, appID, cmpID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get app component: %w", err))
		return
	}

	var req CreateComponentBuildRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	s.createComponentBuildResponse(ctx, cmp, &req)
}

// @ID						CreateComponentBuild
// @Summary				create component build
// @Description.markdown	create_component_build.md
// @Param					component_id	path	string						true	"component ID"
// @Param					req				body	CreateComponentBuildRequest	true	"Input"
// @Tags					components
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Deprecated     true
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				409	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	app.ComponentBuild
// @Router					/v1/components/{component_id}/builds [POST]
func (s *service) CreateComponentBuild(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	cmpID := ctx.Param("component_id")

	cmp, err := s.findComponent(ctx, org.ID, cmpID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to find component %s: %w", cmpID, err))
		return
	}

	var req CreateComponentBuildRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	s.createComponentBuildResponse(ctx, cmp, &req)
}

func (s *service) createComponentBuildResponse(ctx *gin.Context, cmp *app.Component, req *CreateComponentBuildRequest) {
	if err := request.ValidateRequestID(req.RequestID); err != nil {
		ctx.Error(err)
		return
	}

	var bld *app.ComponentBuild
	var err error
	if req.RequestID != "" {
		bld, err = s.createIdempotentComponentBuild(ctx, cmp, req)
	} else {
		bld, err = s.helpers.CreateComponentBuild(ctx, cmp.ID, req.UseLatest, req.GitRef)
		if err == nil {
			err = s.enqueueComponentBuild(ctx, cmp.ID, bld, "")
		}
	}
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusCreated, bld)
}

func (s *service) createIdempotentComponentBuild(ctx context.Context, cmp *app.Component, req *CreateComponentBuildRequest) (*app.ComponentBuild, error) {
	hashReq := *req
	hashReq.RequestID = ""
	hash, err := request.Hash(struct {
		CreateComponentBuildRequest
		ComponentID string `json:"component_id"`
	}{
		CreateComponentBuildRequest: hashReq,
		ComponentID:                 cmp.ID,
	})
	if err != nil {
		return nil, err
	}

	existing, err := s.lookupComponentBuildByRequestID(ctx, cmp.OrgID, cmp.ID, req.RequestID)
	if err == nil {
		if err := request.Check(existing.Request, hash, ""); err != nil {
			return nil, err
		}
		if err := s.enqueueComponentBuild(ctx, cmp.ID, existing, req.RequestID); err != nil {
			return nil, err
		}
		return existing, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	wfRequest := &app.WorkflowRequest{
		RequestID:         req.RequestID,
		RequestHash:       hash,
		PinnedAppConfigID: "",
	}
	bld, err := s.helpers.CreateComponentBuildWithRequest(ctx, cmp.ID, req.UseLatest, req.GitRef, wfRequest)
	if err != nil {
		if !request.IsDuplicateKey(err) {
			return nil, err
		}
		existing, lookupErr := s.lookupComponentBuildByRequestID(ctx, cmp.OrgID, cmp.ID, req.RequestID)
		if lookupErr != nil {
			return nil, fmt.Errorf("unable to look up component build: %w", lookupErr)
		}
		if err := request.Check(existing.Request, hash, ""); err != nil {
			return nil, err
		}
		if err := s.enqueueComponentBuild(ctx, cmp.ID, existing, req.RequestID); err != nil {
			return nil, err
		}
		return existing, nil
	}
	if err := s.enqueueComponentBuild(ctx, cmp.ID, bld, req.RequestID); err != nil {
		return nil, err
	}
	return bld, nil
}

func (s *service) lookupComponentBuildByRequestID(ctx context.Context, orgID, componentID, requestID string) (*app.ComponentBuild, error) {
	var existing app.ComponentBuild
	err := s.db.WithContext(ctx).
		Where(app.ComponentBuild{OrgID: orgID}).
		Where("component_id = ? AND request->>'request_id' = ?", componentID, requestID).
		First(&existing).Error
	if err != nil {
		return nil, err
	}
	return &existing, nil
}

func (s *service) enqueueComponentBuild(ctx context.Context, componentID string, bld *app.ComponentBuild, requestID string) error {
	q, err := s.queueClient.GetDefaultQueueByOwner(ctx, componentID, "components")
	if err != nil {
		return fmt.Errorf("unable to get component queue: %w", err)
	}
	enq := &queueclient.EnqueueSignalRequest{
		QueueID:   q.ID,
		OwnerID:   bld.ID,
		OwnerType: "component_builds",
		Signal: &buildsignal.Signal{
			ComponentID: componentID,
			BuildID:     bld.ID,
		},
	}
	if requestID != "" {
		dedupe := request.DedupeKey("component-build", requestID)
		enq.DedupeKey = &dedupe
	}
	if _, err := s.queueClient.EnqueueSignal(ctx, enq); err != nil {
		return fmt.Errorf("unable to enqueue build signal: %w", err)
	}
	return nil
}

package service

import (
	"fmt"
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	createorg "github.com/nuonco/nuon/services/ctl-api/internal/app/onboarding/signals/create_org"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

type CompleteOrganizationStepRequest struct {
	Name  string `json:"name"`
	OrgID string `json:"org_id"`
}

// @ID						CompleteOnboardingOrganizationStep
// @Summary				Complete the organization step
// @Description			Creates a sandbox organization or attaches an existing one, then advances the onboarding to the app profile step
// @Tags					onboarding
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Param					req	body	CompleteOrganizationStepRequest	true	"Input"
// @Success				200	{object}	app.Onboarding
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Router					/v1/onboarding/current/steps/organization [POST]
func (s *service) CompleteOrganizationStep(ctx *gin.Context) {
	account, err := cctx.AccountFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	onboarding, err := s.getActiveOnboarding(ctx, account.ID)
	if err != nil {
		ctx.Error(err)
		return
	}

	if onboarding.CurrentStep != app.OnboardingStepOrganization {
		ctx.Error(fmt.Errorf("expected step %s but current step is %s", app.OnboardingStepOrganization, onboarding.CurrentStep))
		return
	}

	var req CompleteOrganizationStepRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	// Attach existing org (synchronous path)
	if req.OrgID != "" {
		s.attachExistingOrg(ctx, account, onboarding, req.OrgID)
		return
	}

	// Create new org (async path) — name is required
	if req.Name == "" {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("name is required when org_id is not provided"),
			Description: "either name or org_id must be provided",
		})
		return
	}

	s.createNewOrg(ctx, account, onboarding, req.Name)
}

// attachExistingOrg verifies the org exists and the user has access, then
// attaches it to the onboarding and advances to the next step synchronously.
func (s *service) attachExistingOrg(ctx *gin.Context, account *app.Account, onboarding *app.Onboarding, orgID string) {
	// Verify user has access to the org
	if !slices.Contains(account.OrgIDs, orgID) {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("account does not have access to org %s", orgID),
			Description: "you do not have access to this organization",
		})
		return
	}

	// Verify org exists and is active
	var org app.Org
	if err := s.db.WithContext(ctx).Where("id = ? AND status = ?", orgID, app.OrgStatusActive).First(&org).Error; err != nil {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("org not found or not active: %w", err),
			Description: "organization not found or not active",
		})
		return
	}

	// Update onboarding with org reference and advance step
	onboarding.OrgID = &orgID
	onboarding.CurrentStep = app.OnboardingStepYourStack
	onboarding.StepStatus = app.OnboardingStepStatusIdle

	if err := s.db.WithContext(ctx).Save(onboarding).Error; err != nil {
		ctx.Error(fmt.Errorf("unable to update onboarding: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, onboarding)
}

// createNewOrg enqueues an async signal to create a sandbox org and attach it.
func (s *service) createNewOrg(ctx *gin.Context, account *app.Account, onboarding *app.Onboarding, orgName string) {
	// Set processing status — signal will advance CurrentStep on completion
	onboarding.StepStatus = app.OnboardingStepStatusProcessing

	if err := s.db.WithContext(ctx).Save(onboarding).Error; err != nil {
		ctx.Error(fmt.Errorf("unable to update onboarding: %w", err))
		return
	}

	// Get queue and enqueue signal
	queue, err := s.queueClient.GetQueueByOwner(ctx, onboarding.ID, plugins.TableName(s.db, app.Onboarding{}))
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get onboarding queue: %w", err))
		return
	}

	_, err = s.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
		QueueID: queue.ID,
		Signal: &createorg.Signal{
			OnboardingID: onboarding.ID,
			AccountID:    account.ID,
			OrgName:      orgName,
		},
	})
	if err != nil {
		ctx.Error(fmt.Errorf("unable to enqueue create org signal: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, onboarding)
}

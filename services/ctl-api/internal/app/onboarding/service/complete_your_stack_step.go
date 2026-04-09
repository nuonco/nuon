package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	createapp "github.com/nuonco/nuon/services/ctl-api/internal/app/onboarding/signals/create_app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

type CompleteYourStackStepRequest struct {
	AppType        app.OnboardingAppType `json:"app_type" validate:"required,oneof=custom example" swaggertype:"string" enums:"custom,example"`
	ExampleAppSlug string                `json:"example_app_slug"`
	CloudProvider  string                `json:"cloud_provider"`
	AppAttributes  []string              `json:"app_attributes"`
}

// @ID						CompleteOnboardingYourStackStep
// @Summary				Complete the your stack step
// @Description			Configures the application profile and advances the onboarding to the install step
// @Tags					onboarding
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Param					req	body	CompleteYourStackStepRequest	true	"Input"
// @Success				200	{object}	app.Onboarding
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Router					/v1/onboarding/current/steps/your-stack [POST]
func (s *service) CompleteYourStackStep(ctx *gin.Context) {
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

	if onboarding.CurrentStep != app.OnboardingStepYourStack {
		ctx.Error(fmt.Errorf("expected step %s but current step is %s", app.OnboardingStepYourStack, onboarding.CurrentStep))
		return
	}

	var req CompleteYourStackStepRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	if err := s.v.Struct(&req); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	onboarding.AppType = req.AppType

	// Build the signal with example app fields if applicable
	sig := &createapp.Signal{
		OnboardingID: onboarding.ID,
	}

	switch req.AppType {
	case app.OnboardingAppTypeExample:
		if req.ExampleAppSlug == "" {
			ctx.Error(fmt.Errorf("example_app_slug is required for example app type"))
			return
		}
		exApp := getExampleAppBySlug(req.ExampleAppSlug)
		if exApp == nil {
			ctx.Error(fmt.Errorf("unknown example app slug: %s", req.ExampleAppSlug))
			return
		}
		onboarding.ExampleAppSlug = &req.ExampleAppSlug
		onboarding.CloudProvider = &exApp.CloudProvider
		sig.ExampleRepo = exApp.Repo
		sig.ExampleDirectory = exApp.Directory
		sig.ExampleBranch = exApp.Branch

	case app.OnboardingAppTypeCustom:
		if req.CloudProvider == "" {
			ctx.Error(fmt.Errorf("cloud_provider is required for custom app type"))
			return
		}
		onboarding.CloudProvider = &req.CloudProvider
		if len(req.AppAttributes) > 0 {
			onboarding.AppAttributes = pq.StringArray(req.AppAttributes)
		}
	}

	// Mark step as in-progress; the signal will advance to the next step
	// after the app is created and config is synced.
	onboarding.StepStatus = app.OnboardingStepStatusInProgress
	onboarding.SetCompositeStatus(ctx, app.StatusInProgress)

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
		Signal:  sig,
	})
	if err != nil {
		ctx.Error(fmt.Errorf("unable to enqueue create app signal: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, onboarding)
}

package service

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gorm.io/gorm"

	pkggenerics "github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/stackrun"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

type InstallPhoneHomeRequest map[string]any

type phoneHomeRequestType string

const (
	phoneHomeRequestTypeUpdate = "Update"
	phoneHomeRequestTypeDelete = "Delete"
	phoneHomeRequestTypeCreate = "Create"
)

// @ID PhoneHome
// @Summary				phone home for an install
// @Description.markdown phone_home.md
// @Param					install_id	path	string		true	"install ID"
// @Param					phone_home_id	path	string		true	"phone home ID"
// @Param					req		body	InstallPhoneHomeRequest	true	"Input"
// @Tags					installs
// @Accept					json
// @Produce				json
// @Security				APIKey && OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	app.EmptyResponse
// @Router					/v1/installs/{install_id}/phone-home/{phone_home_id} [post]
func (s *service) InstallPhoneHome(ctx *gin.Context) {
	installID := ctx.Param("install_id")
	phoneHomeID := ctx.Param("phone_home_id")

	var req InstallPhoneHomeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	var requestType string
	if v, ok := req["request_type"]; ok {
		requestType, ok = v.(string)
		if !ok {
			ctx.Error(stderr.NewInvalidRequest(fmt.Errorf("request type param must be a string")))
			return
		}
	} else {
		ctx.Error(stderr.NewInvalidRequest(fmt.Errorf("request type param not present")))
		return
	}

	switch requestType {
	case phoneHomeRequestTypeCreate, phoneHomeRequestTypeUpdate, phoneHomeRequestTypeDelete:
	default:
		ctx.Error(stderr.NewInvalidRequest(fmt.Errorf("invalid request type %q", requestType)))
		return
	}

	// Delete short-circuits ahead of every check, including auth. It carries nothing
	// the control plane trusts, and rejecting it would leave a deprovisioned install's
	// stack undeletable — its tokens are gone by then, so it cannot authenticate even
	// in principle.
	if requestType == phoneHomeRequestTypeDelete {
		ctx.JSON(http.StatusOK, app.EmptyResponse{})
		return
	}

	start := time.Now()
	metricTags := map[string]string{
		"request_type":   requestType,
		"cloud_platform": "aws",
		"status":         "error",
		"reason":         "",
		"org_id":         "",
	}
	defer func() {
		if s.mw != nil {
			s.mw.Timing("phone_home.auth.latency", time.Since(start), metrics.ToTags(metricTags))
		}
	}()

	// Hoisted out of updateInstallPhoneHome so authorization and the write share one
	// lookup, and so a missing version is a 401 rather than the 500 it used to be —
	// that 500 told an unauthenticated caller whether an install exists.
	stackVersion, install, err := s.getPhoneHomeTarget(ctx, installID, phoneHomeID)
	if err != nil {
		var authErr errPhoneHomeAuth
		if errors.As(err, &authErr) {
			metricTags["reason"] = authErr.reason
			s.l.Warn("rejected phone home",
				zap.String("install_id", installID),
				zap.String("phone_home_id", phoneHomeID),
				zap.String("reason", authErr.reason),
			)
		}
		ctx.Error(err)
		return
	}
	metricTags["org_id"] = install.OrgID
	metricTags["cloud_platform"] = string(install.CloudPlatform)

	reason, err := s.authorizePhoneHome(ctx, install, stackVersion, ctx.GetHeader("Authorization"))
	if err == nil && reason == phoneHomeAuthOK {
		reason, err = checkObservedCloudAccount(install, req)
	}
	metricTags["reason"] = reason
	if err != nil {
		var authErr errPhoneHomeAuth
		if errors.As(err, &authErr) {
			// Warn rather than Info: a rejection fails open on the customer's side, so
			// nothing else surfaces it. Never log the token.
			s.l.Warn("rejected phone home",
				zap.String("install_id", install.ID),
				zap.String("org_id", install.OrgID),
				zap.String("stack_version_id", stackVersion.ID),
				zap.String("request_type", requestType),
				zap.String("reason", authErr.reason),
			)
			s.recordPhoneHomeAuthResult(ctx, install, false)
		}
		ctx.Error(err)
		return
	}
	if reason == phoneHomeAuthOK {
		s.recordPhoneHomeAuthResult(ctx, install, true)
	}

	if err := s.updateInstallPhoneHome(ctx, stackVersion, requestType, &req); err != nil {
		ctx.Error(errors.Wrap(err, "unable to update install phone home"))
		return
	}

	metricTags["status"] = "ok"
	ctx.JSON(http.StatusCreated, app.EmptyResponse{})
}

// getPhoneHomeTarget resolves the stack version being called and its install. A miss on
// either is reported as an authentication failure so the endpoint reveals nothing about
// what does or does not exist.
func (s *service) getPhoneHomeTarget(
	ctx context.Context, installID, phoneHomeID string,
) (*app.InstallStackVersion, *app.Install, error) {
	var stackVersion app.InstallStackVersion
	if res := s.db.WithContext(ctx).
		Where(app.InstallStackVersion{
			InstallID:   installID,
			PhoneHomeID: phoneHomeID,
		}).
		First(&stackVersion); res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, nil, rejectPhoneHome(phoneHomeRejectUnknownToken,
				fmt.Errorf("no stack version for install %s", installID))
		}

		return nil, nil, errors.Wrap(res.Error, "unable to find stack")
	}

	// AppRunnerConfig is preloaded because Install.AfterQuery derives CloudPlatform from
	// it, and that is a metric tag on every rejection.
	var install app.Install
	if res := s.db.WithContext(ctx).
		Preload("AppRunnerConfig").
		Where(app.Install{ID: installID}).
		First(&install); res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, nil, rejectPhoneHome(phoneHomeRejectUnknownToken,
				fmt.Errorf("no install %s", installID))
		}

		return nil, nil, errors.Wrap(res.Error, "unable to find install")
	}

	return &stackVersion, &install, nil
}

func (s *service) updateInstallPhoneHome(ctx context.Context, stackVersion *app.InstallStackVersion, requestType string, req *InstallPhoneHomeRequest) error {
	installID := stackVersion.InstallID

	data, err := pkggenerics.ToMapstructureWithJSONTag(req)
	if err != nil {
		return errors.Wrap(err, "unable to convert to mapstructure")
	}

	updatedStack := app.InstallStackVersion{
		ID: stackVersion.ID,
	}
	res := s.db.WithContext(ctx).
		Model(&updatedStack).
		Updates(app.InstallStackVersion{
			Status: app.NewCompositeStatus(ctx, app.InstallStackVersionStatusActive),
			Runs: []app.InstallStackVersionRun{
				{
					Data: generics.ToHstore(pkggenerics.ToStringMap(pkggenerics.EncodeNestedForHstore(data))),
				},
			},
		})
	if res.Error != nil {
		return errors.Wrap(res.Error, "unable to update stack version")
	}

	run := app.InstallStackVersionRun{
		OrgID:                 stackVersion.OrgID,
		CreatedByID:           stackVersion.CreatedByID,
		InstallStackVersionID: stackVersion.ID,
		Data:                  generics.ToHstore(pkggenerics.ToStringMap(pkggenerics.EncodeNestedForHstore(data))),
	}
	if res = s.db.WithContext(ctx).
		Create(&run); res.Error != nil {
		return errors.Wrap(res.Error, "unable to create install stack version run")
	}

	ctx = cctx.SetOrgIDContext(ctx, stackVersion.OrgID)
	ctx = cctx.SetAccountIDContext(ctx, stackVersion.CreatedByID)
	queueID, err := s.getInstallSignalsQueueID(ctx, installID)
	if err != nil {
		return err
	}
	if err := s.enqueueInstallSignal(ctx, queueID, &stackrun.Signal{
		InstallStackID:        stackVersion.InstallStackID,
		InstallStackVersionID: stackVersion.ID,
		RunID:                 run.ID,
		RequestType:           requestType,
	}, "", ""); err != nil {
		return fmt.Errorf("enqueue signal: %w", err)
	}

	return nil
}

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
// @Deprecated				true
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	app.EmptyResponse
// @Router					/v1/installs/{install_id}/phone-home/{phone_home_id} [post]
func (s *service) InstallPhoneHome(ctx *gin.Context) {
	if s.cfg.PhoneHomeAuthRequired {
		s.installPhoneHomeRequired(ctx)
		return
	}

	s.installPhoneHomeLegacy(ctx)
}

func (s *service) installPhoneHomeRequired(ctx *gin.Context) {
	installID := ctx.Param("install_id")
	phoneHomeID := ctx.Param("phone_home_id")
	authHeader := ctx.GetHeader("Authorization")
	fields := []zap.Field{zap.String("install_id", installID), zap.String("phone_home_id", phoneHomeID)}

	metricTags, done := s.phoneHomeMetrics()
	defer done()

	if bearerToken(authHeader) == "" {
		s.failPhoneHome(ctx, metricTags, phoneHomeRejectMissingToken,
			rejectPhoneHome(phoneHomeRejectMissingToken, errors.New("no bearer token presented")), fields...)
		return
	}

	req, requestType, err := bindPhoneHomeRequest(ctx)
	if err != nil {
		s.failPhoneHome(ctx, metricTags, phoneHomeFailInvalidRequest, err, fields...)
		return
	}
	metricTags["request_type"] = requestType
	fields = append(fields, zap.String("request_type", requestType))

	stackVersion, install, err := s.getPhoneHomeTarget(ctx, installID, phoneHomeID)
	if err != nil {
		s.failPhoneHomeAuth(ctx, metricTags, nil, err, fields...)
		return
	}
	metricTags["org_id"] = install.OrgID
	metricTags["cloud_platform"] = string(install.CloudPlatform)
	fields = append(fields, zap.String("org_id", install.OrgID), zap.String("stack_version_id", stackVersion.ID))

	reason, err := s.verifyPhoneHomeCredential(ctx, install, stackVersion, authHeader)
	if err == nil && requestType != phoneHomeRequestTypeDelete {
		reason, err = checkObservedCloudAccount(install, req, true)
	}
	if err != nil {
		s.failPhoneHomeAuth(ctx, metricTags, install, err, fields...)
		return
	}
	metricTags["reason"] = reason
	s.recordPhoneHomeAuthResult(ctx, install, true)

	if requestType == phoneHomeRequestTypeDelete {
		metricTags["status"] = "ok"
		ctx.JSON(http.StatusOK, app.EmptyResponse{})
		return
	}

	if err := s.updateInstallPhoneHome(ctx, stackVersion, requestType, &req); err != nil {
		s.failPhoneHome(ctx, metricTags, phoneHomeFailWrite, errors.Wrap(err, "unable to update install phone home"), fields...)
		return
	}

	metricTags["status"] = "ok"
	ctx.JSON(http.StatusCreated, app.EmptyResponse{})
}

func (s *service) installPhoneHomeLegacy(ctx *gin.Context) {
	installID := ctx.Param("install_id")
	phoneHomeID := ctx.Param("phone_home_id")

	req, requestType, err := bindPhoneHomeRequest(ctx)
	if err != nil {
		ctx.Error(err)
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

	metricTags, done := s.phoneHomeMetrics()
	defer done()
	metricTags["request_type"] = requestType

	// Hoisted out of updateInstallPhoneHome so authorization and the write share one
	// lookup, and so a missing version is a 401 rather than the 500 it used to be —
	// that 500 told an unauthenticated caller whether an install exists.
	stackVersion, install, err := s.getPhoneHomeTarget(ctx, installID, phoneHomeID)
	if err != nil {
		s.rejectPhoneHomeTarget(ctx, metricTags, installID, phoneHomeID, err)
		return
	}
	metricTags["org_id"] = install.OrgID
	metricTags["cloud_platform"] = string(install.CloudPlatform)

	reason, err := s.authorizePhoneHome(ctx, install, stackVersion, ctx.GetHeader("Authorization"))
	if err == nil && reason == phoneHomeAuthOK {
		reason, err = checkObservedCloudAccount(install, req, false)
	}
	metricTags["reason"] = reason
	if err != nil {
		s.rejectPhoneHomeCaller(ctx, install, stackVersion, requestType, err)
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

func bindPhoneHomeRequest(ctx *gin.Context) (InstallPhoneHomeRequest, string, error) {
	var req InstallPhoneHomeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, "", stderr.NewInvalidRequest(err)
	}

	v, ok := req["request_type"]
	if !ok {
		return nil, "", stderr.NewInvalidRequest(fmt.Errorf("request type param not present"))
	}
	requestType, ok := v.(string)
	if !ok {
		return nil, "", stderr.NewInvalidRequest(fmt.Errorf("request type param must be a string"))
	}

	switch requestType {
	case phoneHomeRequestTypeCreate, phoneHomeRequestTypeUpdate, phoneHomeRequestTypeDelete:
	default:
		return nil, "", stderr.NewInvalidRequest(fmt.Errorf("invalid request type %q", requestType))
	}

	return req, requestType, nil
}

func (s *service) phoneHomeMetrics() (map[string]string, func()) {
	start := time.Now()
	metricTags := map[string]string{
		"request_type":   "",
		"cloud_platform": "aws",
		"status":         "error",
		"reason":         "",
		"org_id":         "",
	}

	return metricTags, func() {
		if s.mw != nil {
			s.mw.Timing("phone_home.auth.latency", time.Since(start), metrics.ToTags(metricTags))
		}
	}
}

func (s *service) rejectPhoneHomeTarget(
	ctx *gin.Context, metricTags map[string]string, installID, phoneHomeID string, err error,
) {
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
}

func (s *service) rejectPhoneHomeCaller(
	ctx *gin.Context, install *app.Install, stackVersion *app.InstallStackVersion, requestType string, err error,
) {
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
}

// failPhoneHomeAuth reports an error from target lookup or authorization. A rejection
// also stamps the install, when there is one; anything else is an internal failure.
func (s *service) failPhoneHomeAuth(
	ctx *gin.Context, metricTags map[string]string, install *app.Install, err error, fields ...zap.Field,
) {
	var authErr errPhoneHomeAuth
	if !errors.As(err, &authErr) {
		s.failPhoneHome(ctx, metricTags, phoneHomeFailInternal, err, fields...)
		return
	}

	if install != nil {
		s.recordPhoneHomeAuthResult(ctx, install, false)
	}
	s.failPhoneHome(ctx, metricTags, authErr.reason, err, fields...)
}

// failPhoneHome is the single exit for every failed phone home, so each one carries a
// reason tag and a log line. Warn rather than Info for rejections: a rejection fails
// open on the customer's side, so nothing else surfaces it. Never log the token.
func (s *service) failPhoneHome(
	ctx *gin.Context, metricTags map[string]string, reason string, err error, fields ...zap.Field,
) {
	metricTags["reason"] = reason
	fields = append(fields, zap.String("reason", reason), zap.Error(err))

	switch reason {
	case phoneHomeFailInternal, phoneHomeFailWrite:
		s.l.Error("phone home failed", fields...)
	default:
		s.l.Warn("rejected phone home", fields...)
	}
	ctx.Error(err)
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

	hstoreData := generics.ToHstore(pkggenerics.ToStringMap(pkggenerics.EncodeNestedForHstore(data)))
	run := app.InstallStackVersionRun{
		OrgID:                 stackVersion.OrgID,
		CreatedByID:           stackVersion.CreatedByID,
		InstallStackVersionID: stackVersion.ID,
		Data:                  hstoreData,
	}
	if res := s.db.WithContext(ctx).Create(&run); res.Error != nil {
		return errors.Wrap(res.Error, "unable to create install stack version run")
	}

	updatedStack := app.InstallStackVersion{
		ID: stackVersion.ID,
	}
	res := s.db.WithContext(ctx).
		Model(&updatedStack).
		Where("status->>'status' <> ?", app.StatusCancelled).
		Updates(app.InstallStackVersion{
			Status: app.NewCompositeStatus(ctx, app.InstallStackVersionStatusActive),
		})
	if res.Error != nil {
		return errors.Wrap(res.Error, "unable to update stack version")
	}

	cancelled := false
	if res.RowsAffected == 0 {
		var current app.InstallStackVersion
		if err := s.db.WithContext(ctx).First(&current, "id = ?", stackVersion.ID).Error; err != nil {
			return errors.Wrap(err, "unable to reload stack version")
		}
		if current.Status.Status == app.StatusCancelled {
			cancelled = true
			s.l.Warn("recorded phone-home run on cancelled stack version without reactivating",
				zap.String("install_stack_version_id", stackVersion.ID),
				zap.String("install_stack_version_run_id", run.ID),
				zap.String("install_id", installID),
			)
		}
	}

	if !cancelled {
		if err := s.helpers.RecordInstallStackVersionApplied(ctx, stackVersion, requestType); err != nil {
			return err
		}
	}

	ctx = cctx.SetOrgIDContext(ctx, stackVersion.OrgID)
	ctx = cctx.SetAccountIDContext(ctx, stackVersion.CreatedByID)
	queueID, err := s.getInstallSignalsQueueID(ctx, installID)
	if err != nil {
		return err
	}
	// Still enqueue for cancelled versions so stackrun can clear the await callback;
	// it will not reactivate or apply the cancelled version.
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

// @ID						PhoneHomeRunner
// @Summary				phone home for an install
// @Description			Deprecated. Use POST /v1/stacks/{install_id}/phone-home on the stack API. The caller authenticates as the stack service account, and the path no longer contains phone_home_id.
// @Param					install_id	path	string		true	"install ID"
// @Param					phone_home_id	path	string		true	"phone home ID"
// @Param					req		body	InstallPhoneHomeRequest	true	"Input"
// @Tags					installs/runner
// @Deprecated				true
// @Accept					json
// @Produce				json
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	app.EmptyResponse
// @Router					/v1/installs/{install_id}/phone-home/{phone_home_id} [post]
func (s *service) InstallPhoneHomeRunner(ctx *gin.Context) {
	s.InstallPhoneHome(ctx)
}

package service

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/pkg/shortid/domains"
	awstypes "github.com/nuonco/nuon/pkg/types/aws"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/request"
)

const (
	runnerAuthAWSSTSTypeSTS       = "sts"
	runnerAuthAWSSTSTypeSTSVendor = "sts-vendor"

	vendorRunnerRoleNamePrefix = "nuon-runner-"
)

type RunnerAuthAWSSTSRequest struct {
	Type       string                     `json:"type" validate:"required,oneof=sts sts-vendor" enums:"sts,sts-vendor"`
	RunnerID   string                     `json:"runner_id" validate:"required"`
	STSRequest *awstypes.PresignedRequest `json:"sts" validate:"required"`
}

type RunnerAuthAWSSTSResponse struct {
	Token     string    `json:"token"`
	RunnerID  string    `json:"runner_id"`
	ExpiresAt time.Time `json:"expires_at"`
}

// @ID						RunnerAuthAWSSTS
// @Summary				Authenticate a runner using a bound AWS STS request
// @Description			Validates runner identity by executing a presigned AWS STS GetCallerIdentity request that is bound to this API and to the runner. Type sts is signed by the runner's own IAM role; type sts-vendor is signed by the runner's Nuon-side identity role.
// @Param					req			body	RunnerAuthAWSSTSRequest	true	"Bound presigned AWS STS request"
// @Param					ttl_seconds	query	int						false	"Requested token lifetime in seconds, at most 691200 (8 days). Defaults to the maximum."
// @Tags					runners/auth
// @Accept					json
// @Produce				json
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	RunnerAuthAWSSTSResponse
// @Router					/v1/runner-auth/aws/sts [POST]
func (s *service) RunnerAuthAWSSTS(ctx *gin.Context) {
	start := time.Now()
	metricTags := map[string]string{
		"cloud_provider": "aws",
		"endpoint":       "aws-sts",
		"auth_method":    "unknown",
		"status":         "error",
	}
	defer func() {
		if s.mw != nil {
			tags := metrics.ToTags(metricTags)
			s.mw.Timing("runner.auth.latency", time.Since(start), tags)
			s.mw.Incr("runner.auth", tags)
		}
	}()

	var req RunnerAuthAWSSTSRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		metricTags["reason"] = "invalid_request"
		s.l.Warn("runner auth aws sts: failed to parse request", zap.Error(err))
		ctx.Error(stderr.NewInvalidRequest(errors.New("invalid request format")))
		ctx.Abort()
		return
	}
	if err := s.v.Struct(req); err != nil {
		metricTags["reason"] = "invalid_request"
		s.l.Warn("runner auth aws sts: request validation failed", zap.Error(err))
		ctx.Error(stderr.NewInvalidRequest(errors.New("invalid request: missing required fields")))
		ctx.Abort()
		return
	}
	metricTags["auth_method"] = req.Type

	tokenTTL, err := awsSTSRunnerTokenTTL(ctx.Query("ttl_seconds"))
	if err != nil {
		metricTags["reason"] = "invalid_request"
		s.l.Warn("runner auth aws sts: invalid ttl_seconds", zap.Error(err))
		ctx.Error(stderr.NewInvalidRequest(err))
		ctx.Abort()
		return
	}

	if !domains.IsRunnerID(req.RunnerID) {
		s.refuseAWSSTS(ctx, metricTags, "invalid_runner_id")
		return
	}

	bound, err := parseBoundSTSRequest(req.STSRequest, req.RunnerID, time.Now())
	if err != nil {
		s.refuseAWSSTS(ctx, metricTags, "invalid_bound_request", zap.Error(err))
		return
	}

	reqCtx := ctx.Request.Context()

	serverIDAllowed, err := s.boundServerIDAllowed(reqCtx, bound.serverID)
	if err != nil {
		s.failAWSSTS(ctx, metricTags, "server_id_lookup_error", err)
		return
	}
	if !serverIDAllowed {
		s.refuseAWSSTS(ctx, metricTags, "server_id_mismatch", zap.String("server_id", bound.serverID))
		return
	}

	stsResponse, err := s.executePresignedRequestHeaders(reqCtx, bound.request, presignedRequestTypeSTS, allowedBoundPresignedHeaders)
	if err != nil {
		s.refuseAWSSTS(ctx, metricTags, "sts_request_failed", zap.Error(err))
		return
	}

	var caller GetCallerIdentityResponse
	if err := xml.Unmarshal(stsResponse, &caller); err != nil {
		s.refuseAWSSTS(ctx, metricTags, "invalid_sts_response", zap.Error(err))
		return
	}

	runner, err := s.getRunnerWithGroupSettings(reqCtx, req.RunnerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.refuseAWSSTS(ctx, metricTags, "unknown_runner", zap.String("runner_id", req.RunnerID))
			return
		}
		s.failAWSSTS(ctx, metricTags, "runner_lookup_error", err)
		return
	}
	metricTags["runner_id"] = runner.ID
	metricTags["org_id"] = runner.OrgID

	logFields := []zap.Field{
		zap.String("runner_id", runner.ID),
		zap.String("org_id", runner.OrgID),
		zap.String("type", req.Type),
		zap.String("caller_account", caller.Result.Account),
		zap.String("caller_arn", caller.Result.Arn),
	}

	identity, err := s.getRunnerCloudIdentity(reqCtx, runner.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.refuseAWSSTS(ctx, metricTags, "no_binding", logFields...)
			return
		}
		s.failAWSSTS(ctx, metricTags, "binding_lookup_error", err)
		return
	}
	if identity.Provider != app.CloudPlatformAWS {
		s.refuseAWSSTS(ctx, metricTags, "binding_not_aws", logFields...)
		return
	}

	groupMethod := runner.RunnerGroup.Settings.AWSAuthMethod
	switch req.Type {
	case runnerAuthAWSSTSTypeSTSVendor:
		if groupMethod != app.RunnerAWSAuthMethodSTSVendor || !s.cfg.RunnerIdentityEnabled() {
			s.refuseAWSSTS(ctx, metricTags, "type_not_enabled", append(logFields, zap.String("aws_auth_method", string(groupMethod)))...)
			return
		}
	default:
		if groupMethod == app.RunnerAWSAuthMethodSTSVendor {
			s.refuseAWSSTS(ctx, metricTags, "type_not_enabled", append(logFields, zap.String("aws_auth_method", string(groupMethod)))...)
			return
		}
	}

	if bound.region != identity.Region {
		s.refuseAWSSTS(ctx, metricTags, "region_mismatch", append(logFields,
			zap.String("sts_region", bound.region),
			zap.String("expected_region", identity.Region))...)
		return
	}

	switch req.Type {
	case runnerAuthAWSSTSTypeSTSVendor:
		err = matchVendorRoleIdentity(identity, s.cfg.RunnerIdentityAccountID(), runner.ID, &caller)
	default:
		err = matchRunnerRoleIdentity(identity, &caller)
	}
	if err != nil {
		s.refuseAWSSTS(ctx, metricTags, "identity_mismatch", append(logFields, zap.Error(err))...)
		return
	}

	used := app.RunnerAuthUsedSignature{
		RunnerID:      runner.ID,
		SignatureHash: bound.signatureHash,
		ExpiresAt:     bound.expiresAt,
	}
	if err := s.db.WithContext(reqCtx).Create(&used).Error; err != nil {
		if request.IsDuplicateKey(err) {
			s.refuseAWSSTS(ctx, metricTags, "replay", logFields...)
			return
		}
		s.failAWSSTS(ctx, metricTags, "signature_record_error", err)
		return
	}

	token, err := s.createRunnerToken(reqCtx, runner.ID, tokenTTL)
	if err != nil {
		s.failAWSSTS(ctx, metricTags, "token_error", err)
		return
	}

	metricTags["status"] = "ok"
	s.l.Info("runner auth aws sts: authentication successful", append(logFields,
		zap.Time("expires_at", token.ExpiresAt))...)

	ctx.JSON(http.StatusOK, RunnerAuthAWSSTSResponse{
		Token:     token.Token,
		RunnerID:  runner.ID,
		ExpiresAt: token.ExpiresAt,
	})
}

func awsSTSRunnerTokenTTL(raw string) (time.Duration, error) {
	if raw == "" {
		return awsSTSRunnerTokenTimeout, nil
	}

	seconds, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || seconds <= 0 || seconds > int64(awsSTSRunnerTokenTimeout/time.Second) {
		return 0, fmt.Errorf("ttl_seconds must be between 1 and %d", int64(awsSTSRunnerTokenTimeout/time.Second))
	}

	return time.Duration(seconds) * time.Second, nil
}

func (s *service) refuseAWSSTS(ctx *gin.Context, metricTags map[string]string, reason string, fields ...zap.Field) {
	metricTags["reason"] = reason
	s.l.Warn("runner auth aws sts: refused", append(fields, zap.String("reason", reason))...)
	ctx.Error(stderr.ErrAuthentication{
		Err:         errors.New("authentication failed"),
		Description: "authentication failed",
	})
	ctx.Abort()
}

func (s *service) failAWSSTS(ctx *gin.Context, metricTags map[string]string, reason string, err error) {
	metricTags["reason"] = reason
	s.l.Error("runner auth aws sts: internal error", zap.String("reason", reason), zap.Error(err))
	ctx.Error(stderr.ErrSystem{
		Err:         errors.New("internal error"),
		Description: "internal error",
	})
	ctx.Abort()
}

func (s *service) getRunnerWithGroupSettings(ctx context.Context, runnerID string) (*app.Runner, error) {
	var runner app.Runner
	res := s.db.WithContext(ctx).
		Preload("RunnerGroup.Settings").
		First(&runner, "id = ?", runnerID)
	if res.Error != nil {
		return nil, res.Error
	}
	return &runner, nil
}

func (s *service) getRunnerCloudIdentity(ctx context.Context, runnerID string) (*app.RunnerCloudIdentity, error) {
	var identity app.RunnerCloudIdentity
	res := s.db.WithContext(ctx).
		Where(app.RunnerCloudIdentity{RunnerID: runnerID}).
		First(&identity)
	if res.Error != nil {
		return nil, res.Error
	}
	return &identity, nil
}

func matchRunnerRoleIdentity(identity *app.RunnerCloudIdentity, caller *GetCallerIdentityResponse) error {
	if identity.RunnerRoleARN == "" {
		return errors.New("binding has no runner role")
	}

	expected, err := arn.Parse(identity.RunnerRoleARN)
	if err != nil {
		return fmt.Errorf("binding runner role is not an ARN: %w", err)
	}
	expectedRoleName, err := extractRoleNameFromIAMRoleARN(identity.RunnerRoleARN)
	if err != nil {
		return err
	}
	if expected.AccountID != identity.AccountID {
		return fmt.Errorf("binding runner role is in account %s, binding account is %s", expected.AccountID, identity.AccountID)
	}

	got, err := arn.Parse(caller.Result.Arn)
	if err != nil {
		return fmt.Errorf("caller ARN is invalid: %w", err)
	}
	gotRoleName, err := extractRoleNameFromAssumedRoleARN(caller.Result.Arn)
	if err != nil {
		return err
	}

	if caller.Result.Account != identity.AccountID || got.AccountID != identity.AccountID {
		return fmt.Errorf("caller account %s does not match binding account %s", caller.Result.Account, identity.AccountID)
	}
	if got.Partition != expected.Partition {
		return fmt.Errorf("caller partition %s does not match binding partition %s", got.Partition, expected.Partition)
	}
	if gotRoleName != expectedRoleName {
		return fmt.Errorf("caller role %q does not match binding role %q", gotRoleName, expectedRoleName)
	}

	return nil
}

func matchVendorRoleIdentity(identity *app.RunnerCloudIdentity, identityAccountID, runnerID string, caller *GetCallerIdentityResponse) error {
	if identityAccountID == "" {
		return errors.New("runner identity account is not configured")
	}
	if identity.VendorRoleID == "" {
		return errors.New("binding has no vendor role")
	}

	got, err := arn.Parse(caller.Result.Arn)
	if err != nil {
		return fmt.Errorf("caller ARN is invalid: %w", err)
	}
	gotRoleName, err := extractRoleNameFromAssumedRoleARN(caller.Result.Arn)
	if err != nil {
		return err
	}

	if caller.Result.Account != identityAccountID || got.AccountID != identityAccountID {
		return fmt.Errorf("caller account %s is not the runner identity account", caller.Result.Account)
	}
	if expected := vendorRunnerRoleNamePrefix + runnerID; gotRoleName != expected {
		return fmt.Errorf("caller role %q does not match vendor role %q", gotRoleName, expected)
	}
	if roleID, _, _ := strings.Cut(caller.Result.UserId, ":"); roleID != identity.VendorRoleID {
		return fmt.Errorf("caller role id %q does not match the bound vendor role id", roleID)
	}

	return nil
}

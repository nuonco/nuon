package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"

	runneraws "github.com/nuonco/nuon/pkg/runner/auth/aws"
)

type RunnerAuthAWSIIDRequest struct {
	Document  string `json:"document" validate:"required"`
	Signature string `json:"signature" validate:"required"`
	RunnerID  string `json:"runner_id" validate:"required"`
}

type RunnerAuthAWSIIDResponse struct {
	Authenticated bool   `json:"authenticated"`
	AccountID     string `json:"account_id,omitempty"`
	InstanceID    string `json:"instance_id,omitempty"`
	Region        string `json:"region,omitempty"`
	RunnerID      string `json:"runner_id,omitempty"`
	Token         string `json:"token,omitempty"`
}

// @ID						RunnerAuthAWSIID
// @Summary				Authenticate a runner using AWS Instance Identity Document
// @Description			Validates runner identity by verifying an AWS-signed instance identity document
// @Param					req	body	RunnerAuthAWSIIDRequest	true	"IID auth request"
// @Tags					runners/auth
// @Accept					json
// @Produce				json
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	RunnerAuthAWSIIDResponse
// @Router					/v1/runner-auth/aws-iid [POST]
func (s *service) RunnerAuthAWSIID(ctx *gin.Context) {
	var req RunnerAuthAWSIIDRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		s.l.Warn("runner auth iid: failed to parse request", zap.Error(err))
		ctx.Error(stderr.NewInvalidRequest(errors.New("invalid request format")))
		ctx.Abort()
		return
	}

	if err := s.v.Struct(req); err != nil {
		s.l.Warn("runner auth iid: request validation failed", zap.Error(err))
		ctx.Error(stderr.NewInvalidRequest(errors.New("invalid request: missing required fields")))
		ctx.Abort()
		return
	}

	iid, err := runneraws.ParseAndValidateIID(req.Document)
	if err != nil {
		s.l.Warn("runner auth iid: document validation failed", zap.Error(err))
		ctx.Error(stderr.ErrAuthentication{
			Err:         errors.New("authentication failed"),
			Description: "invalid identity document",
		})
		ctx.Abort()
		return
	}

	if err := runneraws.VerifyIIDSignature(s.certStore, iid.Region, []byte(req.Document), req.Signature); err != nil {
		s.l.Warn("runner auth iid: signature verification failed",
			zap.String("region", iid.Region),
			zap.Error(err))
		ctx.Error(stderr.ErrAuthentication{
			Err:         errors.New("authentication failed"),
			Description: "identity document signature verification failed",
		})
		ctx.Abort()
		return
	}

	runner, err := s.getRunnerWithGroup(ctx.Request.Context(), req.RunnerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.l.Warn("runner auth iid: runner not found", zap.String("runner_id", req.RunnerID))
		} else {
			s.l.Error("runner auth iid: failed to get runner", zap.String("runner_id", req.RunnerID), zap.Error(err))
		}
		ctx.Error(stderr.ErrAuthentication{
			Err:         errors.New("authentication failed"),
			Description: "runner not recognized",
		})
		ctx.Abort()
		return
	}

	reqCtx := ctx.Request.Context()
	if err := s.validateRunnerAWSAccountID(reqCtx, runner, iid.AccountID); err != nil {
		s.l.Warn("runner auth iid: account validation failed",
			zap.String("runner_id", req.RunnerID),
			zap.String("iid_account_id", iid.AccountID),
			zap.Error(err))
		ctx.Error(stderr.ErrAuthorization{
			Err:         errors.New("authorization failed"),
			Description: "runner identity does not match expected configuration",
		})
		ctx.Abort()
		return
	}

	token, err := s.createRunnerToken(ctx.Request.Context(), runner.ID)
	if err != nil {
		s.l.Error("runner auth iid: failed to create token", zap.String("runner_id", req.RunnerID), zap.Error(err))
		ctx.Error(stderr.ErrSystem{
			Err:         errors.New("internal error"),
			Description: "failed to issue authentication token",
		})
		ctx.Abort()
		return
	}

	s.l.Info("runner auth iid: authentication successful",
		zap.String("runner_id", runner.ID),
		zap.String("instance_id", iid.InstanceID),
		zap.String("account_id", iid.AccountID),
		zap.String("region", iid.Region))

	ctx.JSON(http.StatusOK, RunnerAuthAWSIIDResponse{
		Authenticated: true,
		AccountID:     iid.AccountID,
		InstanceID:    iid.InstanceID,
		Region:        iid.Region,
		RunnerID:      runner.ID,
		Token:         token,
	})
}

// validateRunnerAWSAccountID validates the IID account ID against the
// install's stack outputs.
func (s *service) validateRunnerAWSAccountID(ctx context.Context, runner *app.Runner, iidAccountID string) error {
	install, err := s.getInstallByRunnerGroup(ctx, &runner.RunnerGroup)
	if err != nil {
		return fmt.Errorf("failed to get install for runner: %w", err)
	}

	installStack, err := s.getInstallStackWithOutputs(ctx, install.ID)
	if err != nil {
		return fmt.Errorf("failed to get install stack for install %s: %w", install.ID, err)
	}

	if installStack.InstallStackOutputs.AWSStackOutputs == nil {
		return fmt.Errorf("install %s does not have AWS stack outputs configured", install.ID)
	}

	expectedAccountID := installStack.InstallStackOutputs.AWSStackOutputs.AccountID
	if expectedAccountID == "" {
		return fmt.Errorf("install %s does not have an AWS account ID in stack outputs", install.ID)
	}

	if iidAccountID != expectedAccountID {
		return fmt.Errorf("AWS account ID mismatch: got %s, expected %s", iidAccountID, expectedAccountID)
	}

	return nil
}

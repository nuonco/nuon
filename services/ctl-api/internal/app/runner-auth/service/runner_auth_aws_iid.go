package service

import (
	"context"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	mozpkcs7 "go.mozilla.org/pkcs7"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

var awsAccountIDPattern = regexp.MustCompile(`^\d{12}$`)

// InstanceIdentityDocument represents the JSON document returned by
// the EC2 IMDS at /latest/dynamic/instance-identity/document.
type InstanceIdentityDocument struct {
	AccountID        string    `json:"accountId"`
	Architecture     string    `json:"architecture"`
	AvailabilityZone string    `json:"availabilityZone"`
	ImageID          string    `json:"imageId"`
	InstanceID       string    `json:"instanceId"`
	InstanceType     string    `json:"instanceType"`
	PendingTime      time.Time `json:"pendingTime"`
	PrivateIP        string    `json:"privateIp"`
	Region           string    `json:"region"`
	Version          string    `json:"version"`
}

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

	// Parse the IID JSON
	var iid InstanceIdentityDocument
	if err := json.Unmarshal([]byte(req.Document), &iid); err != nil {
		s.l.Warn("runner auth iid: failed to parse IID document", zap.Error(err))
		ctx.Error(stderr.ErrAuthentication{
			Err:         errors.New("authentication failed"),
			Description: "invalid identity document",
		})
		ctx.Abort()
		return
	}

	// Basic IID field validation
	if iid.Region == "" || iid.AccountID == "" || iid.InstanceID == "" {
		s.l.Warn("runner auth iid: IID missing required fields")
		ctx.Error(stderr.ErrAuthentication{
			Err:         errors.New("authentication failed"),
			Description: "identity document missing required fields",
		})
		ctx.Abort()
		return
	}

	if !awsAccountIDPattern.MatchString(iid.AccountID) {
		s.l.Warn("runner auth iid: invalid account ID format",
			zap.String("account_id", iid.AccountID))
		ctx.Error(stderr.ErrAuthentication{
			Err:         errors.New("authentication failed"),
			Description: "invalid account ID format",
		})
		ctx.Abort()
		return
	}

	// Verify the RSA-2048 PKCS7 signature
	if err := s.verifyIIDSignature(iid.Region, []byte(req.Document), req.Signature); err != nil {
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

	// Look up the runner
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

	// Validate account ID against install stack outputs
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

	// Create token
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

// verifyIIDSignature verifies the PKCS7 signature of an instance
// identity document using the AWS public certificate for the given region.
// The signature from IMDS /instance-identity/rsa2048 is a PKCS7/SMIME
// signed message.
func (s *service) verifyIIDSignature(region string, document []byte, signatureB64 string) error {
	cert, err := s.certStore.GetCert(region)
	if err != nil {
		return fmt.Errorf("no certificate for region %s: %w", region, err)
	}

	sigDER, err := base64.StdEncoding.DecodeString(signatureB64)
	if err != nil {
		return fmt.Errorf("failed to decode signature: %w", err)
	}

	p7, err := mozpkcs7.Parse(sigDER)
	if err != nil {
		return fmt.Errorf("failed to parse PKCS7 signature: %w", err)
	}

	// Use the AWS cert as the trust anchor
	p7.Certificates = []*x509.Certificate{cert}

	if err := p7.Verify(); err != nil {
		return fmt.Errorf("PKCS7 signature verification failed: %w", err)
	}

	return nil
}

// validateRunnerAWSAccountID validates the IID account ID against the
// install's stack outputs. This is the same check as validateRunnerAWSIdentity
// but without the IAM role ARN check (IID doesn't provide that).
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

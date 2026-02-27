package service

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/identity"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

// RunnerAuthOIDCRequest represents a request to authenticate a runner using OIDC
type RunnerAuthOIDCRequest struct {
	// Provider indicates which cloud provider: "aws", "gcp", "azure", or "local"
	Provider string `json:"provider" binding:"required,oneof=aws gcp azure local"`

	// Token is the raw OIDC JWT token from the cloud provider's metadata service
	Token string `json:"token" binding:"required"`

	// RunnerID is the unique identifier for the runner
	RunnerID string `json:"runner_id" binding:"required"`
}

// RunnerAuthOIDCResponse represents the response from an OIDC authentication request
type RunnerAuthOIDCResponse struct {
	// Authenticated indicates whether the runner was successfully authenticated
	Authenticated bool `json:"authenticated"`

	// Provider is the cloud provider that was used for authentication
	Provider string `json:"provider"`

	// RunnerID is the unique identifier for the runner
	RunnerID string `json:"runner_id"`

	// Token is the runner API token (90-day service account token)
	Token string `json:"token"`
}

// @ID          RunnerAuthOIDC
// @Summary     Authenticate a runner using OIDC identity tokens
// @Description Validates runner identity using cloud provider OIDC tokens (AWS, GCP, or Azure)
// @Param       req body RunnerAuthOIDCRequest true "OIDC authentication request"
// @Tags        runners/auth
// @Accept      json
// @Produce     json
// @Failure     400 {object} stderr.ErrResponse
// @Failure     401 {object} stderr.ErrResponse
// @Failure     500 {object} stderr.ErrResponse
// @Success     200 {object} RunnerAuthOIDCResponse
// @Router      /v1/runner-auth/oidc [POST]
func (s *service) RunnerAuthOIDC(ctx *gin.Context) {
	var req RunnerAuthOIDCRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		s.l.Warn("runner auth oidc: failed to parse request", zap.Error(err))
		ctx.Error(stderr.NewInvalidRequest(errors.New("invalid request format")))
		ctx.Abort()
		return
	}

	reqCtx := ctx.Request.Context()

	// Get validator for the specified provider
	validator, err := s.getOIDCValidator(req.Provider)
	if err != nil {
		s.l.Error("runner auth oidc: failed to get validator",
			zap.String("provider", req.Provider),
			zap.Error(err))
		ctx.Error(stderr.ErrSystem{
			Err:         err,
			Description: "authentication system error",
		})
		ctx.Abort()
		return
	}

	// Validate token and extract claims
	claims, err := validator.ValidateToken(reqCtx, req.Token)
	if err != nil {
		s.l.Warn("runner auth oidc: token validation failed",
			zap.String("provider", req.Provider),
			zap.String("runner_id", req.RunnerID),
			zap.Error(err))
		ctx.Error(stderr.ErrAuthentication{
			Err:         errors.New("authentication failed"),
			Description: "invalid or expired identity token",
		})
		ctx.Abort()
		return
	}

	// Verify runner exists and get runner with group preloaded
	runner, err := s.getRunnerWithGroup(reqCtx, req.RunnerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.l.Warn("runner auth oidc: runner not found", zap.String("runner_id", req.RunnerID))
		} else {
			s.l.Error("runner auth oidc: failed to get runner",
				zap.String("runner_id", req.RunnerID),
				zap.Error(err))
		}
		ctx.Error(stderr.ErrAuthentication{
			Err:         errors.New("authentication failed"),
			Description: "runner not recognized",
		})
		ctx.Abort()
		return
	}

	// Validate cloud identity against install configuration
	if err := validator.ValidateRunnerIdentity(reqCtx, runner, claims); err != nil {
		s.l.Warn("runner auth oidc: identity validation failed",
			zap.String("runner_id", req.RunnerID),
			zap.String("provider", req.Provider),
			zap.Error(err))
		ctx.Error(stderr.ErrAuthorization{
			Err:         errors.New("authorization failed"),
			Description: "runner identity does not match expected configuration",
		})
		ctx.Abort()
		return
	}

	// Create runner token
	token, err := s.createRunnerToken(reqCtx, runner.ID)
	if err != nil {
		s.l.Error("runner auth oidc: failed to create token",
			zap.String("runner_id", req.RunnerID),
			zap.Error(err))
		ctx.Error(stderr.ErrSystem{
			Err:         errors.New("internal error"),
			Description: "failed to issue authentication token",
		})
		ctx.Abort()
		return
	}

	s.l.Info("runner auth oidc: authentication successful",
		zap.String("runner_id", runner.ID),
		zap.String("provider", req.Provider),
		zap.String("issuer", claims.Issuer))

	ctx.JSON(http.StatusOK, RunnerAuthOIDCResponse{
		Authenticated: true,
		Provider:      req.Provider,
		RunnerID:      runner.ID,
		Token:         token,
	})
}

// getOIDCValidator returns the appropriate identity validator for the given provider
func (s *service) getOIDCValidator(provider string) (*IdentityValidator, error) {
	switch identity.ProviderType(provider) {
	case identity.ProviderTypeAWS:
		return s.awsValidator, nil
	case identity.ProviderTypeGCP:
		return s.gcpValidator, nil
	case identity.ProviderTypeAzure:
		return s.azureValidator, nil
	case identity.ProviderTypeLocal:
		return s.localValidator, nil
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}

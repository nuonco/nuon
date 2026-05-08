package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/api/idtoken"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/metrics"
	gcptypes "github.com/nuonco/nuon/pkg/types/gcp"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

const (
	gcpRunnerIDMetadataKey = "nuon_runner_id"
)

var (
	// Only allow requests to the GCP Compute API for reading instances.
	gcpComputeURLPattern = regexp.MustCompile(`^https://compute\.googleapis\.com/compute/v1/projects/[a-z][a-z0-9-]*/zones/[a-z0-9-]+/instances/[a-zA-Z0-9][a-zA-Z0-9._-]*$`)

	allowedGCPHeaders = map[string]struct{}{
		"authorization": {},
	}
)

type RunnerAuthGCPRequest struct {
	IdentityToken   string                    `json:"identity_token" validate:"required"`
	MetadataRequest *gcptypes.MetadataRequest `json:"metadata" validate:"required"`
}

type RunnerAuthGCPResponse struct {
	Authenticated  bool   `json:"authenticated"`
	ProjectID      string `json:"project_id,omitempty"`
	InstanceID     string `json:"instance_id,omitempty"`
	ServiceAccount string `json:"service_account,omitempty"`
	RunnerID       string `json:"runner_id,omitempty"`
	Token          string `json:"token,omitempty"`
}

// @ID						RunnerAuthGCP
// @Summary				Authenticate a runner using a GCP identity token
// @Description			Validates runner identity by verifying a GCP identity token and independently reading instance metadata
// @Param					req	body	RunnerAuthGCPRequest	true	"GCP identity token and metadata request"
// @Tags					runners/auth
// @Accept					json
// @Produce				json
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	RunnerAuthGCPResponse
// @Router					/v1/runner-auth/gcp [POST]
func (s *service) RunnerAuthGCP(ctx *gin.Context) {
	start := time.Now()
	metricTags := map[string]string{
		"cloud_provider": "gcp",
		"auth_method":    "jwt",
		"status":         "error",
	}
	defer func() {
		if s.mw != nil {
			s.mw.Timing("runner.auth.latency", time.Since(start), metrics.ToTags(metricTags))
		}
	}()

	var req RunnerAuthGCPRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		s.l.Warn("runner auth gcp: failed to parse request", zap.Error(err))
		ctx.Error(stderr.NewInvalidRequest(errors.New("invalid request format")))
		ctx.Abort()
		return
	}

	if err := s.v.Struct(req); err != nil {
		s.l.Warn("runner auth gcp: request validation failed", zap.Error(err))
		ctx.Error(stderr.NewInvalidRequest(errors.New("invalid request: missing required fields")))
		ctx.Abort()
		return
	}

	reqCtx := ctx.Request.Context()

	// Step 1: Verify identity token (JWT) — proves project, SA, instance ID
	payload, err := idtoken.Validate(reqCtx, req.IdentityToken, s.cfg.RunnerAPIURL)
	if err != nil {
		s.l.Warn("runner auth gcp: identity token validation failed", zap.Error(err))
		ctx.Error(stderr.ErrAuthentication{
			Err:         errors.New("authentication failed"),
			Description: "invalid GCP identity token",
		})
		ctx.Abort()
		return
	}

	claims, err := extractGCPClaims(payload)
	if err != nil {
		s.l.Warn("runner auth gcp: failed to extract claims", zap.Error(err))
		ctx.Error(stderr.ErrAuthentication{
			Err:         errors.New("authentication failed"),
			Description: "invalid identity token claims",
		})
		ctx.Abort()
		return
	}

	// Step 2: Execute metadata request — independently reads instance metadata (runner ID)
	// Mirrors the AWS presigned DescribeTags pattern.
	instanceData, err := s.executeGCPMetadataRequest(reqCtx, req.MetadataRequest, claims)
	if err != nil {
		s.l.Warn("runner auth gcp: metadata request failed", zap.Error(err))
		ctx.Error(stderr.ErrAuthentication{
			Err:         errors.New("authentication failed"),
			Description: "failed to verify instance metadata",
		})
		ctx.Abort()
		return
	}

	// Step 3: Cross-validate instance ID from JWT matches Compute API response
	apiInstanceID := instanceData.ID
	if apiInstanceID != claims.instanceID {
		s.l.Warn("runner auth gcp: instance ID mismatch",
			zap.String("jwt_instance_id", claims.instanceID),
			zap.String("api_instance_id", apiInstanceID))
		ctx.Error(stderr.ErrAuthentication{
			Err:         errors.New("authentication failed"),
			Description: "instance identity mismatch",
		})
		ctx.Abort()
		return
	}

	// Step 4: Extract runner ID from instance metadata (independently verified)
	runnerID := extractRunnerIDFromMetadata(instanceData)
	if runnerID == "" {
		s.l.Warn("runner auth gcp: missing runner ID in instance metadata",
			zap.String("instance_id", claims.instanceID))
		ctx.Error(stderr.ErrAuthentication{
			Err:         errors.New("authentication failed"),
			Description: "instance is not a registered runner",
		})
		ctx.Abort()
		return
	}

	runner, err := s.getRunnerWithGroup(reqCtx, runnerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.l.Warn("runner auth gcp: runner not found", zap.String("runner_id", runnerID))
		} else {
			s.l.Error("runner auth gcp: failed to get runner", zap.String("runner_id", runnerID), zap.Error(err))
		}
		ctx.Error(stderr.ErrAuthentication{
			Err:         errors.New("authentication failed"),
			Description: "runner not recognized",
		})
		ctx.Abort()
		return
	}
	metricTags["runner_id"] = runner.ID

	install, err := s.getInstallByRunnerGroup(reqCtx, &runner.RunnerGroup)
	if err != nil {
		s.l.Warn("runner auth gcp: failed to get install for runner",
			zap.String("runner_id", runnerID),
			zap.Error(err))
		ctx.Error(stderr.ErrAuthentication{
			Err:         errors.New("authentication failed"),
			Description: "runner not associated with an install",
		})
		ctx.Abort()
		return
	}
	metricTags["install_id"] = install.ID
	metricTags["install_name"] = install.Name
	metricTags["org_id"] = install.OrgID

	if err := s.validateRunnerGCPIdentity(reqCtx, install, claims); err != nil {
		s.l.Warn("runner auth gcp: identity validation failed",
			zap.String("runner_id", runnerID),
			zap.String("install_id", install.ID),
			zap.String("install_name", install.Name),
			zap.String("org_id", install.OrgID),
			zap.String("project_id", claims.projectID),
			zap.String("service_account", claims.serviceAccount),
			zap.Error(err))
		ctx.Error(stderr.ErrAuthorization{
			Err:         errors.New("authorization failed"),
			Description: "runner identity does not match expected configuration",
		})
		ctx.Abort()
		return
	}

	token, err := s.createRunnerToken(reqCtx, runner.ID)
	if err != nil {
		s.l.Error("runner auth gcp: failed to create token", zap.String("runner_id", runnerID), zap.Error(err))
		ctx.Error(stderr.ErrSystem{
			Err:         errors.New("internal error"),
			Description: "failed to issue authentication token",
		})
		ctx.Abort()
		return
	}

	metricTags["status"] = "ok"
	s.l.Info("runner auth gcp: authentication successful",
		zap.String("install_id", install.ID),
		zap.String("install_name", install.Name),
		zap.String("org_id", install.OrgID),
		zap.String("runner_id", runner.ID),
		zap.String("instance_id", claims.instanceID),
		zap.String("project_id", claims.projectID))

	ctx.JSON(http.StatusOK, RunnerAuthGCPResponse{
		Authenticated:  true,
		ProjectID:      claims.projectID,
		InstanceID:     claims.instanceID,
		ServiceAccount: claims.serviceAccount,
		RunnerID:       runner.ID,
		Token:          token,
	})
}

type gcpClaims struct {
	projectID      string
	instanceID     string
	zone           string
	serviceAccount string
}

func extractGCPClaims(payload *idtoken.Payload) (*gcpClaims, error) {
	google, ok := payload.Claims["google"].(map[string]interface{})
	if !ok {
		return nil, errors.New("missing google claim")
	}

	computeEngine, ok := google["compute_engine"].(map[string]interface{})
	if !ok {
		return nil, errors.New("missing google.compute_engine claim")
	}

	projectID, _ := computeEngine["project_id"].(string)
	if projectID == "" {
		return nil, errors.New("missing project_id in compute_engine claims")
	}

	instanceID, _ := computeEngine["instance_id"].(string)
	if instanceID == "" {
		return nil, errors.New("missing instance_id in compute_engine claims")
	}

	zone, _ := computeEngine["zone"].(string)

	sa, _ := payload.Claims["email"].(string)
	if sa == "" {
		return nil, errors.New("missing email claim (service account)")
	}

	return &gcpClaims{
		projectID:      projectID,
		instanceID:     instanceID,
		zone:           zone,
		serviceAccount: sa,
	}, nil
}

// gcpInstanceResponse is a subset of the Compute API instances.get response.
type gcpInstanceResponse struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Metadata struct {
		Items []struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		} `json:"items"`
	} `json:"metadata"`
}

func validateGCPMetadataRequest(req *gcptypes.MetadataRequest) error {
	if req.Method != http.MethodGet {
		return errors.New("only GET method is allowed")
	}

	u, err := url.Parse(req.URL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	if u.Scheme != "https" {
		return errors.New("only HTTPS scheme is allowed")
	}

	if !gcpComputeURLPattern.MatchString(req.URL) {
		return fmt.Errorf("URL does not match expected GCP Compute API pattern: %s", req.URL)
	}

	for key := range req.Headers {
		if _, ok := allowedGCPHeaders[strings.ToLower(key)]; !ok {
			return fmt.Errorf("header not allowed: %s", key)
		}
	}

	return nil
}

func (s *service) executeGCPMetadataRequest(ctx context.Context, metadataReq *gcptypes.MetadataRequest, claims *gcpClaims) (*gcpInstanceResponse, error) {
	if err := validateGCPMetadataRequest(metadataReq); err != nil {
		return nil, fmt.Errorf("metadata request validation failed: %w", err)
	}

	// Verify the URL targets the same project/zone/instance from the JWT claims
	expectedPrefix := fmt.Sprintf("https://compute.googleapis.com/compute/v1/projects/%s/zones/%s/instances/", claims.projectID, claims.zone)
	u, _ := url.Parse(metadataReq.URL)
	fullURL := u.Scheme + "://" + u.Host + u.Path
	if len(fullURL) < len(expectedPrefix) || fullURL[:len(expectedPrefix)] != expectedPrefix {
		return nil, fmt.Errorf("metadata URL project/zone does not match identity token claims")
	}

	req, err := http.NewRequestWithContext(ctx, metadataReq.Method, metadataReq.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	for key, value := range metadataReq.Headers {
		req.Header.Set(key, value)
	}

	resp, err := presignedHTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxPresignedResponseSize))
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GCP Compute API request failed with status %d", resp.StatusCode)
	}

	var instanceResp gcpInstanceResponse
	if err := json.Unmarshal(body, &instanceResp); err != nil {
		return nil, fmt.Errorf("failed to parse instance response: %w", err)
	}

	return &instanceResp, nil
}

func extractRunnerIDFromMetadata(instance *gcpInstanceResponse) string {
	for _, item := range instance.Metadata.Items {
		if item.Key == gcpRunnerIDMetadataKey {
			return item.Value
		}
	}
	return ""
}

func (s *service) validateRunnerGCPIdentity(ctx context.Context, install *app.Install, claims *gcpClaims) error {
	installStack, err := s.getInstallStackWithOutputs(ctx, install.ID)
	if err != nil {
		return fmt.Errorf("failed to get install stack for install %s: %w", install.ID, err)
	}

	if installStack.InstallStackOutputs.GCPStackOutputs == nil {
		return fmt.Errorf("install %s does not have GCP stack outputs configured", install.ID)
	}

	gcpOutputs := installStack.InstallStackOutputs.GCPStackOutputs

	expectedProjectID := gcpOutputs.ProjectID
	if expectedProjectID == "" {
		return fmt.Errorf("install %s does not have a GCP project ID in stack outputs", install.ID)
	}

	if claims.projectID != expectedProjectID {
		return fmt.Errorf("GCP project ID mismatch: got %s, expected %s", claims.projectID, expectedProjectID)
	}

	expectedSA := gcpOutputs.RunnerServiceAccountEmail
	if expectedSA == "" {
		return fmt.Errorf("install %s does not have a runner service account email in stack outputs", install.ID)
	}

	if claims.serviceAccount != expectedSA {
		return fmt.Errorf("GCP service account mismatch: got %s, expected %s", claims.serviceAccount, expectedSA)
	}

	return nil
}

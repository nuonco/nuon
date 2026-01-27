package service

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	awstypes "github.com/nuonco/nuon/pkg/types/aws"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
)

const (
	// runnerIDTagKey is the EC2 tag key that contains the Nuon runner ID
	runnerIDTagKey = "runner.nuon.co/id"
	// defaultRunnerTokenTimeout is the default expiration time for runner tokens
	defaultRunnerTokenTimeout = time.Hour * 24 * 90
)

// RunnerAuthAWSRequest contains the presigned AWS requests from a runner
type RunnerAuthAWSRequest struct {
	// STSRequest is a presigned GetCallerIdentity request
	STSRequest *awstypes.PresignedRequest `json:"sts" validate:"required"`
	// TagsRequest is a presigned EC2 DescribeTags request for the instance
	TagsRequest *awstypes.PresignedRequest `json:"tags" validate:"required"`
}

// RunnerAuthAWSResponse contains the authentication result
type RunnerAuthAWSResponse struct {
	// Authenticated indicates whether the runner was successfully authenticated
	Authenticated bool `json:"authenticated"`
	// AccountID is the AWS account ID from the STS response
	AccountID string `json:"account_id,omitempty"`
	// ARN is the IAM role/user ARN from the STS response
	ARN string `json:"arn,omitempty"`
	// InstanceID is the EC2 instance ID from tags
	InstanceID string `json:"instance_id,omitempty"`
	// RunnerID is the Nuon runner ID from the instance tags
	RunnerID string `json:"runner_id,omitempty"`
	// Token is the authentication token issued to the runner
	Token string `json:"token,omitempty"`
}

// GetCallerIdentityResponse represents the AWS STS GetCallerIdentity response
type GetCallerIdentityResponse struct {
	XMLName xml.Name `xml:"GetCallerIdentityResponse"`
	Result  struct {
		Arn     string `xml:"Arn"`
		UserId  string `xml:"UserId"`
		Account string `xml:"Account"`
	} `xml:"GetCallerIdentityResult"`
}

// DescribeTagsResponse represents the AWS EC2 DescribeTags response
type DescribeTagsResponse struct {
	XMLName xml.Name `xml:"DescribeTagsResponse"`
	TagSet  struct {
		Items []struct {
			Key          string `xml:"key"`
			Value        string `xml:"value"`
			ResourceId   string `xml:"resourceId"`
			ResourceType string `xml:"resourceType"`
		} `xml:"item"`
	} `xml:"tagSet"`
}

// @ID						RunnerAuthAWS
// @Summary				Authenticate a runner using AWS presigned requests
// @Description			Validates runner identity by executing presigned AWS STS and EC2 requests
// @Param					req	body	RunnerAuthAWSRequest	true	"Presigned AWS requests"
// @Tags					runners/auth
// @Accept					json
// @Produce				json
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	RunnerAuthAWSResponse
// @Router					/v1/runner-auth/aws [POST]
func (s *service) RunnerAuthAWS(ctx *gin.Context) {
	var req RunnerAuthAWSRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	if err := s.v.Struct(req); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	// Execute the presigned STS GetCallerIdentity request
	stsResponse, err := s.executePresignedRequest(ctx, req.STSRequest)
	if err != nil {
		ctx.Error(fmt.Errorf("failed to execute STS request: %w", err))
		return
	}

	// Parse the STS response
	var callerIdentity GetCallerIdentityResponse
	if err := xml.Unmarshal(stsResponse, &callerIdentity); err != nil {
		ctx.Error(fmt.Errorf("failed to parse STS response: %w", err))
		return
	}

	// Execute the presigned EC2 DescribeTags request
	tagsResponse, err := s.executePresignedRequest(ctx, req.TagsRequest)
	if err != nil {
		ctx.Error(fmt.Errorf("failed to execute EC2 tags request: %w", err))
		return
	}

	// Parse the EC2 tags response
	var describeTags DescribeTagsResponse
	if err := xml.Unmarshal(tagsResponse, &describeTags); err != nil {
		ctx.Error(fmt.Errorf("failed to parse EC2 tags response: %w", err))
		return
	}

	// Extract tags into a map
	tags := make(map[string]string)
	var instanceID string
	for _, tag := range describeTags.TagSet.Items {
		tags[tag.Key] = tag.Value
		if instanceID == "" {
			instanceID = tag.ResourceId
		}
	}

	// Get the runner ID from the instance tags
	runnerID, ok := tags[runnerIDTagKey]
	if !ok || runnerID == "" {
		ctx.Error(fmt.Errorf("missing required tag %s on instance %s", runnerIDTagKey, instanceID))
		return
	}

	// Look up the runner with its runner group in the database
	runner, err := s.getRunnerWithGroup(ctx, runnerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.Error(fmt.Errorf("runner not found: %s", runnerID))
			return
		}
		ctx.Error(fmt.Errorf("failed to get runner: %w", err))
		return
	}

	// Validate the caller's AWS account against the install's expected account
	if err := s.validateRunnerAWSIdentity(ctx, runner, callerIdentity.Result.Account, callerIdentity.Result.Arn); err != nil {
		ctx.Error(fmt.Errorf("AWS identity validation failed: %w", err))
		return
	}

	// Generate authentication token for the runner
	token, err := s.createRunnerToken(ctx, runner.ID)
	if err != nil {
		ctx.Error(fmt.Errorf("failed to create runner token: %w", err))
		return
	}

	response := RunnerAuthAWSResponse{
		Authenticated: true,
		AccountID:     callerIdentity.Result.Account,
		ARN:           callerIdentity.Result.Arn,
		InstanceID:    instanceID,
		RunnerID:      runner.ID,
		Token:         token,
	}

	ctx.JSON(http.StatusOK, response)
}

// executePresignedRequest executes a presigned AWS request and returns the response body
func (s *service) executePresignedRequest(ctx context.Context, presignedReq *awstypes.PresignedRequest) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, presignedReq.Method, presignedReq.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set the presigned headers
	for key, value := range presignedReq.Headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("AWS request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// getRunnerWithGroup retrieves a runner by its ID with the runner group preloaded
func (s *service) getRunnerWithGroup(ctx context.Context, runnerID string) (*app.Runner, error) {
	var runner app.Runner
	res := s.db.WithContext(ctx).
		Preload("RunnerGroup").
		First(&runner, "id = ?", runnerID)
	if res.Error != nil {
		return nil, res.Error
	}
	return &runner, nil
}

// getInstallByRunnerGroup retrieves an install by its runner group
func (s *service) getInstallByRunnerGroup(ctx context.Context, runnerGroup *app.RunnerGroup) (*app.Install, error) {
	if runnerGroup.OwnerType != "installs" {
		return nil, fmt.Errorf("runner group is not associated with an install")
	}

	var install app.Install
	res := s.db.WithContext(ctx).
		Preload("AWSAccount").
		First(&install, "id = ?", runnerGroup.OwnerID)
	if res.Error != nil {
		return nil, res.Error
	}
	return &install, nil
}

// TODO(fd): remove the caller ARN
// validateRunnerAWSIdentity validates the caller's AWS identity against the expected install configuration
func (s *service) validateRunnerAWSIdentity(ctx context.Context, runner *app.Runner, callerAccountID, callerARN string) error {
	// Get the install associated with this runner
	install, err := s.getInstallByRunnerGroup(ctx, &runner.RunnerGroup)
	if err != nil {
		return fmt.Errorf("failed to get install for runner: %w", err)
	}

	// Get the install stack with outputs to retrieve the expected AWS account ID
	installStack, err := s.getInstallStackWithOutputs(ctx, install.ID)
	if err != nil {
		return fmt.Errorf("failed to get install stack for install %s: %w", install.ID, err)
	}

	// Verify the install stack has AWS outputs with account ID
	if installStack.InstallStackOutputs.AWSStackOutputs == nil {
		return fmt.Errorf("install %s does not have AWS stack outputs configured", install.ID)
	}

	expectedAccountID := installStack.InstallStackOutputs.AWSStackOutputs.AccountID
	if expectedAccountID == "" {
		return fmt.Errorf("install %s does not have an AWS account ID in stack outputs", install.ID)
	}

	// Verify the caller's AWS account ID matches
	if callerAccountID != expectedAccountID {
		return fmt.Errorf("AWS account ID mismatch: got %s, expected %s", callerAccountID, expectedAccountID)
	}

	// Optionally verify the caller's ARN matches the expected pattern
	// The runner should be using an IAM role in the customer's account
	// For now, we just verify the account ID matches
	_ = callerARN

	return nil
}

// getInstallStackWithOutputs retrieves an install stack with its outputs preloaded
func (s *service) getInstallStackWithOutputs(ctx context.Context, installID string) (*app.InstallStack, error) {
	var installStack app.InstallStack
	res := s.db.WithContext(ctx).
		Preload("InstallStackOutputs").
		Where(app.InstallStack{
			InstallID: installID,
		}).
		First(&installStack)
	if res.Error != nil {
		return nil, res.Error
	}
	return &installStack, nil
}

// createRunnerToken creates an authentication token for the runner
func (s *service) createRunnerToken(ctx context.Context, runnerID string) (string, error) {
	email := account.ServiceAccountEmail(runnerID)

	token, err := s.acctClient.CreateToken(ctx, email, defaultRunnerTokenTimeout)
	if err != nil {
		return "", fmt.Errorf("unable to create token: %w", err)
	}

	return token.Token, nil
}

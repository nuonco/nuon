package service

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	awstypes "github.com/nuonco/nuon/pkg/types/aws"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
)

const (
	runnerIDTagKey            = "runner.nuon.co/id"
	defaultRunnerTokenTimeout = time.Hour * 24 * 90

	presignedRequestTimeout  = 10 * time.Second
	maxPresignedResponseSize = 64 * 1024
)

var (
	awsSTSHostPattern    = regexp.MustCompile(`^sts(\.([a-z]{2}-[a-z]+-\d|us-gov-[a-z]+-\d|cn-[a-z]+-\d))?\.amazonaws\.com$`)
	awsEC2HostPattern    = regexp.MustCompile(`^ec2\.([a-z]{2}-[a-z]+-\d|us-gov-[a-z]+-\d|cn-[a-z]+-\d)\.amazonaws\.com$`)
	ec2InstanceIDPattern = regexp.MustCompile(`^i-[0-9a-f]{8}([0-9a-f]{9})?$`)

	allowedPresignedHeaders = map[string]struct{}{
		"host":                 {},
		"x-amz-date":           {},
		"x-amz-security-token": {},
		"x-amz-content-sha256": {},
		"authorization":        {},
		"x-amz-algorithm":      {},
		"x-amz-credential":     {},
		"x-amz-signedheaders":  {},
		"x-amz-signature":      {},
		"x-amz-expires":        {},
	}

	presignedHTTPClient = &http.Client{
		Timeout: presignedRequestTimeout,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Transport: &http.Transport{
			TLSHandshakeTimeout:   5 * time.Second,
			ResponseHeaderTimeout: 5 * time.Second,
			DisableKeepAlives:     true,
		},
	}
)

type RunnerAuthAWSRequest struct {
	STSRequest  *awstypes.PresignedRequest `json:"sts" validate:"required"`
	TagsRequest *awstypes.PresignedRequest `json:"tags" validate:"required"`
}

type RunnerAuthAWSResponse struct {
	Authenticated bool   `json:"authenticated"`
	AccountID     string `json:"account_id,omitempty"`
	ARN           string `json:"arn,omitempty"`
	InstanceID    string `json:"instance_id,omitempty"`
	RunnerID      string `json:"runner_id,omitempty"`
	Token         string `json:"token,omitempty"`
}

type GetCallerIdentityResponse struct {
	XMLName xml.Name `xml:"GetCallerIdentityResponse"`
	Result  struct {
		Arn     string `xml:"Arn"`
		UserId  string `xml:"UserId"`
		Account string `xml:"Account"`
	} `xml:"GetCallerIdentityResult"`
}

type DescribeTagsResponse struct {
	XMLName xml.Name `xml:"DescribeTagsResponse"`
	TagSet  struct {
		Items []DescribeTagsItem `xml:"item"`
	} `xml:"tagSet"`
}

type DescribeTagsItem struct {
	Key          string `xml:"key"`
	Value        string `xml:"value"`
	ResourceId   string `xml:"resourceId"`
	ResourceType string `xml:"resourceType"`
}

func validateTagResponse(items []DescribeTagsItem) (instanceID string, tags map[string]string, err error) {
	if len(items) == 0 {
		return "", nil, errors.New("no tags returned from instance")
	}

	tags = make(map[string]string, len(items))

	for _, item := range items {
		if item.ResourceType != "instance" {
			return "", nil, fmt.Errorf("unexpected resource type %q, expected instance", item.ResourceType)
		}

		if instanceID == "" {
			instanceID = item.ResourceId
		} else if item.ResourceId != instanceID {
			return "", nil, fmt.Errorf("tags from multiple resources: %s and %s", instanceID, item.ResourceId)
		}

		tags[item.Key] = item.Value
	}

	if !ec2InstanceIDPattern.MatchString(instanceID) {
		return "", nil, fmt.Errorf("invalid instance ID format: %s", instanceID)
	}

	return instanceID, tags, nil
}

func extractInstanceIDFromSTSUserId(userId string) string {
	parts := strings.Split(userId, ":")
	if len(parts) != 2 {
		return ""
	}

	instanceID := parts[1]
	if ec2InstanceIDPattern.MatchString(instanceID) {
		return instanceID
	}

	return ""
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
	if err := ctx.ShouldBindJSON(&req); err != nil {
		s.l.Warn("runner auth: failed to parse request", zap.Error(err))
		ctx.Error(stderr.NewInvalidRequest(errors.New("invalid request format")))
		ctx.Abort()
		return
	}

	if err := s.v.Struct(req); err != nil {
		s.l.Warn("runner auth: request validation failed", zap.Error(err))
		ctx.Error(stderr.NewInvalidRequest(errors.New("invalid request: missing required fields")))
		ctx.Abort()
		return
	}

	reqCtx := ctx.Request.Context()

	stsResponse, err := s.executePresignedRequest(reqCtx, req.STSRequest, presignedRequestTypeSTS)
	if err != nil {
		s.l.Warn("runner auth: STS request failed", zap.Error(err))
		ctx.Error(stderr.ErrAuthentication{
			Err:         errors.New("authentication failed"),
			Description: "failed to verify AWS identity",
		})
		ctx.Abort()
		return
	}

	var callerIdentity GetCallerIdentityResponse
	if err := xml.Unmarshal(stsResponse, &callerIdentity); err != nil {
		s.l.Warn("runner auth: failed to parse STS response", zap.Error(err))
		ctx.Error(stderr.ErrAuthentication{
			Err:         errors.New("authentication failed"),
			Description: "invalid AWS identity response",
		})
		ctx.Abort()
		return
	}

	tagsResponse, err := s.executePresignedRequest(reqCtx, req.TagsRequest, presignedRequestTypeEC2)
	if err != nil {
		s.l.Warn("runner auth: EC2 tags request failed", zap.Error(err))
		ctx.Error(stderr.ErrAuthentication{
			Err:         errors.New("authentication failed"),
			Description: "failed to verify instance tags",
		})
		ctx.Abort()
		return
	}

	var describeTags DescribeTagsResponse
	if err := xml.Unmarshal(tagsResponse, &describeTags); err != nil {
		s.l.Warn("runner auth: failed to parse EC2 tags response", zap.Error(err))
		ctx.Error(stderr.ErrAuthentication{
			Err:         errors.New("authentication failed"),
			Description: "invalid instance tags response",
		})
		ctx.Abort()
		return
	}

	instanceID, tags, err := validateTagResponse(describeTags.TagSet.Items)
	if err != nil {
		s.l.Warn("runner auth: tag validation failed", zap.Error(err))
		ctx.Error(stderr.ErrAuthentication{
			Err:         errors.New("authentication failed"),
			Description: "invalid instance tag data",
		})
		ctx.Abort()
		return
	}

	stsInstanceID := extractInstanceIDFromSTSUserId(callerIdentity.Result.UserId)
	if stsInstanceID != "" && stsInstanceID != instanceID {
		s.l.Warn("runner auth: instance ID mismatch between STS and tags",
			zap.String("sts_instance_id", stsInstanceID),
			zap.String("tags_instance_id", instanceID))
		ctx.Error(stderr.ErrAuthentication{
			Err:         errors.New("authentication failed"),
			Description: "instance identity mismatch",
		})
		ctx.Abort()
		return
	}

	runnerID, ok := tags[runnerIDTagKey]
	if !ok || runnerID == "" {
		s.l.Warn("runner auth: missing runner ID tag",
			zap.String("instance_id", instanceID),
			zap.String("expected_tag", runnerIDTagKey))
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
			s.l.Warn("runner auth: runner not found", zap.String("runner_id", runnerID))
		} else {
			s.l.Error("runner auth: failed to get runner", zap.String("runner_id", runnerID), zap.Error(err))
		}
		ctx.Error(stderr.ErrAuthentication{
			Err:         errors.New("authentication failed"),
			Description: "runner not recognized",
		})
		ctx.Abort()
		return
	}

	if err := s.validateRunnerAWSIdentity(reqCtx, runner, callerIdentity.Result.Account, callerIdentity.Result.Arn); err != nil {
		s.l.Warn("runner auth: AWS identity validation failed",
			zap.String("runner_id", runnerID),
			zap.String("caller_account", callerIdentity.Result.Account),
			zap.String("caller_arn", callerIdentity.Result.Arn),
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
		s.l.Error("runner auth: failed to create token", zap.String("runner_id", runnerID), zap.Error(err))
		ctx.Error(stderr.ErrSystem{
			Err:         errors.New("internal error"),
			Description: "failed to issue authentication token",
		})
		ctx.Abort()
		return
	}

	s.l.Info("runner auth: authentication successful",
		zap.String("runner_id", runner.ID),
		zap.String("instance_id", instanceID),
		zap.String("account_id", callerIdentity.Result.Account))

	ctx.JSON(http.StatusOK, RunnerAuthAWSResponse{
		Authenticated: true,
		AccountID:     callerIdentity.Result.Account,
		ARN:           callerIdentity.Result.Arn,
		InstanceID:    instanceID,
		RunnerID:      runner.ID,
		Token:         token,
	})
}

type presignedRequestType int

const (
	presignedRequestTypeSTS presignedRequestType = iota
	presignedRequestTypeEC2
)

func validatePresignedRequest(presignedReq *awstypes.PresignedRequest, reqType presignedRequestType) error {
	if presignedReq.Method != http.MethodGet {
		return errors.New("only GET methods are allowed")
	}

	u, err := url.Parse(presignedReq.URL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	if u.Scheme != "https" {
		return errors.New("only HTTPS scheme is allowed")
	}

	if u.User != nil {
		return errors.New("URL must not contain userinfo")
	}

	host := strings.ToLower(u.Hostname())

	if err := validateNotIPAddress(host); err != nil {
		return err
	}

	switch reqType {
	case presignedRequestTypeSTS:
		if !awsSTSHostPattern.MatchString(host) {
			return fmt.Errorf("invalid STS host: %s", host)
		}
		if err := validateSTSAction(u.Query()); err != nil {
			return err
		}
	case presignedRequestTypeEC2:
		if !awsEC2HostPattern.MatchString(host) {
			return fmt.Errorf("invalid EC2 host: %s", host)
		}
		if err := validateEC2Action(u.Query()); err != nil {
			return err
		}
	}

	for key := range presignedReq.Headers {
		if _, ok := allowedPresignedHeaders[strings.ToLower(key)]; !ok {
			return fmt.Errorf("header not allowed: %s", key)
		}
	}

	return nil
}

func validateNotIPAddress(host string) error {
	if net.ParseIP(host) != nil {
		return errors.New("IP addresses are not allowed, only FQDNs")
	}
	return nil
}

func validateSTSAction(query url.Values) error {
	action := query.Get("Action")
	if action != "GetCallerIdentity" {
		return fmt.Errorf("only GetCallerIdentity action is allowed, got: %s", action)
	}
	return nil
}

func validateEC2Action(query url.Values) error {
	action := query.Get("Action")
	if action != "DescribeTags" {
		return fmt.Errorf("only DescribeTags action is allowed, got: %s", action)
	}
	return nil
}

func (s *service) executePresignedRequest(ctx context.Context, presignedReq *awstypes.PresignedRequest, reqType presignedRequestType) ([]byte, error) {
	if err := validatePresignedRequest(presignedReq, reqType); err != nil {
		return nil, fmt.Errorf("presigned request validation failed: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, presignedReq.Method, presignedReq.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	for key, value := range presignedReq.Headers {
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
		return nil, fmt.Errorf("AWS request failed with status %d", resp.StatusCode)
	}

	return body, nil
}

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

func (s *service) validateRunnerAWSIdentity(ctx context.Context, runner *app.Runner, callerAccountID string, callerARN string) error {
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

	awsOutputs := installStack.InstallStackOutputs.AWSStackOutputs

	expectedAccountID := awsOutputs.AccountID
	if expectedAccountID == "" {
		return fmt.Errorf("install %s does not have an AWS account ID in stack outputs", install.ID)
	}

	if callerAccountID != expectedAccountID {
		return fmt.Errorf("AWS account ID mismatch: got %s, expected %s", callerAccountID, expectedAccountID)
	}

	expectedRunnerRoleARN := awsOutputs.RunnerIAMRoleARN
	if expectedRunnerRoleARN == "" {
		return fmt.Errorf("install %s does not have a runner IAM role ARN in stack outputs", install.ID)
	}

	callerRoleName, err := extractRoleNameFromAssumedRoleARN(callerARN)
	if err != nil {
		return fmt.Errorf("failed to parse caller ARN: %w", err)
	}

	expectedRoleName := expectedRunnerRoleARN
	if strings.HasPrefix(expectedRunnerRoleARN, "arn:") {
		expectedRoleName, err = extractRoleNameFromIAMRoleARN(expectedRunnerRoleARN)
		if err != nil {
			return fmt.Errorf("failed to parse expected role ARN: %w", err)
		}
	}

	if callerRoleName != expectedRoleName {
		return fmt.Errorf("IAM role mismatch: caller role %q does not match expected role %q", callerRoleName, expectedRoleName)
	}

	return nil
}

func extractRoleNameFromAssumedRoleARN(arn string) (string, error) {
	parts := strings.Split(arn, ":")
	if len(parts) < 6 {
		return "", fmt.Errorf("invalid ARN format: %s", arn)
	}

	resource := parts[5]
	if !strings.HasPrefix(resource, "assumed-role/") {
		return "", fmt.Errorf("ARN is not an assumed-role: %s", arn)
	}

	resourceParts := strings.Split(resource, "/")
	if len(resourceParts) < 2 {
		return "", fmt.Errorf("invalid assumed-role resource format: %s", resource)
	}

	return resourceParts[1], nil
}

func extractRoleNameFromIAMRoleARN(arn string) (string, error) {
	parts := strings.Split(arn, ":")
	if len(parts) < 6 {
		return "", fmt.Errorf("invalid ARN format: %s", arn)
	}

	resource := parts[5]
	if !strings.HasPrefix(resource, "role/") {
		return "", fmt.Errorf("ARN is not an IAM role: %s", arn)
	}

	resourceParts := strings.Split(resource, "/")
	if len(resourceParts) < 2 {
		return "", fmt.Errorf("invalid role resource format: %s", resource)
	}

	return resourceParts[len(resourceParts)-1], nil
}

func (s *service) getInstallStackWithOutputs(ctx context.Context, installID string) (*app.InstallStack, error) {
	var version app.InstallStackVersion
	res := s.db.WithContext(ctx).
		Where("install_id = ?", installID).
		Where("status->>'status' = ?", app.InstallStackVersionStatusActive).
		Order("created_at DESC").
		First(&version)
	if res.Error != nil {
		return nil, fmt.Errorf("no active install stack version found: %w", res.Error)
	}

	var installStack app.InstallStack
	res = s.db.WithContext(ctx).
		Preload("InstallStackOutputs").
		First(&installStack, "id = ?", version.InstallStackID)
	if res.Error != nil {
		return nil, res.Error
	}
	return &installStack, nil
}

func (s *service) createRunnerToken(ctx context.Context, runnerID string) (string, error) {
	email := account.ServiceAccountEmail(runnerID)

	token, err := s.acctClient.CreateToken(ctx, email, defaultRunnerTokenTimeout)
	if err != nil {
		return "", fmt.Errorf("unable to create token: %w", err)
	}

	return token.Token, nil
}

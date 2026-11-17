package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"

	awstypes "github.com/nuonco/nuon/pkg/types/aws"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

const (
	boundServerIDHeader = "X-Nuon-Server-Id"
	boundRunnerIDHeader = "X-Nuon-Runner-Id"
	boundHeaderPrefix   = "x-nuon-"

	boundSTSMaxAge       = 5 * time.Minute
	boundSTSMaxClockSkew = time.Minute

	amzDateLayout = "20060102T150405Z"
)

var allowedBoundPresignedHeaders = func() map[string]struct{} {
	allowed := make(map[string]struct{}, len(allowedPresignedHeaders)+2)
	for name := range allowedPresignedHeaders {
		allowed[name] = struct{}{}
	}
	allowed[strings.ToLower(boundServerIDHeader)] = struct{}{}
	allowed[strings.ToLower(boundRunnerIDHeader)] = struct{}{}
	return allowed
}()

type boundSTSRequest struct {
	request       *awstypes.PresignedRequest
	region        string
	serverID      string
	signatureHash string
	expiresAt     time.Time
}

func parseBoundSTSRequest(presignedReq *awstypes.PresignedRequest, runnerID string, now time.Time) (*boundSTSRequest, error) {
	if err := validatePresignedRequestHeaders(presignedReq, presignedRequestTypeSTS, allowedBoundPresignedHeaders); err != nil {
		return nil, err
	}

	u, err := url.Parse(presignedReq.URL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", withoutSignedURL(err))
	}

	match := awsSTSHostPattern.FindStringSubmatch(strings.ToLower(u.Hostname()))
	if match == nil || match[2] == "" {
		return nil, errors.New("STS host must be a regional endpoint")
	}
	if u.Port() != "" {
		return nil, errors.New("STS URL must not set a port")
	}
	if u.Path != "" && u.Path != "/" {
		return nil, errors.New("STS URL must not set a path")
	}

	query, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		return nil, errors.New("invalid query string")
	}
	seen := make(map[string]struct{}, len(query))
	for key, values := range query {
		lower := strings.ToLower(key)
		if _, ok := seen[lower]; ok || len(values) != 1 {
			return nil, fmt.Errorf("query parameter %q is repeated", lower)
		}
		seen[lower] = struct{}{}
	}

	signature := query.Get("X-Amz-Signature")
	if signature == "" {
		return nil, errors.New("presigned request is missing X-Amz-Signature")
	}

	signedHeaders := strings.Split(strings.ToLower(query.Get("X-Amz-SignedHeaders")), ";")
	for _, name := range []string{boundServerIDHeader, boundRunnerIDHeader} {
		if !slices.Contains(signedHeaders, strings.ToLower(name)) {
			return nil, fmt.Errorf("%s is not a signed header", name)
		}
	}

	serverID, err := singleHeaderValue(presignedReq.Headers, boundServerIDHeader)
	if err != nil {
		return nil, err
	}

	signedAt, err := time.Parse(amzDateLayout, query.Get("X-Amz-Date"))
	if err != nil {
		return nil, errors.New("presigned request has a missing or invalid X-Amz-Date")
	}
	if signedAt.After(now.Add(boundSTSMaxClockSkew)) {
		return nil, errors.New("presigned request X-Amz-Date is in the future")
	}
	if now.Sub(signedAt) > boundSTSMaxAge {
		return nil, fmt.Errorf("presigned request is older than %s", boundSTSMaxAge)
	}
	if expiresStr := query.Get("X-Amz-Expires"); expiresStr != "" {
		expires, err := strconv.Atoi(expiresStr)
		if err != nil || expires <= 0 {
			return nil, errors.New("invalid X-Amz-Expires")
		}
		if now.After(signedAt.Add(time.Duration(expires) * time.Second)) {
			return nil, errors.New("presigned request has expired")
		}
	}

	headers := make(map[string]string, len(presignedReq.Headers)+2)
	for key, value := range presignedReq.Headers {
		if strings.HasPrefix(strings.ToLower(key), boundHeaderPrefix) {
			continue
		}
		headers[key] = value
	}
	headers[boundServerIDHeader] = serverID
	headers[boundRunnerIDHeader] = runnerID

	hash := sha256.Sum256([]byte(strings.ToLower(signature)))

	return &boundSTSRequest{
		request: &awstypes.PresignedRequest{
			Method:  presignedReq.Method,
			URL:     presignedReq.URL,
			Headers: headers,
		},
		region:        match[2],
		serverID:      strings.ToLower(serverID),
		signatureHash: hex.EncodeToString(hash[:]),
		expiresAt:     signedAt.Add(boundSTSMaxAge + boundSTSMaxClockSkew),
	}, nil
}

func singleHeaderValue(headers map[string]string, name string) (string, error) {
	var (
		value string
		found bool
	)
	for key, v := range headers {
		if !strings.EqualFold(key, name) {
			continue
		}
		if found {
			return "", fmt.Errorf("header %s is repeated", name)
		}
		value, found = strings.TrimSpace(v), true
	}
	if value == "" {
		return "", fmt.Errorf("header %s is missing", name)
	}
	return value, nil
}

func runnerAPIHost(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	return strings.ToLower(u.Host)
}

func (s *service) boundServerIDAllowed(ctx context.Context, serverID string) (bool, error) {
	if host := runnerAPIHost(s.cfg.RunnerAPIURL); host != "" && serverID == host {
		return true, nil
	}

	var count int64
	res := s.db.WithContext(ctx).
		Model(&app.RunnerGroupSettings{}).
		Where("runner_api_url IN ?", []string{"https://" + serverID, "https://" + serverID + "/"}).
		Limit(1).
		Count(&count)
	if res.Error != nil {
		return false, res.Error
	}
	return count > 0, nil
}

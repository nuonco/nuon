package nuonrunner

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
)

// appTransport is a transport that injects our authentication token and org id into the api request
type appTransport struct {
	authToken     string
	authTokenFile string
	orgID         string
	clientVersion string

	transport http.RoundTripper
	jobTokens sync.Map
}

func (t *appTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	token := t.authToken
	if t.authTokenFile != "" {
		file, err := os.Open(t.authTokenFile)
		if err != nil {
			return nil, fmt.Errorf("open auth token file: %w", err)
		}
		contents, err := io.ReadAll(io.LimitReader(file, 16*1024+1))
		file.Close()
		if err != nil {
			return nil, fmt.Errorf("read auth token file: %w", err)
		}
		token = strings.TrimSpace(string(contents))
		if token == "" || len(contents) > 16*1024 || strings.ContainsAny(token, " \t\r\n\x00") {
			return nil, fmt.Errorf("auth token file contains an invalid credential")
		}
	}
	if jobID, ok := jobWriteID(req.Method, req.URL.Path); ok {
		if override, found := t.jobTokens.Load(jobID); found {
			if jobToken, _ := override.(string); jobToken != "" {
				token = jobToken
			}
		}
	}
	req = req.Clone(req.Context())
	req.Header.Set("Authorization", "Bearer "+token)
	if t.orgID != "" {
		req.Header.Set("X-Nuon-Org-ID", t.orgID)
	}
	if t.clientVersion != "" {
		req.Header.Set("X-Nuon-Client-Version", t.clientVersion)
	}

	return t.transport.RoundTrip(req)
}

func (c *client) SetJobToken(jobID, token string) {
	c.appTransport.jobTokens.Store(jobID, token)
}

func (c *client) ClearJobToken(jobID string) {
	c.appTransport.jobTokens.Delete(jobID)
}

var jobWritePatterns = []*regexp.Regexp{
	regexp.MustCompile(`^/v1/runners/[^/]+/jobs/([^/]+)$`),
	regexp.MustCompile(`^/v1/runner-jobs/([^/]+)$`),
	regexp.MustCompile(`^/v1/runner-jobs/([^/]+)/executions$`),
	regexp.MustCompile(`^/v1/runner-jobs/([^/]+)/executions/[^/]+$`),
	regexp.MustCompile(`^/v1/runner-jobs/([^/]+)/executions/[^/]+/result$`),
	regexp.MustCompile(`^/v1/runner-jobs/([^/]+)/executions/[^/]+/outputs$`),
}

func jobWriteID(method, path string) (string, bool) {
	switch method {
	case http.MethodPatch, http.MethodPost:
	default:
		return "", false
	}
	if strings.HasSuffix(path, "/token") || strings.Contains(path, "/plan") {
		return "", false
	}
	for _, pattern := range jobWritePatterns {
		if match := pattern.FindStringSubmatch(path); len(match) == 2 {
			if method == http.MethodPost && (strings.HasSuffix(path, "/jobs/"+match[1]) || path == "/v1/runner-jobs/"+match[1]) {
				return "", false
			}
			return match[1], true
		}
	}
	return "", false
}

func (c *client) SetOrgID(orgID string) {
	c.appTransport.orgID = orgID
}

func (c *client) SetClientVersion(version string) {
	c.appTransport.clientVersion = fmt.Sprintf("runner:%s", version)
}

package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/dashboard-ui/server/internal"
)

type SandboxBuildsHandler struct {
	cfg *internal.Config
	l   *zap.Logger
}

func NewSandboxBuildsHandler(cfg *internal.Config, l *zap.Logger) *SandboxBuildsHandler {
	return &SandboxBuildsHandler{cfg: cfg, l: l}
}

func (h *SandboxBuildsHandler) RegisterRoutes(e *gin.Engine) error {
	e.GET("/api/orgs/:orgId/apps/:appId/sandbox-builds/:buildId/sse", h.StreamSandboxBuild)
	return nil
}

type sandboxBuildStatus struct {
	StatusV2 *struct {
		Status string `json:"status"`
	} `json:"status_v2"`
}

func (h *SandboxBuildsHandler) StreamSandboxBuild(c *gin.Context) {
	orgID := c.Param("orgId")
	appID := c.Param("appId")
	buildID := c.Param("buildId")

	token, ok := sseToken(c)
	if !ok {
		return
	}

	httpClient := &http.Client{}

	// The sandbox build endpoint isn't exposed by nuon-go, so fetch it raw.
	fetchBuild := func(ctx context.Context) ([]byte, error) {
		url := fmt.Sprintf("%s/v1/apps/%s/sandbox/builds/%s", h.cfg.APIUrl, appID, buildID)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("X-Nuon-Org-ID", orgID)

		resp, err := httpClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("api returned %d", resp.StatusCode)
		}

		return io.ReadAll(resp.Body)
	}

	runSSEStream(c, sseStreamConfig{
		ClientErrMsg:        "failed to fetch sandbox build",
		FinishedGracePeriod: sseFinishedGracePeriod,
		Log:                 h.l,
		Fetch: func(ctx context.Context) (sseFetchResult, error) {
			data, err := fetchBuild(ctx)
			if err != nil {
				return sseFetchResult{}, err
			}

			finished := false
			var s sandboxBuildStatus
			if json.Unmarshal(data, &s) == nil && s.StatusV2 != nil {
				finished = terminalStatuses[s.StatusV2.Status]
			}

			return sseFetchResult{
				Events:   []sseEvent{{Name: "sandbox-build", Data: data}},
				Finished: finished,
			}, nil
		},
	})
}

// Package telemetry reports one event per CLI command run to the control plane
// the CLI is logged into. It sends the command path without arguments, the CLI
// version, the outcome, and the duration. Opt out with DO_NOT_TRACK=1,
// NUON_DISABLE_TELEMETRY=true, or disable_telemetry: true in the CLI config.
// Local builds (version "development") send nothing unless
// NUON_ENABLE_TELEMETRY=true.
package telemetry

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/nuonco/nuon/bins/cli/internal/config"
	"github.com/nuonco/nuon/bins/cli/internal/services/version"
)

const sendTimeout = 2 * time.Second

type event struct {
	Command    string `json:"command"`
	CLIVersion string `json:"cli_version"`
	Success    bool   `json:"success"`
	DurationMS int64  `json:"duration_ms"`
	OS         string `json:"os"`
	Arch       string `json:"arch"`
	Agent      string `json:"agent,omitempty"`
}

// Disabled reports whether the user opted out of telemetry.
func Disabled(cfg *config.Config) bool {
	if v, err := strconv.ParseBool(os.Getenv("DO_NOT_TRACK")); err == nil && v {
		return true
	}
	return cfg != nil && cfg.GetBool("disable_telemetry")
}

// enabledForLocalBuild reports whether a local build (version
// "development") opted in with enable_telemetry (NUON_ENABLE_TELEMETRY=true).
func enabledForLocalBuild(cfg *config.Config) bool {
	return cfg.GetBool("enable_telemetry")
}

// Send reports a finished command. Local builds report only when
// NUON_ENABLE_TELEMETRY=true. It never returns an error and gives up after a
// short timeout so it cannot hold up the CLI.
func Send(cfg *config.Config, success bool, duration time.Duration) {
	if cfg == nil || cfg.APIToken == "" || cfg.AgentCommand == "" || Disabled(cfg) {
		return
	}
	if version.IsDev() && !enabledForLocalBuild(cfg) {
		return
	}

	body, err := json.Marshal(event{
		Command:    cfg.AgentCommand,
		CLIVersion: version.Version,
		Success:    success,
		DurationMS: duration.Milliseconds(),
		OS:         runtime.GOOS,
		Arch:       runtime.GOARCH,
		Agent:      cfg.Agent,
	})
	if err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), sendTimeout)
	defer cancel()

	url := strings.TrimRight(cfg.APIURL, "/") + "/v1/general/cli-events"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.APIToken)
	req.Header.Set("X-Nuon-Client-Version", version.Version)
	if cfg.OrgID != "" {
		req.Header.Set("X-Nuon-Org-ID", cfg.OrgID)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	resp.Body.Close()
}

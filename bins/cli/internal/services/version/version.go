package version

import (
	"context"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/nuonco/nuon/bins/cli/internal/ui"
)

var Version string = "development"

// remoteLookupTimeout bounds the control-plane lookup. nuon version is often
// the first command someone runs while the API URL is wrong, so a hung dial
// should not stall the client version.
const remoteLookupTimeout = 500 * time.Millisecond

type info struct {
	client      string
	ctlAPI      string
	minCLI      string
	minCLIError string
	minCLIWarn  string
	goVer       string
	os          string
	arch        string
	commit      string
	commitT     string
}

func (s *Service) Version(ctx context.Context, asJSON bool) error {
	got := s.collect(ctx)
	if asJSON {
		ui.PrintJSON(got.json())
		return nil
	}

	ui.Print(got.text())
	return nil
}

func (s *Service) collect(ctx context.Context) info {
	got := info{client: Version}
	if build, ok := debug.ReadBuildInfo(); ok {
		got.goVer = build.GoVersion
		got.os = runtime.GOOS
		got.arch = runtime.GOARCH
		for _, setting := range build.Settings {
			switch setting.Key {
			case "vcs.revision":
				got.commit = setting.Value
			case "vcs.time":
				got.commitT = setting.Value
			}
		}
	}

	if !s.apiConfigured() {
		return got
	}

	ctx, cancel := context.WithTimeout(ctx, remoteLookupTimeout)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		remote, err := s.api.GetVersion(ctx)
		if err != nil || remote == nil {
			return
		}
		got.ctlAPI = strings.TrimSpace(remote.Version)
	}()
	go func() {
		defer wg.Done()
		cfg, err := s.api.GetCLIConfig(ctx)
		if err != nil || cfg == nil {
			return
		}
		got.minCLI, got.minCLIError, got.minCLIWarn = reportedFloors(cfg)
	}()
	wg.Wait()
	return got
}

func (i info) text() string {
	var b strings.Builder
	b.WriteString("client: ")
	b.WriteString(i.client)
	b.WriteByte('\n')
	if i.ctlAPI != "" {
		b.WriteString("ctl-api: ")
		b.WriteString(i.ctlAPI)
		b.WriteByte('\n')
	}
	if i.minCLI != "" {
		b.WriteString("min_cli_supported_version: ")
		b.WriteString(i.minCLI)
		b.WriteByte('\n')
	}
	if i.minCLIError != "" {
		b.WriteString("min_cli_version_error: ")
		b.WriteString(i.minCLIError)
		b.WriteByte('\n')
	}
	if i.minCLIWarn != "" {
		b.WriteString("min_cli_version_warn: ")
		b.WriteString(i.minCLIWarn)
		b.WriteByte('\n')
	}
	return b.String()
}

func (i info) json() map[string]string {
	out := map[string]string{
		"version": i.client,
	}
	if i.ctlAPI != "" {
		out["ctl_api_version"] = i.ctlAPI
	}
	if i.minCLI != "" {
		out["min_cli_supported_version"] = i.minCLI
	}
	if i.minCLIError != "" {
		out["min_cli_version_error"] = i.minCLIError
	}
	if i.minCLIWarn != "" {
		out["min_cli_version_warn"] = i.minCLIWarn
	}
	if i.goVer != "" {
		out["go_version"] = i.goVer
	}
	if i.os != "" {
		out["os"] = i.os
	}
	if i.arch != "" {
		out["arch"] = i.arch
	}
	if i.commit != "" {
		out["commit"] = i.commit
	}
	if i.commitT != "" {
		out["commit_time"] = i.commitT
	}
	return out
}

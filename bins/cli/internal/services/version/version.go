package version

import (
	"context"
	"runtime"
	"runtime/debug"

	"github.com/nuonco/nuon/bins/cli/internal/ui"
)

var Version string = "development"

func (s *Service) Version(ctx context.Context, asJSON bool) error {
	if asJSON {
		out := map[string]string{
			"version": Version,
		}

		if info, ok := debug.ReadBuildInfo(); ok {
			out["go_version"] = info.GoVersion
			out["os"] = runtime.GOOS
			out["arch"] = runtime.GOARCH
			for _, setting := range info.Settings {
				switch setting.Key {
				case "vcs.revision":
					out["commit"] = setting.Value
				case "vcs.time":
					out["commit_time"] = setting.Value
				}
			}
		}

		ui.PrintJSON(out)
		return nil
	}

	ui.Println(Version)
	return nil
}

package version

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
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

		enc := json.NewEncoder(os.Stdout)
		return enc.Encode(out)
	}

	fmt.Println(Version)
	return nil
}

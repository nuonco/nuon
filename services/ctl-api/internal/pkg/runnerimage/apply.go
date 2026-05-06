package runnerimage

import (
	"strings"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// ApplyDigestPin returns a copy of the runner-group settings with the runner's
// pinned image digest folded into ContainerImageTag as `tag@sha256:<digest>`.
// Existing init.sh / docker pull paths concatenate URL and TAG with a colon,
// producing `URL:TAG@sha256:<digest>` which Docker resolves to the digest
// regardless of subsequent movement on the tag. When the runner has no pinned
// digest, the settings are returned unchanged.
func ApplyDigestPin(runner *app.Runner, settings app.RunnerGroupSettings) app.RunnerGroupSettings {
	digest := strings.TrimSpace(runner.ContainerImageDigest)
	if digest == "" {
		return settings
	}
	if strings.Contains(settings.ContainerImageTag, "@") {
		return settings
	}
	settings.ContainerImageTag = settings.ContainerImageTag + "@" + digest
	return settings
}

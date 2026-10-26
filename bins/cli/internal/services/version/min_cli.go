package version

import (
	"fmt"
	"regexp"

	"github.com/Masterminds/semver/v3"
)

var commitSHARegex = regexp.MustCompile(`^[0-9a-f]{7}$`)

// MinCLIWarning is the warning to show when this build is older than the control
// plane's floor. An empty or unparseable floor, a development build, and a
// 7-character commit SHA produce no warning.
func MinCLIWarning(minimum string) string {
	if IsDev() || commitSHARegex.MatchString(Version) || minimum == "" {
		return ""
	}

	minVer, err := semver.NewVersion(minimum)
	if err != nil {
		return ""
	}

	current, err := semver.NewVersion(Version)
	if err != nil || !current.LessThan(minVer) {
		return ""
	}

	return fmt.Sprintf(
		"your CLI (%s) is older than the minimum %s required by this control plane. see https://docs.nuon.co/cli to update.",
		Version, minimum,
	)
}

package internal

import (
	"strings"

	"github.com/Masterminds/semver/v3"
)

func (c *Config) MinCLIErrorFloor() string {
	if c == nil {
		return ""
	}
	v, _ := pinnedSemver(c.MinCLIVersionError)
	return v
}

func WarnAboveError(warn, errorFloor string) string {
	warn = strings.TrimSpace(warn)
	errorFloor = strings.TrimSpace(errorFloor)
	if warn == "" || errorFloor == "" {
		return warn
	}
	warnV, warnErr := semver.NewVersion(warn)
	errorV, errorErr := semver.NewVersion(errorFloor)
	if warnErr != nil || errorErr != nil {
		return warn
	}
	if warnV.LessThan(errorV) {
		return ""
	}
	return warn
}

func pinnedSemver(raw string) (string, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" || strings.EqualFold(raw, "latest") {
		return "", false
	}
	if _, err := semver.NewVersion(raw); err != nil {
		return "", false
	}
	return raw, true
}

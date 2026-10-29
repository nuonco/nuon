package version

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/Masterminds/semver/v3"

	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

var commitSHARegex = regexp.MustCompile(`^[0-9a-f]{7}$`)

// Assess compares this build to the control plane floors on cli-config.
// A policy object is authoritative. When it is absent, min_cli_version is the
// warning floor published by older control planes. An empty or unparseable
// floor, the version string "development", and a 7-character commit SHA match nothing.
// A warn floor below the error floor is ignored.
func Assess(cfg *models.ServiceCLIConfig) (warn string, block error) {
	if IsDev() || commitSHARegex.MatchString(Version) || cfg == nil {
		return "", nil
	}

	current, err := semver.NewVersion(Version)
	if err != nil {
		return "", nil
	}

	errorFloor, warnFloor := floors(cfg)
	if below(current, errorFloor) {
		return "", fmt.Errorf(
			"Your CLI (%s) is older than %s, the oldest version this control plane accepts. see https://docs.nuon.co/cli to update.",
			Version, strings.TrimSpace(errorFloor),
		)
	}
	if below(current, warnFloor) {
		return fmt.Sprintf(
			"Your CLI (%s) is older than the minimum %s required by this control plane. see https://docs.nuon.co/cli to update.",
			Version, strings.TrimSpace(warnFloor),
		), nil
	}
	return "", nil
}

func reportedFloors(cfg *models.ServiceCLIConfig) (supported, errorFloor, warnFloor string) {
	if cfg == nil {
		return "", "", ""
	}
	if cfg.MinCliVersionPolicy != nil {
		errorFloor = strings.TrimSpace(cfg.MinCliVersionPolicy.Error)
		warnFloor = strings.TrimSpace(cfg.MinCliVersionPolicy.Warn)
		supported = errorFloor
		if supported == "" {
			supported = strings.TrimSpace(cfg.MinCliVersion)
		}
		return supported, errorFloor, warnFloor
	}
	return strings.TrimSpace(cfg.MinCliVersion), "", ""
}

func floors(cfg *models.ServiceCLIConfig) (errorFloor, warnFloor string) {
	if cfg.MinCliVersionPolicy != nil {
		errorFloor = strings.TrimSpace(cfg.MinCliVersionPolicy.Error)
		warnFloor = strings.TrimSpace(cfg.MinCliVersionPolicy.Warn)
		if semverLess(warnFloor, errorFloor) {
			warnFloor = ""
		}
		return errorFloor, warnFloor
	}
	return "", strings.TrimSpace(cfg.MinCliVersion)
}

func below(current *semver.Version, minimum string) bool {
	minimum = strings.TrimSpace(minimum)
	if minimum == "" {
		return false
	}
	minVer, err := semver.NewVersion(minimum)
	if err != nil {
		return false
	}
	return current.LessThan(minVer)
}

func semverLess(left, right string) bool {
	l, errL := semver.NewVersion(strings.TrimSpace(left))
	r, errR := semver.NewVersion(strings.TrimSpace(right))
	if errL != nil || errR != nil {
		return false
	}
	return l.LessThan(r)
}

package version

import (
	"fmt"
	"regexp"

	"github.com/Masterminds/semver/v3"

	"github.com/nuonco/nuon/bins/cli/internal/ui"
)

var commitSHARegex = regexp.MustCompile(`^[0-9a-f]{7}$`)

// RequireMinCLI reports when this build is older than the control plane's floor.
// An empty or unparseable floor, a development build, and a 7-character commit
// SHA do not block the command.
func RequireMinCLI(minimum string) error {
	if IsDev() || commitSHARegex.MatchString(Version) || minimum == "" {
		return nil
	}

	minVer, err := semver.NewVersion(minimum)
	if err != nil {
		return nil
	}

	current, err := semver.NewVersion(Version)
	if err != nil || !current.LessThan(minVer) {
		return nil
	}

	return &minCLIError{
		msg: fmt.Sprintf(
			"your CLI (%s) is older than the minimum %s required by this control plane. see https://docs.nuon.co/cli to update.",
			Version, minimum,
		),
	}
}

// minCLIError is returned unprinted so fang can render the ERROR box.
// Unwrap lets agent output classify it as a user error.
type minCLIError struct {
	msg string
}

func (e *minCLIError) Error() string { return e.msg }

func (e *minCLIError) Unwrap() error { return &ui.CLIUserError{Msg: e.msg} }

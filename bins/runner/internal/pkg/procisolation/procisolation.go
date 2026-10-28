// Package procisolation keeps the runner's control-plane credentials away from
// the processes it starts. Actions, terraform, and helm run vendor code as
// children of the runner, so anything left in the runner's environment or
// readable through /proc/<runner pid> is readable by that code.
package procisolation

import (
	"os"
	"strings"

	"github.com/pkg/errors"
)

// Apply must run after the runner config has been loaded from the environment.
// Every RUNNER_* variable is dropped, including the API token, the runner id,
// and the API URL. Child processes inherit os.Environ(), and fetch-token reads
// those variables to mint a new token.
//
// Unsetting does not rewrite the original environment block /proc/<pid>/environ
// serves, so the process is also marked non-dumpable. That denies same-uid
// processes without CAP_SYS_PTRACE access to /proc/<pid>/environ and
// /proc/<pid>/mem.
func Apply() error {
	for _, envVar := range os.Environ() {
		name, _, _ := strings.Cut(envVar, "=")
		if !strings.HasPrefix(name, "RUNNER_") {
			continue
		}
		if err := os.Unsetenv(name); err != nil {
			return errors.Wrapf(err, "unable to unset %s", name)
		}
	}

	if err := setNonDumpable(); err != nil {
		return errors.Wrap(err, "unable to mark runner process non-dumpable")
	}

	return nil
}

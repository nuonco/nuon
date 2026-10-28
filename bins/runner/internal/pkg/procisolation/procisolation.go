// Package procisolation keeps the runner's control-plane credentials away from
// the processes it starts. Actions, terraform, and helm run vendor code as
// children of the runner, so anything left in the runner's environment or
// readable through /proc/<runner pid> is readable by that code.
package procisolation

import (
	"os"

	"github.com/pkg/errors"
)

// secretEnvVars are read once into the runner config and must not reach child
// processes, which inherit os.Environ().
var secretEnvVars = []string{
	"RUNNER_API_TOKEN",
}

// Apply must run after the runner config has been loaded from the environment.
// Unsetting only hides the token from os.Environ(); /proc/<pid>/environ still
// serves the original environment block, so the process is also marked
// non-dumpable, which denies same-uid processes without CAP_SYS_PTRACE access
// to its /proc/<pid>/environ and /proc/<pid>/mem.
func Apply() error {
	for _, name := range secretEnvVars {
		if err := os.Unsetenv(name); err != nil {
			return errors.Wrapf(err, "unable to unset %s", name)
		}
	}

	if err := setNonDumpable(); err != nil {
		return errors.Wrap(err, "unable to mark runner process non-dumpable")
	}

	return nil
}

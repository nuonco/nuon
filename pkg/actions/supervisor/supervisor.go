// Package supervisor is the entrypoint Nuon mounts into an image-backed
// action's container. It provisions the nuon_output helper, executes the
// rendered step script inside the image, and propagates the script's exit
// status. Outputs are left in NUON_ACTIONS_OUTPUT_FILEPATH for the runner to
// read from the shared workspace mount after the container exits.
package supervisor

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

const (
	// OutputFilepathEnvVar names the file the nuon_output helper appends to.
	OutputFilepathEnvVar = "NUON_ACTIONS_OUTPUT_FILEPATH"
	// RootEnvVar names the workspace root inside the container.
	RootEnvVar = "NUON_ACTIONS_ROOT"

	nuonOutputShim = `#!/bin/sh
printf '%s=%s\n' "$1" "$2" >> "$NUON_ACTIONS_OUTPUT_FILEPATH"
`
)

// Run provisions the nuon_output helper, ensures the script is executable, and
// executes it in workdir, streaming stdout/stderr through. It returns the
// script's exit code (0 on success) alongside any execution error.
func Run(ctx context.Context, scriptPath, workdir string) (int, error) {
	root := os.Getenv(RootEnvVar)
	if root == "" {
		root = workdir
	}

	binDir := filepath.Join(root, ".nuon", "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		return 1, errors.Wrap(err, "unable to create nuon bin directory")
	}

	shimPath := filepath.Join(binDir, "nuon_output")
	if err := os.WriteFile(shimPath, []byte(nuonOutputShim), 0o755); err != nil {
		return 1, errors.Wrap(err, "unable to write nuon_output helper")
	}

	if err := ensureExecutableScript(scriptPath); err != nil {
		return 1, err
	}

	env := envWithPathPrefix(os.Environ(), binDir)

	cmd := exec.CommandContext(ctx, scriptPath)
	cmd.Dir = workdir
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return exitErr.ExitCode(), nil
		}
		return 1, errors.Wrap(err, "unable to execute action script")
	}

	return 0, nil
}

// ensureExecutableScript adds a POSIX shebang when the script lacks one and
// makes it executable, so any base image with /bin/sh can run it.
func ensureExecutableScript(scriptPath string) error {
	contents, err := os.ReadFile(scriptPath)
	if err != nil {
		return errors.Wrap(err, "unable to read action script")
	}

	if !strings.HasPrefix(string(contents), "#!") {
		contents = append([]byte("#!/bin/sh\n"), contents...)
		if err := os.WriteFile(scriptPath, contents, 0o755); err != nil {
			return errors.Wrap(err, "unable to rewrite action script with shebang")
		}
	}

	// os.WriteFile does not change the mode of an existing file, so chmod
	// explicitly to ensure the script is executable in any base image.
	if err := os.Chmod(scriptPath, 0o755); err != nil {
		return errors.Wrap(err, "unable to make action script executable")
	}
	return nil
}

// envWithPathPrefix returns env with binDir prepended to PATH.
func envWithPathPrefix(env []string, binDir string) []string {
	out := make([]string, 0, len(env))
	found := false
	for _, kv := range env {
		if strings.HasPrefix(kv, "PATH=") {
			out = append(out, fmt.Sprintf("PATH=%s%c%s", binDir, os.PathListSeparator, strings.TrimPrefix(kv, "PATH=")))
			found = true
			continue
		}
		out = append(out, kv)
	}
	if !found {
		out = append(out, fmt.Sprintf("PATH=%s%c/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin", binDir, os.PathListSeparator))
	}
	return out
}

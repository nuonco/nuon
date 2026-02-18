package extensions

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Exec runs an installed extension with the given arguments and environment variables.
func (m *Manager) Exec(name string, args []string, env map[string]string) error {
	ext, err := m.Get(name)
	if err != nil {
		return err
	}
	if ext == nil {
		return fmt.Errorf("extension %q is not installed", name)
	}

	// Check auth requirements and warn (not hard fail)
	if ext.RequiresToken {
		if env["NUON_API_TOKEN"] == "" {
			fmt.Fprintf(os.Stderr, "Warning: extension %q requires an API token but none is configured\n", name)
		}
	}
	if ext.RequiresOrg {
		if env["NUON_ORG_ID"] == "" {
			fmt.Fprintf(os.Stderr, "Warning: extension %q requires an org to be selected but none is configured\n", name)
		}
	}

	// Resolve binary path
	binaryPath := filepath.Join(m.dir, "nuon-ext-"+name, ext.Binary)
	if _, err := os.Stat(binaryPath); err != nil {
		return fmt.Errorf("extension binary not found: %s", binaryPath)
	}

	// Build command
	cmd := exec.Command(binaryPath, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Set environment: inherit current env + add extension env vars
	cmd.Env = os.Environ()
	for k, v := range env {
		cmd.Env = append(cmd.Env, k+"="+v)
	}

	// Add extension-specific env vars
	cmd.Env = append(cmd.Env,
		"NUON_EXT_NAME="+name,
		"NUON_EXT_DIR="+filepath.Join(m.dir, "nuon-ext-"+name),
	)

	// Run the extension
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		return fmt.Errorf("unable to run extension: %w", err)
	}

	return nil
}

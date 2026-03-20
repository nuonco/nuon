package extensions

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/nuonco/nuon/bins/cli/internal/ui"
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

	// Check auth/context requirements and warn (not hard fail)
	if ext.RequiresToken && env["NUON_API_TOKEN"] == "" {
		ui.PrintWarning(fmt.Sprintf("extension %q requires an API token but none is configured", name))
	}
	if ext.RequiresOrg && env["NUON_ORG_ID"] == "" {
		ui.PrintWarning(fmt.Sprintf("extension %q requires an org to be selected but none is configured", name))
	}
	if ext.RequiresApp && env["NUON_APP_ID"] == "" {
		ui.PrintWarning(fmt.Sprintf("extension %q requires an app to be selected but none is configured", name))
	}
	if ext.RequiresInstall && env["NUON_INSTALL_ID"] == "" {
		ui.PrintWarning(fmt.Sprintf("extension %q requires an install to be selected but none is configured", name))
	}

	extDir := filepath.Join(m.dir, "nuon-ext-"+name)
	extType := ext.Type
	if extType == "" {
		extType = ExtTypeBinary
	}
	ui.PrintDebug(fmt.Sprintf("executing extension %s (type=%s)", name, extType))

	var cmd *exec.Cmd

	switch extType {
	case ExtTypePython:
		ui.PrintDebug(fmt.Sprintf("running: uv run %s %v", ext.Entrypoint, args))
		uvArgs := append([]string{"run", ext.Entrypoint}, args...)
		cmd = exec.Command("uv", uvArgs...)
		cmd.Dir = extDir

	case ExtTypeScript:
		entrypoint := ext.Entrypoint
		if entrypoint == "" {
			entrypoint = extensionBinaryName(name)
		}
		scriptPath := filepath.Join(extDir, entrypoint)
		ui.PrintDebug(fmt.Sprintf("running script: %s", scriptPath))
		if _, err := os.Stat(scriptPath); err != nil {
			return fmt.Errorf("extension script not found: %s", scriptPath)
		}
		cmd = exec.Command(scriptPath, args...)
		cmd.Dir = extDir

	default: // ExtTypeBinary
		binaryPath := filepath.Join(extDir, ext.Binary)
		ui.PrintDebug(fmt.Sprintf("running binary: %s", binaryPath))
		if _, err := os.Stat(binaryPath); err != nil {
			return fmt.Errorf("extension binary not found: %s", binaryPath)
		}
		cmd = exec.Command(binaryPath, args...)
	}

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
		"NUON_EXT_DIR="+extDir,
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

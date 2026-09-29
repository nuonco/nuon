package workspace

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	defaultVariablesFilenameTmpl string = "variables-%d.json"

	terraformrcFilename string = ".terraformrc"

	tfCLIConfigFileEnvVar string = "TF_CLI_CONFIG_FILE"
)

func (w *workspace) getEnvironment() map[string]string {
	envVars := make(map[string]string)
	for _, val := range os.Environ() {
		pieces := strings.SplitN(val, "=", 2)
		envVars[pieces[0]] = pieces[1]
	}

	return envVars
}

func (w *workspace) mergeMaps(a map[string]string, bs ...map[string]string) map[string]string {
	for _, b := range bs {
		for k, v := range b {
			a[k] = v
		}
	}

	return a
}

func (w *workspace) LoadVariables(ctx context.Context) error {
	w.envVars = w.getEnvironment()

	for _, vars := range w.Variables {
		if err := vars.Init(ctx); err != nil {
			return fmt.Errorf("unable to init variables: %w", err)
		}

		varEnvVars, err := vars.GetEnv(ctx)
		if err != nil {
			return fmt.Errorf("unable to get env variables: %w", err)
		}
		w.envVars = w.mergeMaps(w.envVars, varEnvVars)

		files, err := vars.GetFiles(ctx)
		if err != nil {
			return fmt.Errorf("unable to get file variables: %w", err)
		}
		for _, file := range files {
			if len(file.Contents) < 1 {
				continue
			}

			if err := w.writeFile(file.Filename, file.Contents, defaultFilePermissions); err != nil {
				return fmt.Errorf("unable to write file: %w", err)
			}

			w.varsPaths = append(w.varsPaths, file.Filename)
		}
	}

	if w.FilesystemMirrorPath != "" {
		rcPath, err := w.writeTerraformRC()
		if err != nil {
			return fmt.Errorf("unable to write terraformrc for filesystem mirror: %w", err)
		}
		w.envVars[tfCLIConfigFileEnvVar] = rcPath
	}

	return nil
}

func (w *workspace) resolveFilesystemMirrorPath() string {
	if filepath.IsAbs(w.FilesystemMirrorPath) {
		return w.FilesystemMirrorPath
	}
	return filepath.Join(w.root, w.FilesystemMirrorPath)
}

// why: writeTerraformRC writes a .terraformrc into the workspace root that
// configures terraform to source providers exclusively from the configured
// filesystem mirror. It returns the absolute path to the written file.
//
// The `direct { exclude = ["*/*"] }` block is the airgap guarantee: it
// instructs terraform to never reach out to the public registry as a
// fallback. If a provider is missing from the mirror, init will fail
// loudly rather than silently downloading it.
func (w *workspace) writeTerraformRC() (string, error) {
	mirrorPath := w.resolveFilesystemMirrorPath()

	contents := fmt.Sprintf(`provider_installation {
  filesystem_mirror {
    path = %q
  }
  direct {
    exclude = ["*/*"]
  }
}
`, mirrorPath)

	if err := w.writeFile(terraformrcFilename, []byte(contents), defaultFilePermissions); err != nil {
		return "", fmt.Errorf("unable to write %s: %w", terraformrcFilename, err)
	}

	return filepath.Join(w.root, terraformrcFilename), nil
}

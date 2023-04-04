package services

import (
	"fmt"
	"os/exec"
)

func getDockerPath() (string, error) {
	path, err := exec.LookPath("podman")
	if err == nil {
		return path, nil
	}

	path, err = exec.LookPath("docker")
	if err != nil {
		return "", fmt.Errorf("unable to look up docker or podman in path: %w", err)
	}

	return path, nil
}

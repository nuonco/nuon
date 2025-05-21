package local

import (
	"fmt"
	"os/exec"

	"github.com/powertoolsdev/mono/bins/runner/internal"
)

// When we are building and pushing images locally, we have to do things _slightly_ different, to ensure we interoperate
// with the docker vm correctly.
//
// When using `nctl run-local`, we need to tag and push the image as `host.containers.internal`, because the docker
// build is happening _inside_ the vm. This will push it to our local registry running on port 5001 (ultimately, on the
// host)
//
// When we copy the image from the local registry to ECR, we can always access it as localhost:5001/runner.
//
// For running builds inside kaniko, we can simply use localhost:5001 for everything.

// GetContainerRuntime returns the container runtime that is installed.
func GetContainerRuntime() string {
	if _, err := exec.LookPath("docker"); err == nil {
		return "docker"
	}
	return "podman"
}

// GetLocalhostAlias returns the localhost alias for the installed container runtime.
func GetLocalhostAlias() string {
	if runtime := GetContainerRuntime(); runtime == "docker" {
		return "host.docker.internal"
	}
	return "host.containers.internal"
}

// Return a tag that can be used to build+push from the run-local environment
func GetLocalTag(cfg *internal.Config, version string) string {
	return fmt.Sprintf("%s:%d/runner:%s", GetLocalhostAlias(), cfg.RegistryPort, version)
}

// Return a tag that can be used inside kaniko
func GetKanikoTag(cfg *internal.Config, version string) string {
	return fmt.Sprintf("localhost:%d/runner:%s", cfg.RegistryPort, version)
}

// Return the tag we _always_ use for copying
func GetCopyTag(cfg *internal.Config, version string) string {
	return fmt.Sprintf("localhost:%d/runner:%s", cfg.RegistryPort, version)
}

// Return the repo we _always_ use for copying
func GetCopyRepo(cfg *internal.Config) string {
	return fmt.Sprintf("localhost:%d/runner", cfg.RegistryPort)
}

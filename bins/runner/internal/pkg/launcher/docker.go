package launcher

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

// DockerLauncher runs action containers via the host docker CLI. It mirrors the
// mechanism mng already uses to launch the install container (shelling docker),
// rather than adding a docker Go SDK dependency.
type DockerLauncher struct {
	dockerPath string
}

var _ Launcher = (*DockerLauncher)(nil)

func NewDockerLauncher() *DockerLauncher {
	path := "docker"
	if p, err := exec.LookPath("docker"); err == nil {
		path = p
	}
	return &DockerLauncher{dockerPath: path}
}

func (d *DockerLauncher) Run(ctx context.Context, spec RunSpec) error {
	// Pull credentials, when present, are written to a throwaway DOCKER_CONFIG
	// so they never touch argv or the shared docker login state.
	var dockerConfigDir string
	if spec.PullUsername != "" || spec.PullPassword != "" {
		var err error
		dockerConfigDir, err = writeDockerConfig(spec.Image, spec.PullUsername, spec.PullPassword)
		if err != nil {
			return errors.Wrap(err, "unable to write docker auth config")
		}
		defer os.RemoveAll(dockerConfigDir)
	}

	if err := d.pull(ctx, spec, dockerConfigDir); err != nil {
		return errors.Wrap(err, "unable to pull action image")
	}

	// Best-effort remove of any stale container with the same name, plus a
	// guaranteed cleanup after the run (--rm covers the normal path).
	d.remove(context.WithoutCancel(ctx), spec.ContainerName)
	defer d.remove(context.WithoutCancel(ctx), spec.ContainerName)

	return d.run(ctx, spec, dockerConfigDir)
}

func (d *DockerLauncher) pull(ctx context.Context, spec RunSpec, dockerConfigDir string) error {
	args := []string{"pull", spec.Image}
	cmd := exec.CommandContext(ctx, d.dockerPath, args...)
	cmd.Env = dockerEnv(dockerConfigDir)
	cmd.Stdout = spec.Stderr // pull progress is noise; keep it off the action stdout
	cmd.Stderr = spec.Stderr
	return cmd.Run()
}

func (d *DockerLauncher) run(ctx context.Context, spec RunSpec, dockerConfigDir string) error {
	args := []string{
		"run", "--rm",
		"--name", spec.ContainerName,
		// Isolation: default bridge network (outbound only, so kubectl/cloud
		// APIs still work) — never --network host. No added privileges.
		"--security-opt", "no-new-privileges",
	}

	if spec.Memory != "" {
		args = append(args, "--memory", spec.Memory)
	}
	if spec.CPUs != "" {
		args = append(args, "--cpus", spec.CPUs)
	}
	if spec.PidsLimit > 0 {
		args = append(args, "--pids-limit", fmt.Sprintf("%d", spec.PidsLimit))
	}

	for _, m := range spec.Mounts {
		vol := fmt.Sprintf("%s:%s", m.HostPath, m.ContainerPath)
		if m.ReadOnly {
			vol += ":ro"
		}
		args = append(args, "--volume", vol)
	}

	// Explicit env only — never --env-file of the host environment.
	for k, v := range spec.Env {
		args = append(args, "--env", fmt.Sprintf("%s=%s", k, v))
	}
	for k, v := range spec.Labels {
		args = append(args, "--label", fmt.Sprintf("%s=%s", k, v))
	}

	if len(spec.Command) > 0 {
		args = append(args, "--entrypoint", spec.Command[0])
	}

	args = append(args, spec.Image)

	if len(spec.Command) > 1 {
		args = append(args, spec.Command[1:]...)
	}

	cmd := exec.CommandContext(ctx, d.dockerPath, args...)
	cmd.Env = dockerEnv(dockerConfigDir)
	cmd.Stdout = spec.Stdout
	cmd.Stderr = spec.Stderr
	return cmd.Run()
}

func (d *DockerLauncher) remove(ctx context.Context, name string) {
	if name == "" {
		return
	}
	cmd := exec.CommandContext(ctx, d.dockerPath, "rm", "-f", name)
	_ = cmd.Run()
}

// dockerEnv returns the host environment for the docker CLI itself (so it can
// find the daemon and its config), optionally overriding DOCKER_CONFIG to the
// throwaway auth dir. This is the docker process env, NOT the container env.
func dockerEnv(dockerConfigDir string) []string {
	env := os.Environ()
	if dockerConfigDir != "" {
		env = append(env, fmt.Sprintf("DOCKER_CONFIG=%s", dockerConfigDir))
	}
	return env
}

func writeDockerConfig(image, username, password string) (string, error) {
	registry := registryHost(image)
	auth := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))

	cfg := map[string]any{
		"auths": map[string]any{
			registry: map[string]string{"auth": auth},
		},
	}

	dir, err := os.MkdirTemp("", "nuon-action-docker-cfg-")
	if err != nil {
		return "", err
	}

	data, err := json.Marshal(cfg)
	if err != nil {
		os.RemoveAll(dir)
		return "", err
	}

	if err := os.WriteFile(filepath.Join(dir, "config.json"), data, 0o600); err != nil {
		os.RemoveAll(dir)
		return "", err
	}

	return dir, nil
}

// registryHost extracts the registry host from an image ref for the docker
// config auths key (everything before the first "/" when it looks like a host).
func registryHost(image string) string {
	parts := strings.SplitN(image, "/", 2)
	if len(parts) == 2 && (strings.ContainsAny(parts[0], ".:") || parts[0] == "localhost") {
		return parts[0]
	}
	return "https://index.docker.io/v1/"
}

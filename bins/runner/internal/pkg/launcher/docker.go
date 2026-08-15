package launcher

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"github.com/pkg/errors"
)

// DockerLauncher runs action containers via the host docker CLI. It mirrors the
// mechanism mng already uses to launch the install container (shelling docker),
// rather than adding a docker Go SDK dependency.
type DockerLauncher struct {
	dockerPath string
	cache      *ImageCache
}

var _ Launcher = (*DockerLauncher)(nil)

func NewDockerLauncher(cache *ImageCache) *DockerLauncher {
	path := "docker"
	if p, err := exec.LookPath("docker"); err == nil {
		path = p
	}
	return &DockerLauncher{dockerPath: path, cache: cache}
}

// Prepare pulls the image once for the whole job and leases it, so every step
// reuses that pull and collection can't remove the image between two steps.
func (d *DockerLauncher) Prepare(ctx context.Context, spec PrepareSpec) error {
	unlock, err := d.cache.Lock(ctx)
	if err != nil {
		return errors.Wrap(err, "unable to lock action image cache")
	}
	defer unlock()

	// Lease before pulling so the image is protected for the whole window, not
	// just from the moment the pull finishes.
	if err := d.cache.Lease(spec.LeaseID, spec.Image); err != nil {
		return errors.Wrap(err, "unable to lease action image")
	}

	// Pull credentials, when present, are written to a throwaway DOCKER_CONFIG
	// so they never touch argv or the shared docker login state.
	var dockerConfigDir string
	if spec.PullUsername != "" || spec.PullPassword != "" {
		dockerConfigDir, err = writeDockerConfig(spec.Image, spec.PullUsername, spec.PullPassword)
		if err != nil {
			return errors.Wrap(err, "unable to write docker auth config")
		}
		defer os.RemoveAll(dockerConfigDir)
	}

	if err := d.pull(ctx, spec, dockerConfigDir); err != nil {
		return errors.Wrap(err, "unable to pull action image")
	}

	return d.cache.Record(spec.Image)
}

func (d *DockerLauncher) Release(leaseID string) {
	d.cache.Unlease(leaseID)
}

func (d *DockerLauncher) Run(ctx context.Context, spec RunSpec) error {
	// Prepare already pulled and leased the image, and nothing here removes it:
	// the ref is content-addressed and shared by every run of the same image, so
	// image lifetime belongs to the host-level cache, not to a single step.
	// --rm covers the normal container teardown; the defer covers the rest.
	defer d.remove(context.WithoutCancel(ctx), spec.ContainerName)

	return d.run(ctx, spec)
}

func (d *DockerLauncher) pull(ctx context.Context, spec PrepareSpec, dockerConfigDir string) error {
	args := []string{"pull", spec.Image}
	cmd := exec.CommandContext(ctx, d.dockerPath, args...)
	cmd.Env = dockerEnv(dockerConfigDir, nil)

	// Pull progress goes to the dedicated pull log (INFO), not the action's
	// stdout/stderr — docker writes it to stderr but it isn't an error.
	cmd.Stdout = spec.PullLog
	cmd.Stderr = spec.PullLog
	return cmd.Run()
}

func (d *DockerLauncher) run(ctx context.Context, spec RunSpec) error {
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
	if spec.CPUShares > 0 {
		args = append(args, "--cpu-shares", fmt.Sprintf("%d", spec.CPUShares))
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

	// Valueless "--env KEY" flags with the values supplied through the docker
	// CLI's own environment: keeps secrets out of argv (ps, /proc/pid/cmdline)
	// like an --env-file did, without silently dropping multiline values such as
	// PEM keys that --env-file cannot express.
	envArgs, envVars, err := containerEnvArgs(spec.Env)
	if err != nil {
		return errors.Wrap(err, "unable to build container environment")
	}
	args = append(args, envArgs...)
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
	// No DOCKER_CONFIG: the image is already local, so running it needs no
	// registry auth and the credentials never reach this step.
	cmd.Env = dockerEnv("", envVars)
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

// dockerEnv returns the environment for the docker CLI process: the host env so
// it can find the daemon, then the container's env vars (which docker forwards
// for each valueless "--env KEY"), then DOCKER_CONFIG last. os/exec keeps the
// last duplicate of a name, so container values beat the host env while the
// CLI's own auth settings still beat the container.
func dockerEnv(dockerConfigDir string, containerEnv []string) []string {
	env := os.Environ()
	env = append(env, containerEnv...)
	if dockerConfigDir != "" {
		env = append(env, fmt.Sprintf("DOCKER_CONFIG=%s", dockerConfigDir))
	}
	return env
}

// containerEnvArgs converts the container env into valueless "--env KEY" docker
// flags plus the matching KEY=VALUE entries for the docker CLI's environment.
// Names are validated rather than silently corrupting the invocation, and
// values may contain newlines. Sorted so the invocation is deterministic.
func containerEnvArgs(env map[string]string) (args []string, vars []string, err error) {
	names := make([]string, 0, len(env))
	for name := range env {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		if err := validateEnvName(name); err != nil {
			return nil, nil, err
		}
		value := env[name]
		if strings.ContainsRune(value, 0) {
			return nil, nil, errors.Errorf("value for environment variable %q contains a NUL byte", name)
		}

		args = append(args, "--env", name)
		vars = append(vars, name+"="+value)
	}

	return args, vars, nil
}

// validateEnvName rejects names that can't survive the trip through the docker
// CLI's environment, and DOCKER_* names that would let action config redirect
// the CLI itself (its daemon, auth, or API version).
func validateEnvName(name string) error {
	if name == "" {
		return errors.New("environment variable name is empty")
	}
	for _, r := range name {
		if r == '=' || r == 0 || unicode.IsSpace(r) {
			return errors.Errorf("environment variable name %q contains an unsupported character", name)
		}
	}
	if strings.HasPrefix(name, "DOCKER_") {
		return errors.Errorf("environment variable name %q is reserved for the runner's container runtime", name)
	}
	return nil
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

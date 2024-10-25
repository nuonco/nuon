package docker

import (
	"context"
	"os"
	"os/exec"

	"github.com/hashicorp/go-hclog"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/registry/local"
)

const (
	defaultKanikoLocation string = "/kaniko/executor"
)

func (b *handler) kanikoPath() (string, error) {
	_, err := os.Stat(defaultKanikoLocation)
	if err == nil {
		return defaultKanikoLocation, nil
	}

	path, err := exec.LookPath("executor")
	if err != nil {
		return "", errors.Wrap(err, "unable to find kaniko executor")
	}

	return path, nil
}

func (b *handler) buildWithKaniko(
	ctx context.Context,
	log hclog.Logger,
	dockerfilePath string,
	contextDir string,
	buildArgs map[string]*string,
) (string, error) {
	log.Info("Building Docker image with kaniko...")
	localRef := local.GetKanikoTag(b.state.resultTag)

	kanikoPath, err := b.kanikoPath()
	if err != nil {
		localRef = local.GetLocalTag(b.state.resultTag)
		log.Info("building locally")
		return localRef, b.buildLocal(
			ctx,
			log,
			dockerfilePath,
			contextDir,
			buildArgs,
			localRef,
		)
	}

	// Start constructing our arg string for img
	args := []string{
		kanikoPath,
		"--context", "dir://" + contextDir,
		"-f", dockerfilePath,
		"--destination", localRef,
	}

	if b.state.cfg.Target != "" {
		args = append(args, "--target", b.state.cfg.Target)
	}

	// If we have build args we append each
	for k, v := range buildArgs {
		// v should always not be nil but guard just in case to avoid a panic
		if v != nil {
			args = append(args, "--build-arg", k+"="+*v)
		}
	}

	log.Debug("executing kaniko", "args", args)
	log.Info("Executing kaniko...")

	// Command output should go to the step
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = cmd.Stdout

	if err := cmd.Run(); err != nil {
		return "", err
	}

	return "", nil
}

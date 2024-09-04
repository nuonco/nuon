package docker

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/hashicorp/go-hclog"
)

func (b *handler) buildWithKaniko(
	ctx context.Context,
	log hclog.Logger,
	dockerfilePath string,
	contextDir string,
	buildArgs map[string]*string,
) error {
	log.Info("Building Docker image with kaniko...")

	localRef := fmt.Sprintf("localhost:5000/runner:%s", b.state.resultTag)

	// Start constructing our arg string for img
	args := []string{
		"/kaniko/executor",
		"--context", "dir://" + contextDir,
		"-f", dockerfilePath,
		"-d", localRef,
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
		return err
	}

	log.Info("Image pushed to '%s:%s'", b.state.resultTag)

	return nil
}

package docker

import (
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/cli/cli/command/image/build"
	"github.com/hashicorp/go-hclog"
	"github.com/oklog/ulid/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/workspace"
)

func (b *handler) getBuildContext(
	src *workspace.Source,
	log hclog.Logger,
) (string, string, error) {
	dockerfile := b.state.cfg.Dockerfile
	if dockerfile == "" {
		dockerfile = "Dockerfile"
	}
	if !filepath.IsAbs(dockerfile) {
		dockerfile = filepath.Join(src.AbsPath(), dockerfile)
	}

	// If the dockerfile is outside of our build context, then we copy it
	// into our build context.
	relDockerfile, err := filepath.Rel(src.AbsPath(), dockerfile)
	if err != nil || strings.HasPrefix(relDockerfile, "..") {
		id, err := ulid.New(ulid.Now(), rand.Reader)
		if err != nil {
			return "", "", err
		}

		newPath := filepath.Join(src.AbsPath(), fmt.Sprintf("Dockerfile-%s", id.String()))
		if err := copyFile(dockerfile, newPath); err != nil {
			return "", "", err
		}
		defer os.Remove(newPath)

		dockerfile = newPath
	}

	path := src.AbsPath()

	if b.state.cfg.Context != "" {
		path = b.state.cfg.Context
	}

	contextDir, relDockerfile, err := build.GetContextFromLocalDir(path, dockerfile)
	if err != nil {
		return "", "", status.Errorf(codes.FailedPrecondition, "unable to create Docker context: %s", err)
	}
	log.Debug("loaded Docker context",
		"context_dir", contextDir,
		"dockerfile", relDockerfile,
	)

	log.Info("executing build via kaniko")

	return relDockerfile, contextDir, nil
}

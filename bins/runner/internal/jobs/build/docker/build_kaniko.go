package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/registry/local"
	"github.com/powertoolsdev/mono/pkg/command"
	"github.com/powertoolsdev/mono/pkg/zapwriter"
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

type kanikoBuildLine struct {
	Msg   string `json:"msg"`
	Level string `json:"level"`
	Time  string `json:"time"`
}

func (b *handler) buildWithKaniko(
	ctx context.Context,
	l *zap.Logger,
	dockerfilePath string,
	contextDir string,
	buildArgs map[string]*string,
) (string, error) {
	l.Info("Building Docker image with kaniko...")
	localRef := local.GetKanikoTag(b.cfg, b.state.resultTag)

	lf := zapwriter.NewWithOpts(l,
		zapwriter.WithLogLevel(zapcore.InfoLevel),
		zapwriter.WithLineFormatter(func(str string) string {
			var line kanikoBuildLine
			if err := json.Unmarshal([]byte(str), &line); err != nil {
				return str
			}
			return line.Msg
		}),
		zapwriter.WithLineLeveler(func(str string) zapcore.Level {
			var line kanikoBuildLine
			if err := json.Unmarshal([]byte(str), &line); err != nil {
				return zapcore.ErrorLevel
			}

			if line.Level == "info" {
				return zapcore.InfoLevel
			}
			if line.Level == "err" {
				return zapcore.ErrorLevel
			}

			return zapcore.InfoLevel
		}),
	)

	kanikoPath, err := b.kanikoPath()
	if err != nil {
		localRef = local.GetLocalTag(b.cfg, b.state.resultTag)
		l.Info("building locally")
		return localRef, b.buildLocal(
			ctx,
			l,
			dockerfilePath,
			contextDir,
			buildArgs,
			localRef,
		)
	}

	// Start constructing our arg string for img
	l.Error("context-dir is set to "+contextDir+" assuming that + "+dockerfilePath+" is either within this directory or resolves relative to this.", zap.String("dir", contextDir))
	args := []string{
		"--context", "dir://.",
		"-f", dockerfilePath,
		"--log-format", "json",
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

	l.Debug("executing kaniko",
		zap.Any("args", args),
		zap.String("cwd", contextDir),
	)
	cmd, err := command.New(b.v,
		command.WithCmd(kanikoPath),
		command.WithCwd(contextDir),
		command.WithArgs(args),
		command.WithEnv(map[string]string{}),
		command.WithStdout(lf),
		command.WithStderr(lf),
	)
	if err != nil {
		return "", fmt.Errorf("unable to create build command: %w", err)
	}
	if err := cmd.Exec(ctx); err != nil {
		return "", fmt.Errorf("unable to build: %w", err)
	}

	return localRef, nil
}
